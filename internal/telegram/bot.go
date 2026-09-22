package telegram

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/example/jobparser/internal/models"
	"github.com/example/jobparser/internal/search"
	"github.com/example/jobparser/internal/storage"
)

type Bot struct {
	api    *tgbotapi.BotAPI
	db     *storage.DB
	search *search.SearchEngine
}

func New(token string, db *storage.DB, se *search.SearchEngine) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	return &Bot{api: api, db: db, search: se}, nil
}

// NotifyNewVacancy — рассылает вакансию всем подпискам, чьим ключевым словам она соответствует.
// Вызывается менеджером парсеров через Manager.OnNewVacancy
func (b *Bot) NotifyNewVacancy(v models.Vacancy) {
	subs, err := b.db.ListSubscriptions()
	if err != nil {
		log.Printf("[bot] не удалось получить подписки: %v", err)
		return
	}

	for _, sub := range subs {
		if !matches(sub, v) {
			continue
		}
		msg := tgbotapi.NewMessage(sub.ChatID, formatVacancy(v))
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.DisableWebPagePreview = false
		if _, err := b.api.Send(msg); err != nil {
			log.Printf("[bot] ошибка отправки chat_id=%d: %v", sub.ChatID, err)
		}
	}
}

func matches(sub models.TelegramSubscription, v models.Vacancy) bool {
	if sub.Query != "" && !strings.Contains(strings.ToLower(v.Title), strings.ToLower(sub.Query)) {
		return false
	}
	if sub.City != "" && !strings.Contains(strings.ToLower(v.City), strings.ToLower(sub.City)) {
		return false
	}
	if len(sub.Sources) > 0 {
		found := false
		for _, s := range sub.Sources {
			if s == v.Source {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func formatVacancy(v models.Vacancy) string {
	salary := "не указана"
	if v.SalaryFrom != nil || v.SalaryTo != nil {
		from, to := 0, 0
		if v.SalaryFrom != nil {
			from = *v.SalaryFrom
		}
		if v.SalaryTo != nil {
			to = *v.SalaryTo
		}
		salary = fmt.Sprintf("%d - %d %s", from, to, v.Currency)
	}
	return fmt.Sprintf("*%s*\n%s | %s\n💰 %s\n🔗 %s", v.Title, v.Company, v.City, salary, v.URL)
}

// Start запускает long-polling и обработку команд
func (b *Bot) Start() {
	log.Printf("Telegram bot запущен: @%s", b.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		b.handleMessage(update.Message)
	}
}

func (b *Bot) handleMessage(msg *tgbotapi.Message) {
	text := strings.TrimSpace(msg.Text)
	chatID := msg.Chat.ID

	switch {
	case text == "/start":
		b.reply(chatID, "Привет! Я ищу вакансии на hh, habr career и qyzmet.enbek.kz.\n\n"+
			"/search <запрос> — разовый поиск\n"+
			"/subscribe <запрос> — подписаться на новые вакансии по ключевому слову\n"+
			"/unsubscribe — отписаться от всех подписок")

	case strings.HasPrefix(text, "/search"):
		query := strings.TrimSpace(strings.TrimPrefix(text, "/search"))
		if query == "" {
			b.reply(chatID, "Укажите запрос, например: /search Golang разработчик")
			return
		}
		b.doSearch(chatID, query)

	case strings.HasPrefix(text, "/subscribe"):
		query := strings.TrimSpace(strings.TrimPrefix(text, "/subscribe"))
		if query == "" {
			b.reply(chatID, "Укажите запрос, например: /subscribe Golang")
			return
		}
		sub := &models.TelegramSubscription{ChatID: chatID, Query: query}
		if err := b.db.AddSubscription(sub); err != nil {
			b.reply(chatID, "Не удалось сохранить подписку: "+err.Error())
			return
		}
		b.reply(chatID, fmt.Sprintf("Подписка оформлена на запрос: %q. Буду присылать новые вакансии.", query))

	case text == "/unsubscribe":
		if err := b.db.RemoveSubscriptions(chatID); err != nil {
			b.reply(chatID, "Ошибка отписки: "+err.Error())
			return
		}
		b.reply(chatID, "Все подписки удалены.")

	default:
		b.reply(chatID, "Не понял команду. Используйте /search, /subscribe или /unsubscribe.")
	}
}

func (b *Bot) doSearch(chatID int64, query string) {
	res, err := b.search.Search(search.SearchParams{Query: query, Limit: 10})
	if err != nil {
		b.reply(chatID, "Ошибка поиска: "+err.Error())
		return
	}
	if len(res.Hits) == 0 {
		b.reply(chatID, "Ничего не найдено по запросу: "+query)
		return
	}
	for _, hit := range res.Hits {
		data, _ := hit.(map[string]interface{})
		title, _ := data["title"].(string)
		company, _ := data["company"].(string)
		city, _ := data["city"].(string)
		url, _ := data["url"].(string)
		text := fmt.Sprintf("*%s*\n%s | %s\n🔗 %s", title, company, city, url)
		m := tgbotapi.NewMessage(chatID, text)
		m.ParseMode = tgbotapi.ModeMarkdown
		b.api.Send(m)
	}
}

func (b *Bot) reply(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	b.api.Send(msg)
}

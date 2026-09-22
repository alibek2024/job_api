CREATE TABLE IF NOT EXISTS vacancies (
    id BIGSERIAL PRIMARY KEY,
    external_id TEXT NOT NULL,
    source TEXT NOT NULL,
    title TEXT NOT NULL,
    company TEXT,
    city TEXT,
    salary_from INTEGER,
    salary_to INTEGER,
    currency TEXT,
    url TEXT UNIQUE NOT NULL,
    description TEXT,
    skills TEXT[],
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE (external_id, source)
);

CREATE INDEX IF NOT EXISTS idx_vacancies_title ON vacancies (title);
CREATE INDEX IF NOT EXISTS idx_vacancies_city ON vacancies (city);
CREATE INDEX IF NOT EXISTS idx_vacancies_source ON vacancies (source);
CREATE INDEX IF NOT EXISTS idx_vacancies_published_at ON vacancies (published_at DESC);

CREATE TABLE IF NOT EXISTS telegram_subscriptions (
    id BIGSERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL,
    query TEXT,
    city TEXT,
    sources TEXT[],
    created_at TIMESTAMPTZ DEFAULT now(),
    last_notify TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_chat_id ON telegram_subscriptions (chat_id);

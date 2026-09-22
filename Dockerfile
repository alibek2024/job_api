FROM golang:1.22-alpine AS build
ARG CMD_PATH
WORKDIR /app
COPY go.mod ./
RUN go mod download || true
COPY . .
RUN go build -o /app/bin/app ${CMD_PATH}

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /app/bin/app /app/app
ENTRYPOINT ["/app/app"]

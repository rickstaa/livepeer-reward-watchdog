FROM golang:1.25.1-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o reward-watchdog main.go

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/reward-watchdog .
COPY ABI ./ABI
# Pass TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID at runtime for better security
CMD ["/app/reward-watchdog"]

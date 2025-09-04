# Livepeer Reward Watchdog

This Go script monitors the Livepeer protocol on Arbitrum and alerts you via Telegram if your orchestrator's reward has not been called in a round after a configurable delay. It's an additional safety net alongside the [web3-livepeer-bot](https://github.com/0xVires/web3-livepeer-bot) by @0xVires.

## Features

- Monitors blockchain rounds and reward calls in real-time using Ethereum event subscriptions.
- Sends alerts to Telegram when:
  - A reward is called for your orchestrator in a round.
  - A reward has not been called after a configurable delay (e.g., 2 hours).
  - Repeats warnings every configurable interval (e.g., every N hours) until the reward is called.
- Both the delay and repeat interval for alerts are fully configurable via command-line flags.

## Requirements

- [Go 1.21+](https://go.dev/)
- A working Ethereum WebSocket RPC endpoint (e.g., `wss://arb1.arbitrum.io/ws`).
- Telegram bot token and chat ID (required for Telegram alerts).
- Discord webhook URL (required for Discord alerts).

## Alert Setup Instructions

### Telegram Bot Setup

1. Open Telegram and search for [@BotFather](https://t.me/BotFather).
2. Start a chat and send `/newbot` to create a new bot. Follow the instructions to get your bot token.
3. Start a chat with your new bot (search for its username and click "Start").
4. To get your chat ID:

   - Send a message to your bot.
   - Visit: `https://api.telegram.org/bot<your_bot_token>/getUpdates` in your browser (replace `<your_bot_token>`).
   - Look for `"chat":{"id":...}` in the response; that's your chat ID.
   - For group chats, add the bot to the group, send a message, and use the same method.

5. Set `TELEGRAM_BOT_TOKEN` and `TELEGRAM_CHAT_ID` as environment variables.

More info: [Telegram Bot API docs](https://core.telegram.org/bots#botfather)

### Discord Webhook Setup

1. Go to your Discord server and open the channel you want alerts in.
2. Click the gear icon (Edit Channel) > Integrations > Webhooks > New Webhook.
3. Name your webhook and copy the webhook URL.
4. Set `DISCORD_WEBHOOK_URL` as an environment variable.

More info: [Discord Webhooks Guide](https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks)

## Usage

### Local Setup

Run the script directly on your machine:

```bash
export TELEGRAM_BOT_TOKEN=your_bot_token
export TELEGRAM_CHAT_ID=your_chat_id
export DISCORD_WEBHOOK_URL=your_webhook_url

go run main.go --delay=2h --check-interval=1h <orchestrator-address> wss://arb1.arbitrum.io/ws
```

- `--delay` sets how long to wait after a new round before sending the first warning (default: 2h).
- `--check-interval` sets how often to check and repeat the warning if the reward is not called (default: 1h).
- `--only-reward-errors` if set, only sends alerts for reward call errors (ignores success alerts).
- `--max-retry-time` sets the maximum time to retry RPC connections before giving up (default: 30m, 0 = retry forever).

### Docker & Docker Compose

Docker and Docker Compose setups are provided for convenience. See:

- [`Dockerfile`](./Dockerfile)
- [`docker-compose.yml`](./docker-compose.yml)

Experienced users can use these files for containerized/server deployment.

## How it works

- Subscribes to [`NewRound`](https://arbiscan.io/address/0xdd6f56DcC28D3F5f27084381fE8Df634985cc39f#code) and [`Reward`](https://arbiscan.io/address/0x35Bcf3c30594191d53231E4FF333E8A770453e40#code) events from the Livepeer protocol contracts on Arbitrum.
- On new round: resets state and starts the timer.
- On reward call: sends success alert and stops further warnings for that round.
- If the delay passes and no reward is called: sends warning and repeats every interval until reward is called.

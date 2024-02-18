# narrator-bot
Telegram bot for voice message narration.

## Set up local server
1. Get server url
```sh
ssh -R 80:localhost:8080 serveo.net
```
Response example:
```sh
Forwarding HTTP traffic from https://test.serveo.net
```
2. Set SERVER_URL env variable:
```sh
export SERVER_URL=https://test.serveo.net
```

# Run application

```sh
go run cmd/app/main.go
```

# Build docker image

```sh
sh scripts/make_image.sh
```

# Run in docker compose
```sh
cd deployments
docker compose up
```
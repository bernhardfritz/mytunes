# mytunes

## Configuration

```yaml
# traefik/config/traefik.yaml
      email: "changeme@gmail.com"
```

```yaml
# traefik/.env
DUCKDNS_TOKEN = 'changeme'
PROVIDERS_DISCORD_CLIENT_ID = 'changeme' # https://discord.com/developers/applications
PROVIDERS_DISCORD_CLIENT_SECRET = 'changeme'
SECRET = 'changeme' # openssl rand -hex 16
WHITELIST = 'changeme'
```

```yaml
# docker-compose.yaml
      - traefik.http.routers.mytunes-https.rule=Host(`mytunes.changeme.duckdns.org`)
```
Go to https://discord.com/developers/applications

Create a new application for mytunes

Go to OAuth2 settings

Set `https://mytunes.changeme.duckdns.org/_oauth` as redirect URL

## Usage

```bash
docker network create proxy
docker compose up -d
```

## Development

```bash
docker compose -f docker-compose.dev.yaml up -d
```

```bash
go test ./...
```

```bash
go run .
```

## FAQ

### Why choose Discord over other OAuth2 providers?

Discord allows to register redirect URLs with custom URL schemes like `vlc://`.
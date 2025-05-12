# mytunes

## Configuration

```yaml
# traefik/config/traefik.yaml
      email: "changeme@gmail.com"
```

```bash
# traefik/.env
DUCKDNS_TOKEN = 'changeme'
PROVIDERS_OIDC_CLIENT_ID = 'changeme' # https://console.developers.google.com/auth/clients
PROVIDERS_OIDC_CLIENT_SECRET = 'changeme'
SECRET = 'changeme' # openssl rand -hex 16
WHITELIST = 'changeme'
```

```yaml
# docker-compose.yaml
      - traefik.http.routers.mytunes-https.rule=Host(`mytunes.changeme.duckdns.org`)
```

```yaml
# traefik/docker-compose.yaml
      - --rule.mytunes.rule=Host(`mytunes.changeme.duckdns.org`)&&Path(`/_vlc`)
```

Go to https://console.developers.google.com/auth/clients (or any other [OIDC](https://openid.net/developers/how-connect-works/) provider of your choice)

Create a new OAuth 2.0 Client for mytunes

Select Application type "Web application"

Add `https://mytunes.changeme.duckdns.org/_oauth` to Authorised redirect URIs

## Usage

```bash
docker network create proxy
docker compose up -d
```

Open in browser: https://mytunes.changeme.duckdns.org/

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

Open in VLC: http://localhost:8080/index.m3u

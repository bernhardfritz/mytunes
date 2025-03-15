# mytunes

## Configuration

```yaml
# traefik/config/traefik.yaml
      email: "changeme@gmail.com"
```

```yaml
# traefik/.env
DUCKDNS_TOKEN = 'changeme'
```

```yaml
# docker-compose.yaml
      - traefik.http.routers.mytunes-https.rule=Host(`mytunes.changeme.duckdns.org`)
```

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

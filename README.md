# mytunes

## Usage

```bash
docker build -t mytunes .
docker run -it --rm --name mytunes -v ~/Music/mytunes:/var/lib/mytunes --env-file .env mytunes
```

## Development

```bash
docker build -t mytunes --target build .
docker run -it --rm --name mytunes -v ~/Music/mytunes:/var/lib/mytunes --env-file .env -v "$PWD":/usr/src/mytunes -v "$PWD"/.vscode-server:/root/.vscode-server -v "$PWD"/.vscode/godev:/usr/src/mytunes/.vscode/godev mytunes
apt update
apt install ffmpeg
```

```bash
go test ./...
```

```bash
go run .
```

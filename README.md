# mytunes

## Usage

```bash
docker build -t mytunes .
docker run -it --rm --name mytunes -v ~/Music/mytunes:/var/lib/mytunes mytunes
```

## Development

```bash
docker build -t mytunes --target build .
docker run -it --rm --name mytunes -v ~/Music/mytunes:/var/lib/mytunes -v "$PWD":/usr/src/mytunes mytunes
apt update
apt install ffmpeg
```

```bash
go test ./...
```

```bash
go run .
```

# gotunes

## Usage

```bash
docker build -t gotunes .
docker run -it --rm --name gotunes -v ~/Music/gotunes:/var/lib/gotunes gotunes
```

## Development

```bash
docker build -t gotunes --target build .
docker run -it --rm --name gotunes -v ~/Music/gotunes:/var/lib/gotunes -v "$PWD":/usr/src/gotunes gotunes
apt update
apt install ffmpeg
```

```bash
go test ./...
```

```bash
go run .
```

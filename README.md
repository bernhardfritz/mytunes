# gotunes

## Usage

```bash
docker build -t gotunes .
docker run -it --rm --name gotunes gotunes
```

## Development

```bash
docker build -t gotunes --target build .
docker run -it --rm --name gotunes -v "$PWD":/usr/src/gotunes gotunes
```

```bash
go test ./...
```

```bash
go run .
```

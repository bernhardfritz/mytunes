FROM golang:1.23.2-alpine AS build

WORKDIR /usr/src/mytunes

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin ./...

FROM build AS dev
RUN apk add --no-cache ffmpeg && mkdir -p /var/lib/mytunes

FROM alpine:3.22.0

RUN apk add --no-cache ffmpeg && mkdir -p /var/lib/mytunes
COPY --from=build /usr/local/bin/mytunes /usr/local/bin/

CMD ["mytunes", "/var/lib/mytunes"]
EXPOSE 8080
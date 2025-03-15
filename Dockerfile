FROM golang:1.23.2 AS build

WORKDIR /usr/src/mytunes

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin ./...

FROM build AS dev
RUN apt-get -y update && apt-get install -y --no-install-recommends ffmpeg && mkdir -p /var/lib/mytunes && ln -s /usr/src/mytunes/index.m3u /var/lib/mytunes/index.m3u

FROM debian:bookworm-slim

RUN apt-get -y update && apt-get install -y --no-install-recommends ffmpeg
COPY --from=build /usr/local/bin/mytunes /usr/local/bin/
COPY index.m3u /var/lib/mytunes/

CMD ["mytunes"]
EXPOSE 8080
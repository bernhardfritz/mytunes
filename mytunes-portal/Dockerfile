FROM golang:1.23.2-alpine AS build

WORKDIR /usr/src/mytunes-portal

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN apk add --update gcc musl-dev && CGO_ENABLED=1 go build -v -o /usr/local/bin ./...

FROM alpine:3.22.0

COPY --from=build /usr/local/bin/mytunes-portal /usr/local/bin/

CMD ["mytunes-portal", "/var/lib/mytunes-portal"]
EXPOSE 8080
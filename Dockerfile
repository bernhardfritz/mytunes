FROM golang:1.23.2 AS build

WORKDIR /usr/src/gotunes

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin ./...

FROM scratch

COPY --from=build /usr/local/bin/gotunes /usr/local/bin/gotunes

CMD ["gotunes"]

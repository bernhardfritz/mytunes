FROM --platform=$BUILDPLATFORM golang:1.23.2-alpine AS build
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

WORKDIR /usr/src/mytunes-portal

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN GOOS=${TARGETOS}; \
    GOARCH=${TARGETARCH}; \
    if [ "${TARGETARCH}" = "arm" ] && [ "${TARGETVARIANT}" ]; then \
        GOARM="${TARGETVARIANT#v}"; \
    fi; \
    go build -v -o /usr/local/bin ./...

FROM scratch

COPY --from=build /usr/local/bin/mytunes-portal /usr/local/bin/

CMD ["mytunes-portal"]
EXPOSE 8080
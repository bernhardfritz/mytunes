FROM --platform=$BUILDPLATFORM golang:1.23.2-alpine AS build
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

WORKDIR /tmp
RUN if [ "${TARGETARCH}" = "arm" ]; then \
      if [ "${TARGETVARIANT}" = "v5" ] || [ "${TARGETVARIANT}" = "v6" ]; then \
        FFMPEGARCH="armel"; \
      else \
        FFMPEGARCH="armhf"; \
      fi; \
    else \
      FFMPEGARCH="${TARGETARCH}"; \
    fi; \
    wget "https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-${FFMPEGARCH}-static.tar.xz" \
    && wget "https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-${FFMPEGARCH}-static.tar.xz.md5" \
    && md5sum -c "ffmpeg-release-${FFMPEGARCH}-static.tar.xz.md5" \
    && tar xvf "ffmpeg-release-${FFMPEGARCH}-static.tar.xz" \
    && mv */ffmpeg /usr/local/bin/

WORKDIR /usr/src/mytunes

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 go build -v -o /usr/local/bin ./...

FROM scratch 

COPY --from=build /usr/local/bin/ffmpeg /usr/local/bin/
COPY --from=build /usr/local/bin/mytunes /usr/local/bin/

CMD ["mytunes", "/var/lib/mytunes"]
EXPOSE 8080
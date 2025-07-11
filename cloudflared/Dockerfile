FROM --platform=$BUILDPLATFORM alpine:3.22.0 AS build
RUN apk --update add ca-certificates

FROM scratch
ARG TARGETOS
ARG TARGETARCH
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
ADD --chmod=755 https://github.com/cloudflare/cloudflared/releases/download/2025.6.1/cloudflared-${TARGETOS}-${TARGETARCH} /usr/local/bin/cloudflared
CMD ["cloudflared", "tunnel", "--no-autoupdate", "run"] 
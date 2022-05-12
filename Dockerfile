# Build binary
FROM --platform=$BUILDPLATFORM golang:1.17-alpine AS build-env
ADD . /app
WORKDIR /app
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" . -o echo-server

# Create image
FROM scratch
COPY --from=build-env /app/echo-server /app/
ENV GEO_LITE_2_PATH=/app/config/geolite2/
ENTRYPOINT ["/app/echo-server"]
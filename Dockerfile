# Build a small, statically-linked binary
FROM golang:1.23-alpine AS build
WORKDIR /app
# Download modules first so this layer is cached across source-only changes
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /apod .

# Minimal, non-root runtime image
FROM alpine:3.21
RUN apk add --no-cache ca-certificates \
  && adduser -D -H -u 10001 apod
COPY --from=build /apod /apod
USER apod
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget -q --spider http://localhost:8080/ || exit 1
ENTRYPOINT ["/apod"]

# syntax=docker/dockerfile:1

# --- build ---
FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/server ./cmd/server

# --- runtime ---
FROM alpine:3.21 AS runtime
RUN apk add --no-cache ca-certificates tzdata && adduser -D -u 1001 app
WORKDIR /app
COPY --from=build /out/server /app/server
# Writable dir for the local storage provider (uploads), owned by the app user.
RUN mkdir -p /app/storage && chown -R app:app /app/storage
VOLUME ["/app/storage"]
USER app
EXPOSE 8080
ENTRYPOINT ["/app/server"]

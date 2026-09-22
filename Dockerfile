# Stage 1: Builder
FROM golang:1.23-alpine AS builder
ENV GOTOOLCHAIN=auto
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server ./cmd/server

# Stage 2: Runner with Chromium
FROM alpine:3.19
RUN apk add --no-cache chromium chromium-chromedriver ttf-freefont ca-certificates tzdata
ENV CHROME_BIN=/usr/bin/chromium-browser
WORKDIR /app
COPY --from=builder /bin/server /bin/server
COPY migrations/ /app/migrations/
COPY internal/templates/ /app/templates/
COPY static/ /app/static/
EXPOSE 8080
ENTRYPOINT ["/bin/server"]

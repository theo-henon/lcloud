FROM node:20-alpine AS web-builder

WORKDIR /app/web
COPY web/package.json web/package-lock.json* ./
RUN npm install
COPY web/ ./
RUN npm run build

FROM golang:1.25-alpine AS go-builder

RUN apk add --no-cache git ca-certificates
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=web-builder /app/web/dist ./cmd/server/static

RUN CGO_ENABLED=0 GOOS=linux go build -o /lcloud ./cmd/server

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata su-exec \
  && adduser -D -g '' appuser

WORKDIR /app
COPY --from=go-builder /lcloud /app/lcloud
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["/docker-entrypoint.sh"]

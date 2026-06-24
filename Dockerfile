FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/api ./cmd/api

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tesseract-ocr

COPY --from=builder /app/api /usr/local/bin/api

EXPOSE 10000

CMD ["api"]

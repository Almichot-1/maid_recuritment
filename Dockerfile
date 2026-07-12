FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/api ./cmd/api

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tesseract-ocr tesseract-ocr-data-eng wget && \
    wget -q -O /usr/share/tessdata/ocrb.traineddata \
      "https://raw.githubusercontent.com/Shreeshrii/tessdata_ocrb/master/ocrb.traineddata" && \
    rm -f /var/cache/apk/*

COPY --from=builder /app/api /usr/local/bin/api

EXPOSE 8080

CMD ["api"]

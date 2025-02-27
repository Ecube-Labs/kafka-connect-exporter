FROM golang:1.23.4-alpine AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o exporter ./cmd/exporter

FROM alpine:latest

WORKDIR /

COPY --from=builder /app/exporter .

ENV PORT=9113
EXPOSE 9113

CMD ["./exporter"]
FROM golang:1.23.4-alpine

WORKDIR /

COPY . .

RUN CGO_ENABLED=0 go build -o exporter ./cmd/exporter

ENV PORT=9113
EXPOSE 9113

CMD ["./exporter"]

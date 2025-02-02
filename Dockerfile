FROM golang:1.23-alpine

WORKDIR /

COPY . .

RUN go build -o exporter ./cmd/exporter

# expose port with env
ARG PORT
EXPOSE $PORT

CMD ["./exporter"]

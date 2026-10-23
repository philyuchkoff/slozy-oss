FROM golang:1.27.3-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o slozy ./cmd/slozy-server

EXPOSE 8080

CMD ["./slozy"]

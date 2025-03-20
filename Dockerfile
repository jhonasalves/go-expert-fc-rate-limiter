FROM golang:1.23-alpine

WORKDIR /app

RUN apk add --no-cache git bash && \
    go install github.com/rakyll/hey@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY /cmd/server/.env .

RUN go build -o server cmd/server/main.go

CMD ["./server"]
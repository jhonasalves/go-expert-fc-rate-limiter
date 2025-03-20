FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY . .

RUN GOOS=linux CGO_ENABLED=0 go build -ldflags="-w -s" -o server cmd/server/main.go

FROM scratch

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/cmd/server/.env .

CMD ["./server"]
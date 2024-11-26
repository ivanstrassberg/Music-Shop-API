FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOARCH=amd64 go build -o /app/main ./app/main.go

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app/


COPY --from=builder /app/main /app/main

RUN chmod +x /app/main

COPY .env .env

EXPOSE 8080

CMD ["./main"]

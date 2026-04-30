FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o storage_node ./cmd/main/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/storage_node /storage_node

CMD ["./storage_node"]
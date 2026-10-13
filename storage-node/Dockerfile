FROM golang:1.26.2 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o storage_node ./cmd/main/main.go
RUN CGO_ENABLED=0 go build -o elect ./cmd/elect

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/storage_node /storage_node
COPY --from=builder /app/elect /elect

CMD ["./storage_node"]
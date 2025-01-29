FROM golang:1.23 as builder
LABEL authors="piorus"

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o raft .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/raft .
EXPOSE 1234

CMD ["./raft", "--port", "1234"]
FROM golang:1.26 AS builder
WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o cpu-collector cmd/agent/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/cpu-collector .

CMD ["./cpu-collector"]
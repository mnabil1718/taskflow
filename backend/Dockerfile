FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o taskflow ./cmd

FROM alpine:3.21
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/taskflow .
COPY --from=builder /app/migrations ./migrations
EXPOSE 8080
CMD ["./taskflow"]

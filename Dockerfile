FROM golang:1.25.4-alpine AS builder
WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o pr-management-service ./cmd/pr-management-service


FROM alpine:3.20
RUN addgroup -S app && adduser -S -G app app

WORKDIR /app

COPY --from=builder /app/pr-management-service /app/pr-management-service

RUN chmod +x /app/pr-management-service

USER app

EXPOSE 8080

CMD ["/app/pr-management-service"]
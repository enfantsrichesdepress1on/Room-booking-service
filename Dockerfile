FROM golang:1.23 AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download || true
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /room-booking-service ./cmd/server && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /seed ./cmd/seed

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /room-booking-service /room-booking-service
COPY --from=builder /seed /seed
EXPOSE 8080
ENTRYPOINT ["/room-booking-service"]

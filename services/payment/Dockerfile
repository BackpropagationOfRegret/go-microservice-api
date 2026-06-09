FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.work ./
COPY pkg/ ./pkg/
COPY services/payment/ ./services/payment/
WORKDIR /app/services/payment
ENV GOWORK=off
RUN go mod download
RUN CGO_ENABLED=0 go build -o /payment-svc ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /payment-svc /payment-svc
EXPOSE 8084
ENTRYPOINT ["/payment-svc"]

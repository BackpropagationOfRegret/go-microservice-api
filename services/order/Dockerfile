FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.work ./
COPY pkg/ ./pkg/
COPY services/order/ ./services/order/
WORKDIR /app/services/order
ENV GOWORK=off
RUN go mod download
RUN CGO_ENABLED=0 go build -o /order-svc ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /order-svc /order-svc
EXPOSE 8083
ENTRYPOINT ["/order-svc"]

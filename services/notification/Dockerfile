FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.work ./
COPY pkg/ ./pkg/
COPY services/notification/ ./services/notification/
WORKDIR /app/services/notification
ENV GOWORK=off
RUN go mod download
RUN CGO_ENABLED=0 go build -o /notification-svc ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /notification-svc /notification-svc
EXPOSE 8086
ENTRYPOINT ["/notification-svc"]

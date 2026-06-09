FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.work ./
COPY pkg/ ./pkg/
COPY services/gateway/ ./services/gateway/
WORKDIR /app/services/gateway
ENV GOWORK=off
RUN go mod download
RUN CGO_ENABLED=0 go build -o /gateway ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /gateway /gateway
EXPOSE 8080
ENTRYPOINT ["/gateway"]

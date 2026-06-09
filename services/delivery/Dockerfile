FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.work ./
COPY pkg/ ./pkg/
COPY services/delivery/ ./services/delivery/
WORKDIR /app/services/delivery
ENV GOWORK=off
RUN go mod download
RUN CGO_ENABLED=0 go build -o /delivery-svc ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /delivery-svc /delivery-svc
EXPOSE 8085
ENTRYPOINT ["/delivery-svc"]

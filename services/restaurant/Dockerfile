FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.work ./
COPY pkg/ ./pkg/
COPY services/restaurant/ ./services/restaurant/
WORKDIR /app/services/restaurant
ENV GOWORK=off
RUN go mod download
RUN CGO_ENABLED=0 go build -o /restaurant-svc ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /restaurant-svc /restaurant-svc
EXPOSE 8082
ENTRYPOINT ["/restaurant-svc"]

FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.work ./
COPY pkg/ ./pkg/
COPY services/user/ ./services/user/
WORKDIR /app/services/user
ENV GOWORK=off
RUN go mod download
RUN CGO_ENABLED=0 go build -o /user-svc ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /user-svc /user-svc
EXPOSE 8081
ENTRYPOINT ["/user-svc"]

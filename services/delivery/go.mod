module github.com/kostayne/go-microservice/services/delivery

go 1.22

require (
	github.com/go-chi/chi/v5 v5.1.0
	github.com/google/uuid v1.6.0
	github.com/kostayne/go-microservice/pkg/events v0.0.0
	github.com/kostayne/go-microservice/pkg/kafka v0.0.0
	github.com/lib/pq v1.10.9
)

require (
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/segmentio/kafka-go v0.4.47 // indirect
)

replace (
	github.com/kostayne/go-microservice/pkg/events => ../../pkg/events
	github.com/kostayne/go-microservice/pkg/kafka => ../../pkg/kafka
)

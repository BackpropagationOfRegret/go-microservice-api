module github.com/kostayne/go-microservice/services/gateway

go 1.22

require (
	github.com/go-chi/chi/v5 v5.1.0
	github.com/go-chi/cors v1.2.1
	github.com/kostayne/go-microservice/pkg/auth v0.0.0
	github.com/kostayne/go-microservice/pkg/config v0.0.0
)

require github.com/golang-jwt/jwt/v5 v5.2.1 // indirect

replace (
	github.com/kostayne/go-microservice/pkg/auth => ../../pkg/auth
	github.com/kostayne/go-microservice/pkg/config => ../../pkg/config
)

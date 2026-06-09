module github.com/kostayne/go-microservice/services/user

go 1.22

require (
	github.com/go-chi/chi/v5 v5.1.0
	github.com/google/uuid v1.6.0
	github.com/kostayne/go-microservice/pkg/auth v0.0.0
	github.com/kostayne/go-microservice/pkg/config v0.0.0
	github.com/lib/pq v1.10.9
	golang.org/x/crypto v0.28.0
)

require github.com/golang-jwt/jwt/v5 v5.2.1 // indirect

replace (
	github.com/kostayne/go-microservice/pkg/auth => ../../pkg/auth
	github.com/kostayne/go-microservice/pkg/config => ../../pkg/config
)

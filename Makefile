create-migrations:
	migrate create -ext sql -dir ./migrations -seq init

migrateup:
	migrate -path ./migrations -database 'postgres://postgres:postgres@localhost:5433/gw-currency-wallet?sslmode=disable' up

migratedown:
	migrate -path ./migrations -database 'postgres://postgres:postgres@localhost:5433/gw-currency-wallet?sslmode=disable' down

test-mock:
	mockgen -source=internal/service/service.go -destination=internal/service/mocks/mock_service.go -package=mocks

gen-docs:
	swag init -g ./cmd/main.go -o ./docs

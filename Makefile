create-migrations:
	migrate create -ext sql -dir ./internal/storage/migrations -seq init

migrateup:
	migrate -path ./internal/storage/migrations -database 'postgres://postgres:postgres@localhost:5432/gw-exchanger?sslmode=disable' up

migratedown:
	migrate -path ./internal/storage/migrations -database 'postgres://postgres:postgres@localhost:5432/gw-exchanger?sslmode=disable' down

test-mock:
	mockgen -source=internal/service/service.go -destination=internal/service/mocks/mock_service.go -package=mocks

gen-docs:
	swag init -g ./cmd/main.go -o ./docs

gen-proto:
	protoc --go_out=./proto/gen/exchange --go_opt=paths=source_relative \
	       --go-grpc_out=./proto/gen/exchange --go-grpc_opt=paths=source_relative \
	       ./proto/**/*.proto
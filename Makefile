export LDFLAGS="-w -s"
export POSTGRES_DSN="postgres://healthcheck:healthcheck@localhost:5432/healthcheck?sslmode=disable"

build:
	CGO_ENABLED=1 go build -ldflags $(LDFLAGS)  ./cmd/gokkan

install:
	CGO_ENABLED=1 go install -ldflags $(LDFLAGS) ./cmd/gokkan

check-migrate:
	which migrate || go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate-create: check-migrate
	migrate create -ext sql -dir ./migrations $(NAME)

migrate-up: check-migrate
	migrate -verbose  -path ./migrations -database $(POSTGRES_DSN) up

migrate-down: check-migrate
	 migrate -path ./migrations -database $(POSTGRES_DSN) down

migrate-reset: check-migrate
	 migrate -path ./migrations -database $(POSTGRES_DSN) drop
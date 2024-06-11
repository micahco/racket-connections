docker:
	@docker start racket-connections

run: docker
	go run ./cmd/web -dsn=postgres://postgres:postgres@localhost:5432/postgres
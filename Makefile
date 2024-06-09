build:
	@go build -o bin/server

docker:
	@docker start racket-connections

run: docker build
	./bin/server

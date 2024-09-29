include .env

build: css-minify
	go build -o=./bin/web ./cmd/web

run:
	go run ./cmd/web -dev -port=4000

css:
	tailwindcss -i ./ui/input.css -o ./ui/static/main.css --watch

css-minify:
	tailwindcss -i ./ui/input.css -o ./ui/static/main.css --minify

dev:
	${MAKE} -j2 css run
	
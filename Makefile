include .env

build: css-minify
	go build -o ./bin/main

run:
	go run . -dev -port=4000

css:
	./tailwindcss -i ./ui/input.css -o ./ui/static/main.css --watch

css-minify:
	./tailwindcss -i ./ui/input.css -o ./ui/static/main.css --minify

dev:
	${MAKE} -j2 css run

sync:
	rsync -a ./bin/ ${DO_USER}@${DO_HOST}:/home/${DO_USER}/racket-connections

deploy: build sync
	ssh -t ${DO_USER}@${DO_HOST} 'sudo systemctl restart racket-connections'

connect:
	ssh ${DO_USER}@${DO_HOST}
	
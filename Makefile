include .env

build: css-minify
	go build -o ./bin/main

deploy: build
	rsync -a ./bin/ ${DO_USER}@${DO_HOST}:/home/${DO_USER}/racket-connections

connect:
	ssh ${DO_USER}@${DO_HOST}

docker:
	docker start racket-connections && sleep 1

pg-drop:
	cat ./sql/drop.sql | docker exec -i racket-connections psql -U postgres -d postgres

pg-init:
	cat ./sql/init.sql | docker exec -i racket-connections psql -U postgres -d postgres

pg-sample:
	cat ./sql/sample.sql | docker exec -i racket-connections psql -U postgres -d postgres

css:
	./tailwindcss -i ./ui/input.css -o ./ui/static/main.css --watch

css-minify:
	./tailwindcss -i ./ui/input.css -o ./ui/static/main.css --minify

run: docker pg-drop pg-init pg-sample
	go run . -dev

dev:
	${MAKE} docker && ${MAKE} -j3 css run

include .env

build: css-minify
	go build -o ./bin/main

sync:
	rsync -a ./bin/ ${DO_USER}@${DO_HOST}:/home/${DO_USER}/racket-connections

deploy: build sync
	ssh -t ${DO_USER}@${DO_HOST} 'sudo systemctl restart racket-connections'

connect:
	ssh ${DO_USER}@${DO_HOST}

docker:
	docker start racket-connections && sleep 1

run: docker
	go run . -dev

css:
	./tailwindcss -i ./ui/input.css -o ./ui/static/main.css --watch

css-minify:
	./tailwindcss -i ./ui/input.css -o ./ui/static/main.css --minify

dev:
	${MAKE} docker && ${MAKE} -j3 css run

pg-drop: docker
	cat ./sql/drop.sql | docker exec -i racket-connections psql -U postgres -d postgres

pg-init: docker
	cat ./sql/init.sql | docker exec -i racket-connections psql -U postgres -d postgres

pg-sample: docker
	cat ./sql/sample.sql | docker exec -i racket-connections psql -U postgres -d postgres
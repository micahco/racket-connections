include .env

docker-run:
	docker run --name racket-connections -e POSTGRES_PASSWORD=${RC_DB_PASS} -d -p 5432:5432 postgres

docker-start:
	docker start racket-connections && sleep 1

pg-drop:
	cat ./sql/drop.sql | docker exec -i racket-connections psql -U postgres -d postgres

pg-init:
	cat ./sql/init.sql | docker exec -i racket-connections psql -U postgres -d postgres

pg-sample:
	cat ./sql/sample.sql | docker exec -i racket-connections psql -U postgres -d postgres

css:
	./tailwindcss -i ./templates/main.css -o ./static/main.css --watch

css-minify:
	./tailwindcss -i ./templates/main.css -o ./static/main.css --minify

run: docker-start pg-drop pg-init pg-sample
	go run . \
		-dsn=postgres://postgres:${RC_DB_PASS}@localhost:5432/postgres \
		-smtp-host=${RC_SMTP_HOST} \
		-smtp-pass=${RC_SMTP_PASS} \
		-smtp-port=${RC_SMTP_PORT} \
		-smtp-user=${RC_SMTP_USER} \

dev:
	${MAKE} -j3 css run

build:
	go build -o ./bin/rc

deploy: docker-start css-minify build
	./bin/rc \
		-dsn=postgres://postgres:${RC_DB_PASS}@localhost:5432/postgres \
		-smtp-host=${RC_SMTP_HOST} \
		-smtp-pass=${RC_SMTP_PASS} \
		-smtp-port=${RC_SMTP_PORT} \
		-smtp-user=${RC_SMTP_USER} \
		-prod

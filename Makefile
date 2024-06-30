include .env

build:
	go build -o ./bin/rc-server

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
	go run . \
		-dsn=postgres://postgres:${RC_DB_PASS}@localhost:5432/postgres \
		-smtp-host=${RC_SMTP_HOST} \
		-smtp-pass=${RC_SMTP_PASS} \
		-smtp-port=${RC_SMTP_PORT} \
		-smtp-user=${RC_SMTP_USER} \

dev:
	${MAKE} docker && ${MAKE} -j3 css run

deploy: docker css-minify pg-init build
	./bin/rc-server \
		-dsn=postgres://postgres:${RC_DB_PASS}@localhost:5432/postgres \
		-smtp-host=${RC_SMTP_HOST} \
		-smtp-pass=${RC_SMTP_PASS} \
		-smtp-port=${RC_SMTP_PORT} \
		-smtp-user=${RC_SMTP_USER} \
		-prod

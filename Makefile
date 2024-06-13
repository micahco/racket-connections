include .env

docker-run:
	docker run --name racket-connections -e POSTGRES_PASSWORD=${RC_DB_PASS} -d -p 5432:5432 postgres

docker-start:
	docker start racket-connections && sleep 1

pg-reset:
	cat ./sql/reset.sql | docker exec -i racket-connections psql -U postgres -d postgres

pg-init:
	cat ./sql/init.sql | docker exec -i racket-connections psql -U postgres -d postgres

run: docker-start pg-reset pg-init
	go run ./cmd/web \
		-dsn=postgres://postgres:${RC_DB_PASS}@localhost:5432/postgres \
		-smtp-host=${RC_SMTP_HOST} \
		-smtp-pass=${RC_SMTP_PASS} \
		-smtp-port=${RC_SMTP_PORT} \
		-smtp-user=${RC_SMTP_USER} \
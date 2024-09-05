# Racket Connections

Racket Connections is an online board for OSU students interested in playing court sports and making friends.

## Development

1. Clone the repo.

1. Open directory with VS Code in [Dev Container](https://code.visualstudio.com/docs/devcontainers/containers).

1. Execute the scripts in the `sql` directory to build the database schema and insert mock data:
    ```
    for file in sql/*; do [ -f "$file" ] && cat "$file" | psql -h localhost -d postgres -U postgres ; done
    ```

1. Launch the development server:

    ```
    make dev
    ```

## Environment

Runtime variables are stored in a `.env` file:

    DATABASE_URL=postgresql://postgres:postgres@localhost/postgres
    HOSTNAME=http://localhost:4000
    SMTP_HOST=****
    SMTP_PORT=****
    SMTP_USER=****
    SMTP_PASS=****

# Racket Connections

Racket Connections is an online board for OSU students interested in playing court sports and making friends.

## Development

1. Clone the repo.

2. Open directory with VS Code in [Dev Container](https://code.visualstudio.com/docs/devcontainers/containers).

3. Download the Tailwind CSS [standalone CLI](https://tailwindcss.com/blog/standalone-cli) to the project directory.

4. Execute the scripts in the `sql` directory to build the database schema and insert mock data:
    
    `for file in sql/*; do [ -f "$file" ] && cat "$file" | psql -h localhost -d postgres -U postgres ; done`

5. Launch the development server:

    `make dev`

## Environment

Runtime variables are stored in a `.env` file. Here is what that file should like:

    RC_BASE_URL="http://localhost:4000"
    RC_DB_DSN="postgresql://postgres:postgres@localhost/postgres"
    ...

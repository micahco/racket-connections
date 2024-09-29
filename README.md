# Racket Connections

Racket Connections is an online board for OSU students interested in playing court sports and making friends.


## Development

Guide for VS Code using a [Dev Container](https://code.visualstudio.com/docs/devcontainers/containers).

1. Fork the repository to your GitHub account.

1. Start VS Code and run **Dev Containers: Clone Repository in Container Volume...** from the Command Palette.

1. Select **Clone a repository from GitHub in a Container Volume**.

1. Select the repository that you just forked: **username/racket-connections** then **main** branch to initiate the container build. This may take a while. Once it is done, you should see a bunch of SQL queries being outputted to the dev container terminal, indicating that the post create `setup.sh` script has successfully executed.

1. Open a new terminal window and run `make dev`. This may take a second while the Go CLI downloads the required packages.
## Environment

Runtime variables (secrets) are stored in a `.env` file. See `.env.public` for an example.

## Resources

* [lets-go.alexedwards.net](https://lets-go.alexedwards.net/)

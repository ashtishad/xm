# XM: Company Microservice

## Quick Start

1. Clone your new repository locally with ssh:
   ```
   git clone git@github.com:ashtishad/xm.git
   ```

2. Copy `app.env.example` from the `/local-dev` directory to the project root as `app.env`:
   ```
   cp local-dev/app.env.example app.env
   ```

3. Run `make up` to start the Docker services in the background.

4. Run `make run` to start the application.

5. (Or) Run `make watch` for live reload the application.

Refer to **Makefile** for more details on local development commands.


## Progress

| Requirement                                                            | Status      |
|------------------------------------------------------------------------|-------------|
| User-Auth Endpoints (Login, Register)                                  |    ✅       |
| Companies Endpoints                                                    |    🔄       |
| Postgresql (raw-sql, db transactions, migrations)                      |    ✅       |
| Authentication with JWT ES-256 and route protections                   |    ✅       |
| Kafka Event production for mutating operations                         |    🔄       |
| Dockerization of application                                           |    ✅       |
| Docker setup for external services (postgres, kafka)                   |    ✅       |
| Unit, Fuzz, Integration tests                                          |    🔄       |
| Linter(Golangci-Lint)                                                  |    ✅       |
| Configuration (app.env), managed with viper                            |    ✅       |
| Github Actions Workflow                                                |    ✅       |


## Postman Workspace

https://web.postman.co/workspace/3b0ac53d-a562-4470-9036-2537ea268429

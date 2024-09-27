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
| Companies Endpoints (Create, Get One, Patch, Delete)                   |    ✅       |
| Postgresql (raw-sql, db transactions, migrations)                      |    ✅       |
| Authentication with JWT ES-256 and route protections (authMiddleware)  |    ✅       |
| On each mutating operation, an event should be produced.               |    ✅       |
| Dockerization of application                                           |    ✅       |
| Docker setup for external services (postgres, kafka)                   |    ✅       |
| Unit, Fuzz, Integration tests                                          |    ✅       |
| Linter(Golangci-Lint)                                                  |    ✅       |
| Configuration (app.env), managed with viper                            |    ✅       |
| Github Actions Workflow                                                |    ✅       |

## Tools/Libraries Used

#### Used in the Core API
- [Gin](https://github.com/gin-gonic/gin): HTTP routing and middleware.
- [pgx](https://github.com/jackc/pgx): Database driver and connection pooling, using standard *sql.DB handle.
- [golang-jwt](https://github.com/golang-jwt/jwt): JSON Web Token handling.
- [golang-migrate](https://github.com/golang-migrate/migrate): Database migrations.
- [viper](https://github.com/spf13/viper): For configuration management. (config: app.env)

#### Development Tools
- [swaggo/swag](https://github.com/swaggo/swag): Swagger API documentation.
- [Air](https://github.com/cosmtrek/air): Live reloading.
- [golangci-lint](https://golangci-lint.run/): Linting (config: .golangci.yaml)


## API Collection Documentation

### Postman Setup

1. Import the provided Postman collection, `XM.postman_collection.json`
2. Create an environment and set the following variables:
   - `api_url`:
     If you run app with docker compose and make run: Base URL of the API (e.g., `127.0.0.1:8080/api`)
     If build the docker image, then run the app: Base URL of the API (e.g., `0.0.0.1:8080/api`)
   - `accessToken`: Will be automatically set after login/register

Postman collection link: https://web.postman.co/workspace/3b0ac53d-a562-4470-9036-2537ea268429

### Authentication Endpoints

#### Register User

- **URL**: `POST {{api_url}}/register`
- **Body**:
  ```json
  {
    "name": "John Doe",
    "email": "john@example.com",
    "password": "securepass123"
  }
  ```
- **Success Response**: 201 Created
  ```json
  {
    "user": {
      "userId": "7a6ecf17-e8b0-48a0-a285-ba3ab6e4e708",
      "email": "john@example.com",
      "name": "John Doe",
      "status": "active",
      "createdAt": "2024-09-27T16:55:36.267459Z",
      "updatedAt": "2024-09-27T16:55:36.267459Z"
    }
  }
  ```
- **Error Responses**:
  - 400 Bad Request: Invalid input
    ```json
    {
      "error": "Email must be a valid email"
    }
    ```
  - 409 Conflict: Email already exists
    ```json
    {
      "error": "user with this email already exists"
    }
    ```
  - 500 Internal Server Error
    ```json
    {
      "error": "An unexpected error occurred"
    }
    ```

#### Login

- **URL**: `POST {{api_url}}/login`
- **Body**:
  ```json
  {
    "email": "john@example.com",
    "password": "securepass123"
  }
  ```
- **Success Response**: 200 OK
  ```json
  {
    "user": {
      "userId": "7a6ecf17-e8b0-48a0-a285-ba3ab6e4e708",
      "email": "john@example.com",
      "name": "John Doe",
      "status": "active",
      "createdAt": "2024-09-27T22:55:36.267459+06:00",
      "updatedAt": "2024-09-27T22:55:36.267459+06:00"
    }
  }
  ```
- **Error Responses**:
  - 400 Bad Request: Invalid input
    ```json
    {
      "error": "Email must be a valid email"
    }
    ```
  - 401 Unauthorized: Incorrect password
    ```json
    {
      "error": "incorrect password"
    }
    ```
  - 404 Not Found: User not found
    ```json
    {
      "error": "user not found"
    }
    ```
  - 500 Internal Server Error
    ```json
    {
      "error": "An unexpected error occurred"
    }
    ```

### Company Endpoints

All company endpoints require authentication. The `accessToken` is automatically set as a Bearer token.

#### Create Company

- **URL**: `POST {{api_url}}/companies`
- **Body**:
  ```json
  {
    "name": "TechCorp",
    "description": "Innovative solutions",
    "amountOfEmployees": 100,
    "registered": true,
    "type": "Corporations"
  }
  ```
- **Success Response**: 201 Created
  ```json
  {
    "id": "e3f7c0d3-ccb9-4ce4-926f-ffdbd7fdb687",
    "name": "TechCorp",
    "description": "Innovative solutions",
    "amountOfEmployees": 100,
    "registered": true,
    "type": "Corporations",
    "createdAt": "2024-09-27T22:59:02.406648+06:00",
    "updatedAt": "2024-09-27T22:59:02.406648+06:00"
  }
  ```
- **Error Responses**:
  - 400 Bad Request: Invalid input
    ```json
    {
      "error": "Type must be one of [Corporations NonProfit Cooperative 'Sole Proprietorship']"
    }
    ```
  - 409 Conflict: Company name already exists
    ```json
    {
      "error": "company with this name already exists"
    }
    ```
  - 401 Unauthorized: Missing or invalid token
    ```json
    {
      "error": "Missing authorization header"
    }
    ```
  - 500 Internal Server Error
    ```json
    {
      "error": "An unexpected error occurred"
    }
    ```

#### Get Company

- **URL**: `GET {{api_url}}/companies/{{companyId}}`
- **Success Response**: 200 OK
  ```json
  {
    "id": "e3f7c0d3-ccb9-4ce4-926f-ffdbd7fdb687",
    "name": "TechCorp",
    "description": "Innovative solutions",
    "amountOfEmployees": 100,
    "registered": true,
    "type": "Corporations",
    "createdAt": "2024-09-27T22:59:02.406648+06:00",
    "updatedAt": "2024-09-27T22:59:02.406648+06:00"
  }
  ```
- **Error Responses**:
  - 400 Bad Request: Invalid company ID
    ```json
    {
      "error": "Invalid company ID"
    }
    ```
  - 404 Not Found: Company not found
    ```json
    {
      "error": "company not found"
    }
    ```
  - 401 Unauthorized: Missing or invalid token
    ```json
    {
      "error": "Missing authorization header"
    }
    ```
  - 500 Internal Server Error
    ```json
    {
      "error": "An unexpected error occurred"
    }
    ```

#### Update Company

- **URL**: `PATCH {{api_url}}/companies/{{companyId}}`
- **Body** (all fields optional):
  ```json
  {
    "name": "UpdatedTechCorp",
    "amountOfEmployees": 150,
    "registered": false
  }
  ```
- **Success Response**: 200 OK
  ```json
  {
    "id": "e3f7c0d3-ccb9-4ce4-926f-ffdbd7fdb687",
    "name": "UpdatedTechCorp",
    "description": "Innovative solutions",
    "amountOfEmployees": 150,
    "registered": false,
    "type": "Corporations",
    "createdAt": "2024-09-27T22:59:02.406648+06:00",
    "updatedAt": "2024-09-27T22:59:39.871742+06:00"
  }
  ```
- **Error Responses**:
  - 400 Bad Request: Invalid input or company ID
    ```json
    {
      "error": "Invalid company ID"
    }
    ```
  - 404 Not Found: Company not found
    ```json
    {
      "error": "company not found"
    }
    ```
  - 401 Unauthorized: Missing or invalid token
    ```json
    {
      "error": "Missing authorization header"
    }
    ```
  - 500 Internal Server Error
    ```json
    {
      "error": "An unexpected error occurred"
    }
    ```

#### Delete Company

- **URL**: `DELETE {{api_url}}/companies/{{companyId}}`
- **Success Response**: 204 No Content
- **Error Responses**:
  - 400 Bad Request: Invalid company ID
    ```json
    {
      "error": "Invalid company ID"
    }
    ```
  - 404 Not Found: Company not found or already deleted
    ```json
    {
      "error": "company not found or already deleted"
    }
    ```
  - 401 Unauthorized: Missing or invalid token
    ```json
    {
      "error": "Missing authorization header"
    }
    ```
  - 500 Internal Server Error
    ```json
    {
      "error": "An unexpected error occurred"
    }
    ```

# be-kbtg Backend Service

Lightweight user auth/profile API built with Go, Fiber, GORM (SQLite) and JWT.

## Features
- User registration (/auth/register)
- User login issuing JWT (/auth/login)
- Authenticated profile fetch/update (/me)
- Auto migration of users table
- Simple in-process JWT manager (HS256) with exp claim
- Embedded OpenAPI/Swagger spec (served at /swagger/)

## Tech Stack
- Go + Fiber web framework
- GORM ORM + SQLite (file: app.db)
- golang-jwt/jwt/v5 for token signing
- bcrypt for password hashing

## Directory Structure (excerpt)
```
be-kbtg/
  main.go          # Server wiring & handlers
  auth/            # JWT manager
  config/          # Env config loader
  models/          # DTOs & User model
  docs/            # Swagger spec (docs.go, generated string)
  detail.md        # Sequence + ER diagram (Mermaid)
  app.db           # SQLite database file (generated at runtime)
```

## Environment Variables
| Name        | Default                | Description |
|-------------|------------------------|-------------|
| PORT        | 3000                   | HTTP listen port |
| DB_PATH     | app.db                 | SQLite file path |
| JWT_SECRET  | dev-secret-change      | HS256 signing secret (override in non-dev) |

JWT expiry duration is fixed in code (config.Load) currently at 1h.

## Running Locally
```
cd be-kbtg
go run ./...
```
Server starts on :3000 (or PORT). Database file auto-created; users table auto-migrated.

### Using Docker
A Dockerfile is provided. Example build & run:
```
docker build -t be-kbtg .
docker run --rm -p 3000:3000 -e JWT_SECRET=change-me be-kbtg
```

## API Overview
Base path: (root)

Public:
- POST /auth/register  { email, password } -> 201 Created
- POST /auth/login     { email, password } -> 200 { token }

Protected (Authorization: Bearer <token>):
- GET  /me             -> current user
- PUT  /me             -> update profile fields

Swagger UI: http://localhost:3000/swagger/index.html
Spec JSON: /swagger/doc.json

### Request / Response Examples
Register:
```
curl -X POST http://localhost:3000/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"Passw0rd!"}'
```
Login:
```
TOKEN=$(curl -s -X POST http://localhost:3000/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"Passw0rd!"}' | jq -r .token)
```
Get Profile:
```
curl -H "Authorization: Bearer $TOKEN" http://localhost:3000/me
```
Update Profile:
```
curl -X PUT http://localhost:3000/me \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"first_name":"John","last_name":"Doe"}'
```

## Security Notes
- Passwords stored only as bcrypt hashes
- JWT HS256 secret must be changed in production
- Add HTTPS & refresh token flow for real deployment (see detail.md suggestions)

## Diagrams & Design
See detail.md for sequence diagrams (register/login) and ER diagram of users table.

## Future Improvements
- Refresh token table & rotation
- Role/permission system (RBAC)
- Rate limiting & structured logging
- Pagination & filtering for future list endpoints

## License
Internal workshop/example project.

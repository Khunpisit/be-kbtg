package docs

// Spec is a minimal static Swagger 2.0 JSON describing the auth endpoints.
// Served at /swagger/doc.json
const Spec = `{
  "swagger": "2.0",
  "info": {
    "title": "Auth API",
    "version": "1.0",
    "description": "Simple authentication API with register/login"
  },
  "basePath": "/",
  "paths": {
    "/": {"get": {"summary": "Health","responses": {"200": {"description": "OK"}}}},
    "/auth/register": {
      "post": {
        "summary": "Register user",
        "parameters": [
          {"in":"body","name":"body","schema":{"type":"object","required":["email","password"],"properties":{"email":{"type":"string"},"password":{"type":"string","minLength":6}}}}
        ],
        "responses": {"201": {"description": "Created"},"400":{"description":"Bad Request"},"409":{"description":"Email exists"}}
      }
    },
    "/auth/login": {
      "post": {
        "summary": "Login user",
        "parameters": [
          {"in":"body","name":"body","schema":{"type":"object","required":["email","password"],"properties":{"email":{"type":"string"},"password":{"type":"string"}}}}
        ],
        "responses": {"200": {"description": "OK"},"401":{"description":"Invalid credentials"}}
      }
    }
  }
}`

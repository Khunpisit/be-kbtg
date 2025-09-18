package docs

// Spec is a minimal static Swagger 2.0 JSON describing the auth & profile endpoints.
// Served at /swagger/doc.json
const Spec = `{
  "swagger": "2.0",
  "info": {
    "title": "Auth & Profile API",
    "version": "1.1",
    "description": "Authentication with register/login and profile management"
  },
  "basePath": "/",
  "schemes": ["http"],
  "consumes": ["application/json"],
  "produces": ["application/json"],
  "securityDefinitions": {
    "BearerAuth": {
      "type": "apiKey",
      "name": "Authorization",
      "in": "header",
      "description": "Format: Bearer <token>"
    }
  },
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
    },
    "/me/": {
      "get": {
        "summary": "Get current user profile",
        "security": [{"BearerAuth": []}],
        "responses": {"200": {"description": "OK"},"401":{"description":"Unauthorized"}}
      },
      "put": {
        "summary": "Update current user profile",
        "security": [{"BearerAuth": []}],
        "parameters": [
          {"in":"body","name":"body","schema":{"type":"object","properties":{
            "first_name":{"type":"string"},
            "last_name":{"type":"string"},
            "display_name":{"type":"string"},
            "phone":{"type":"string"},
            "avatar_url":{"type":"string"},
            "bio":{"type":"string","maxLength":500}
          }}}
        ],
        "responses": {"200": {"description": "Updated"},"400":{"description":"Bad Request"},"401":{"description":"Unauthorized"}}
      }
    }
  }
}`

# be-kbtg Contribution & Maintenance Guide

This document defines repository-wide standards, workflows, and best practices for the be-kbtg backend service.

## 1. Purpose & Scope
Maintain a lightweight, secure user auth & profile API. Keep the codebase simple, explicit, and production-ready with minimal tech (Fiber, GORM, SQLite / pluggable DB, JWT).

## 2. Architecture Overview
- Layering (current minimal):
  - main.go: wiring, route registration, lightweight handlers.
  - auth/: JWT manager (sign/parse only).
  - config/: env loading.
  - models/: persistence models + request/response DTO structs + validation methods.
  - docs/: embedded Swagger JSON (static string).
- Server composition: Server struct holds Fiber app, *gorm.DB, auth.Manager.
- No service/usecase layer yet; add one if business logic becomes non-trivial (see 7.3).

## 3. Tech Stack
- Go (>= 1.21 recommended)
- Fiber (HTTP framework)
- GORM (ORM) + SQLite file (default) – easy to swap to Postgres/MySQL.
- golang-jwt/jwt/v5 for HS256 tokens.
- bcrypt (password hashing).

## 4. Local Development
1. cd be-kbtg
2. go run ./...
3. Swagger: http://localhost:3000/swagger/index.html
4. DB created automatically (app.db). Delete file to reset.

Optional helpers to add later: makefile, air (live reload), taskfile.

## 5. Configuration
Environment variables (see readme): PORT, DB_PATH, JWT_SECRET.
- Add new vars in config/config.go with constant + default.
- Avoid introducing implicit behavior; every configurable knob must be explicit.
- Never hardcode secrets; defaults must be clearly unsafe for prod (e.g., dev-secret-change).

## 6. Database & Migrations
- AutoMigrate is used now for User model only.
- If schema evolves (indexes, constraints, historical migrations) switch to a migration tool (e.g., goose / atlas). Checklist:
  1. Introduce /migrations folder.
  2. Disable destructive AutoMigrate operations in production.
  3. Document migration run command in this file.
- Adding a column: Add struct field with GORM tags, run locally (AutoMigrate), update Swagger if exposed.

## 7. Models & DTOs
- models/user.go: persistence entity (gorm.Model fields) + optional domain methods.
- DTO naming: <Action>Request, <Action>Response.
- Validation: Implement Validate() error on request structs (keep pure; no DB calls).
- Keep separation: do not embed DB-specific tags in outward response-specific structs if not needed.
- When adding complex mapping, introduce a mapper function (e.g., ToUserProfileResponse(u User) UserProfileResponse).

## 8. Routing & Handlers
- Register routes in Server.register* methods (grouped by feature).
- Use fiber.Map for simple error responses; structured typed responses for success when stable.
- Keep handlers thin: validation + orchestration + persistence call. Extract business logic if > ~50 lines or reused.

## 9. Authentication & Authorization
- Current: Bearer access token only (no refresh tokens, no roles).
- Future: Introduce roles (enum or bitmask) and middleware (e.g., RequireRole(...roles)).
- Claims: sub (user id), email, exp. Avoid putting mutable fields (display name) into token.
- Rotation: For refresh flow, add refresh_tokens table with hashed tokens + metadata.

## 10. Validation & Error Handling
Consistent error responses: {"error": "message"}
Guidelines:
- 400: client format/validation error
- 401: auth invalid/missing
- 403: authorized user lacks permission (not yet used)
- 404: resource missing (if future endpoints)
- 409: conflict (duplicate email)
- 422 (optional later): semantic validation failure
Centralize repetitive parsing/validation into helper functions as surface grows.

## 11. Logging & Observability (Future)
- Replace log.Printf with structured logger (zerolog / slog) once needed.
- Correlation ID middleware (X-Request-ID) + latency metrics (Prometheus) before scaling.

## 12. Dependency Management
- Use go mod tidy before commits.
- Avoid adding heavy libraries; justify each dependency in PR description.
- Keep auth logic in-house unless adopting a battle-tested abstraction (e.g., OIDC library).

## 13. Testing Strategy
- Unit tests: pure functions (validation, token manager) – fast and isolated.
- Integration tests: spin up Fiber app + ephemeral SQLite (in-memory: file::memory:?cache=shared) for handler tests.
- Add a test package structure:
  /auth (manager tests)
  /models (validation tests)
  /internal/testutil (shared helpers) – create when first needed.
- Run: go test ./...

## 14. API Versioning
- Currently unversioned (/). Introduce /v1 prefix when first breaking change considered.
- Avoid breaking response shapes w/o version bump; additive changes only (new optional fields).

## 15. Swagger / OpenAPI Maintenance
- docs/docs.go is static JSON string.
Workflow to change API:
1. Implement handler change.
2. Update docs.Spec JSON.
3. Verify served JSON & UI.
4. Include snippet diff in PR description.
Consider moving to go:embed + external JSON file or generate via swag when complexity increases.

## 16. Security Practices
- Always hash passwords with bcrypt (cost default ok; consider raising for prod performance envelope).
- NEVER log passwords, tokens, or secrets.
- Use constant-time compares where applicable (not needed for bcrypt output comparisons).
- Ensure JWT secret rotation plan (document in ops runbook). Provide downtime-less strategy: accept old + new during window.
- Add rate limiting (IP + user) before public exposure (Fiber middleware or gateway layer).
- Enforce HTTPS at reverse proxy (doc assumption; app can optionally redirect).

## 17. Performance Considerations
- Fiber + SQLite is fine for workshop; for concurrency scale move to Postgres + connection pool tuning.
- Avoid N+1 queries when related models added (use Preload judiciously).
- Use proper indexes for frequently filtered columns (add GORM index tags or raw migration).

## 18. Contribution Workflow
1. Create branch: feature/<short-desc>, bugfix/<issue-id>-<short>, chore/<task>, docs/<scope>.
2. Use Conventional Commits: feat:, fix:, docs:, chore:, refactor:, test:, perf:, build:, ci:.
3. Keep PRs small (< 400 LOC diff preferred). Split if larger.
4. PR Template (to add later) should include: Summary, Changes, Validation Steps, Screenshots (if HTTP examples), Risk, Rollback Plan.
5. Require at least one reviewer (self-review first: go test, lint, run locally).

## 19. Code Style
- Run go fmt (implicit via goimports / editor tooling).
- Run go vet for suspicious constructs.
- Optional future: staticcheck.
- Naming: exported types/methods only when needed externally; keep package surface minimal.
- Avoid long parameter lists – refactor into struct when > 4 logically related values.

## 20. Release / Deployment (Placeholder)
- Tag pattern: vX.Y.Z (semantic version once versioning starts).
- Changelog generated from Conventional Commits (future tooling: git-chglog / goreleaser). For now manual notes.

## 21. Directory Structure Snapshot
be-kbtg/
  main.go
  auth/
  config/
  models/
  docs/
  detail.md (design diagrams)
  readme.md (overview)
  CONTRIBUTING.md (this file)
  app.db (runtime) – DO NOT COMMIT if replaced with persistent path

## 22. Adding a New Endpoint Checklist
- [ ] Define request/response DTO in models/ (validation method if needed).
- [ ] Add route registration in register*Routes or new grouping function.
- [ ] Implement handler (≤ ~60 lines; extract logic if bigger).
- [ ] Add tests (happy path + failure cases).
- [ ] Update swagger JSON.
- [ ] Update README examples if user-facing.
- [ ] Consider security implications (auth? rate limit?).
- [ ] Run go vet & go test.

## 23. Adding a Config Variable
- [ ] Add const EnvX in config/config.go.
- [ ] Add field to Config struct.
- [ ] Load in Load() with sensible default.
- [ ] Thread through where required (avoid global state).
- [ ] Document in README (Environment Variables table).

## 24. Handling Secrets
- Never commit real secrets.
- Use environment variables or secret manager in production.
- Provide sample .env.example (future) – do not commit real .env.

## 25. Future Enhancements (Tracking Alignment)
Refer to readme.md Future Improvements; ensure new contributions align: refresh tokens, RBAC, logging, rate limiting, pagination, etc.

## 26. Documentation Quality Bar
- Every externally visible change must update README / Swagger / diagrams (if impacted).
- Keep diagrams (detail.md) synchronized; stale diagrams create onboarding friction.

## 27. Decision Records (Lightweight)
For non-trivial architectural decisions (e.g., switching DB, introducing service layer), add docs/adr-<yyyymmdd>-<slug>.md summarizing context, decision, alternatives, consequences.

## 28. Open Questions / To Refine
- Introduce integration test harness? (Pending scale.)
- Introduce structured logging replacement? (When > basic debugging needed.)
- Introduce context propagation & request IDs? (When observability begins.)

## 29. Contact / Ownership
Owned by workshop team. Escalation: create issue or tag maintainer in PR.

---
Adhere to this guide to keep the codebase consistent, secure, and maintainable. Propose updates via PR when process changes are needed.

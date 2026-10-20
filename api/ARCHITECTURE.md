# API Architecture

The API is a single Go binary with separate `server`, `worker`, and migration
commands. The HTTP and worker processes share capability packages, generated
database access, and infrastructure adapters, but each process builds only the
runtime graph it needs.

## Dependency flow

```text
HTTP / queue transport
        |
        v
capability services and read models
        |                |
        |                v
        |         sqlc queries (directly, or via an
        |         unexported store in the same package
        |         when writes need a transaction)
        v
ports (interfaces owned by the consuming capability)
        ^
        |
queue, Shopware, mail, Redis, and S3 adapters
```

`server.go` and `worker.go` are composition roots. They are the places where
concrete adapters are constructed and passed to services. Business packages
must not reach back into a composition root or HTTP package.

## Package roles

- `internal/handler` and `internal/auth` are HTTP adapters. They translate
  OpenAPI requests into service commands and map service results or errors to
  HTTP responses. They do not query PostgreSQL or call queues, mail, storage,
  or Redis directly.
- `internal/middleware` authenticates requests through the
  `identity.SessionAuthenticator` boundary and stores the neutral
  `access.Principal` in the request context.
- `internal/identity`, `organization`, `monitoring`, `deployment`, `catalog`,
  `notification`, `packagesmirror`, `suppression`, and `audit` own application
  behavior and their persistence. Services either use the generated sqlc
  queries directly or through an unexported `store` in the same package when
  writes need a transaction (organization creation, credential changes,
  monitoring and deployment writes). There are no `postgres` adapter packages.
  Interfaces are defined only for external systems (Shopware, OIDC, mail,
  queues, storage), cross-capability authorization, and internal seams a
  unit test substitutes (organization authorization, session authentication,
  monitoring orchestration, job scheduling). Capability packages contain no
  HTTP types.
- Adapter subpackages such as `queue`, `shopware`, `s3output`, `oidc`,
  `credentialmail`, and `redisstate` implement capability-owned ports for
  external systems.
- `internal/readmodel` owns query-side projections for endpoints that assemble
  data from several tables. Read models may depend on sqlc and HTTP DTO types
  from `internal/api`; they must not contain mutations or authorization
  policy.
- Capability-owned worker packages such as `monitoring/scrape`,
  `monitoring/sitespeed`, `catalog/sync`, and `catalog/changelog` implement job
  behavior. `internal/jobs` contains stable queue message contracts, the
  dispatch-only bus, worker handler registration, and recurring scheduling
  rather than business workflows. Only the worker registers executable
  handlers; the API and fixture processes merely dispatch messages.
- `internal/api` and `internal/authapi` hold hand-written HTTP DTO types used
  by handlers and read models. Operations are registered with Huma (`huma.Register`)
  from `internal/handler` and `internal/auth` onto chi via `humachi`.
- OpenAPI is generated from registered Huma routes (`shopmon openapi`).
  `mise run generate` dumps it to gitignored `openapi/spec.yaml` only as input
  for frontend type generation. Do not commit or edit that dump.
- `internal/database/queries` is generated code. Change the SQL sources and
  regenerate them; do not edit generated files.

## Boundary rules

1. HTTP packages depend on capability services and read models, never concrete
   persistence or infrastructure clients.
2. A capability defines the interface it consumes for external systems and
   cross-capability calls; concrete implementations live in a subpackage and
   are connected at a composition root. Persistence needs no port or adapter
   package — the service takes `*queries.Queries` (and the pool, when it runs
   transactions) and owns its queries in-package.
3. Capability packages do not import HTTP handlers, auth transports, or other
   composition roots.
4. Cross-capability authorization uses small interfaces, such as organization
   membership authorization, instead of database access.
5. Transactions that protect an invariant live in one store operation in the
   owning package, such as creating an organization with its owner or
   accepting an invitation — never spread across service calls.
6. Query projections are the explicit CQRS exception: they may return
   endpoint-shaped data, but mutations stay in capability services.

The tests in `internal/architecture` enforce the most important import
boundaries.

## Request lifecycle

1. Chi and OpenTelemetry middleware receive the request.
2. Optional authentication validates a bearer token and attaches an
   `access.Principal`.
3. Huma (via the Chi adapter) matches the operation and calls the handler.
4. The handler checks the principal and delegates to a capability service or
   read model.
5. The service applies authorization and business rules, then calls sqlc
   queries and its ports.
6. Adapters perform queue, remote API, mail, or storage work.
7. The handler returns a Huma output or a `{"message":"..."}` status error.

## Job lifecycle

1. The scheduler or a capability service explicitly dispatches a typed message
   defined by `internal/jobs` to the stable asynchronous transport.
2. The worker composition root registers capability-owned services for those
   message types; dispatch-only processes do not construct worker services.
3. go-queue supplies the context and OpenTelemetry span; the worker service
   performs the workflow and returns a contextual error for retry handling.

## Adding behavior

For a new API mutation:

1. Register a Huma operation with the existing `operationId` and implement the
   handler in `internal/handler` or `internal/auth`. OpenAPI is generated from
   the registered routes; do not maintain a spec file.
2. Add the command and operation to the owning capability service.
3. For persistence, call sqlc queries from the service or its in-package
   store (add a query in `sql/queries/` and run `mise run generate`); keep
   multi-statement writes inside one store operation with a transaction. For
   an external system, add or extend a capability-owned port and implement it
   in an adapter subpackage.
4. Wire the service (and any new adapter) in `server.go`.
5. Keep the handler limited to request mapping, authorization context, service
   invocation, and response mapping.
6. Run `mise run generate` to dump OpenAPI locally and regenerate frontend
   API types (`frontend/src/api/generated/`).

For a new query spanning several aggregates, add it to an appropriate read
model. For a new background workflow, define a typed job contract and keep its
implementation in the owning capability package.

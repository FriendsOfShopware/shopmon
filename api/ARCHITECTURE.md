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
        |
        v
ports (interfaces owned by the consuming capability)
        ^
        |
PostgreSQL, queue, Shopware, mail, Redis, and S3 adapters
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
  `notification`, `packagesmirror`, and `audit` own application behavior and
  their required interfaces. Their top-level packages contain no HTTP or
  generated SQL types.
- Adapter subpackages such as `postgres`, `queue`, `shopware`, `s3output`,
  `oidc`, `credentialmail`, and `redisstate` implement capability-owned ports.
- `internal/readmodel` owns query-side projections for endpoints that assemble
  data from several tables. Read models may depend on sqlc and generated
  OpenAPI response types; they must not contain mutations or authorization
  policy.
- Capability-owned worker packages such as `monitoring/scrape`,
  `monitoring/sitespeed`, `catalog/sync`, and `catalog/changelog` implement job
  behavior. `internal/jobs` contains stable queue message contracts, the
  dispatch-only bus, worker handler registration, and recurring scheduling
  rather than business workflows. Only the worker registers executable
  handlers; the API and fixture processes merely dispatch messages.
- `internal/api`, `internal/authapi`, and `internal/database/queries` are
  generated code. Change their OpenAPI or SQL sources and regenerate them; do
  not edit generated files.

## Boundary rules

1. HTTP packages depend on capability services and read models, never concrete
   persistence or infrastructure clients.
2. A capability defines the interface it consumes. Concrete implementations
   live in a subpackage and are connected at a composition root.
3. Capability packages do not import HTTP handlers, auth transports, or other
   composition roots.
4. Cross-capability authorization uses small interfaces, such as organization
   membership authorization, instead of database access.
5. Transactions that protect an invariant live behind a repository operation,
   such as creating an organization with its owner or accepting an invitation.
6. Query projections are the explicit CQRS exception: they may return
   endpoint-shaped data, but mutations stay in capability services.

The tests in `internal/architecture` enforce the most important import
boundaries.

## Request lifecycle

1. Chi and OpenTelemetry middleware receive the request.
2. Optional authentication validates a bearer token and attaches an
   `access.Principal`.
3. The generated OpenAPI router calls an auth or API handler.
4. The handler checks the principal and delegates to a capability service or
   read model.
5. The service applies authorization and business rules, then calls its ports.
6. Adapters perform database, queue, remote API, mail, or storage work.
7. The handler maps the result with `httputil.WriteJSON` or
   `httputil.WriteError`.

## Job lifecycle

1. The scheduler or a capability service explicitly dispatches a typed message
   defined by `internal/jobs` to the stable asynchronous transport.
2. The worker composition root registers capability-owned services for those
   message types; dispatch-only processes do not construct worker services.
3. go-queue supplies the context and OpenTelemetry span; the worker service
   performs the workflow and returns a contextual error for retry handling.

## Adding behavior

For a new API mutation:

1. Update `openapi/spec.yaml` and regenerate the server interfaces.
2. Add the command and operation to the owning capability service.
3. Add or extend a capability-owned port and implement it in an adapter
   subpackage if persistence or an external system is needed.
4. Wire the adapter and service in `server.go`.
5. Keep the handler limited to request mapping, authorization context, service
   invocation, and response mapping.

For a new query spanning several aggregates, add it to an appropriate read
model. For a new background workflow, define a typed job contract and keep its
implementation in the owning capability package.

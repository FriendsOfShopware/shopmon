# Project Architecture & Development Guidelines

This project follows a **Domain-Driven Modular Architecture**.
This structure groups related functionality by **Domain (Feature)** rather than by Technical Layer.

## Directory Structure

```
api/src/
  ├── modules/
  │   ├── shop/                 <-- Domain Module
  │   │   ├── shop.repository.ts   (Data Access)
  │   │   ├── shop.service.ts      (Business Logic)
  │   │   ├── shop.router.ts       (API / Transport)
  │   │   └── shop.types.ts        (Domain Types)
  │   ├── user/
  │   └── ...
  ├── trpc/                     <-- Shared / Legacy Routers
  ├── service/                  <-- Shared / Legacy Services
  └── repository/               <-- Shared / Legacy Repositories
```

## The 3-Layer Pattern (Applied per Module)

Within each module, we strictly enforce the 3-Layer Pattern:

### 1. Router (`*.router.ts`)
*   **Responsibility:**
    *   Handle HTTP/tRPC requests.
    *   Validate inputs (Zod schemas).
    *   Check permissions/middleware.
    *   **DELEGATE** work to the Service.
*   **Rules:**
    *   ❌ **NO** direct database queries.
    *   ❌ **NO** complex business logic.
    *   ✅ **ONLY** call Services.

### 2. Service (`*.service.ts`)
*   **Responsibility:**
    *   Contain **ALL** business logic.
    *   Orchestrate workflows.
    *   Handle third-party integrations (Stripe, Mailer, etc.).
*   **Rules:**
    *   ✅ Can call Repositories.
    *   ✅ Can call other Services.
    *   ❌ Should not deal with HTTP-specifics.

### 3. Repository (`*.repository.ts`)
*   **Responsibility:**
    *   Pure Data Access Object (DAO).
    *   Handle **CRUD** operations.
*   **Rules:**
    *   ✅ **ONLY** interact with the database.
    *   ❌ **NO** business logic.
    *   ❌ **NO** side effects (sending emails).
    *   ❌ **NEVER** call Services.

---

## Refactoring Checklist

When working on existing code:

1.  **Identify the Domain:** Does this belong to `Shop`, `User`, `Project`, etc.?
2.  **Create/Update Module:** Move files to `api/src/modules/<domain>/`.
3.  **Rename Files:** Use the `entity.layer.ts` convention (e.g., `user.service.ts`).
4.  **Refactor Logic:** Ensure logic is in the Service, DB calls in the Repository, and Validation in the Router.
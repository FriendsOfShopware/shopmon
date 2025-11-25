# Project Architecture & Development Guidelines

This project follows a strict **3-Layer Architecture** to ensure separation of concerns, testability, and maintainability. All new features and refactors must adhere to this pattern.

## The 3-Layer Pattern

### 1. Router Layer (`api/src/trpc/**/*.ts`)
*   **Responsibility:**
    *   Handle HTTP/tRPC requests.
    *   Validate inputs (Zod schemas).
    *   Check permissions/middleware.
    *   **DELEGATE** work to the Service layer.
*   **Rules:**
    *   ❌ **NO** direct database queries (no `drizzle` or `db` imports for business logic).
    *   ❌ **NO** complex business logic.
    *   ✅ **ONLY** call `Services`.

### 2. Service Layer (`api/src/service/**/*.ts`)
*   **Responsibility:**
    *   Contain **ALL** business logic.
    *   Orchestrate workflows (e.g., "Create User" -> "Save to DB" -> "Send Email").
    *   Call multiple Repositories if needed.
    *   Handle third-party integrations (Stripe, Mailer, etc.).
*   **Rules:**
    *   ✅ Can call `Repositories`.
    *   ✅ Can call other `Services`.
    *   ✅ Can call Helpers/Utils.
    *   ❌ Should not deal with HTTP-specifics (like `req`, `res`).

### 3. Repository Layer (`api/src/repository/**/*.ts`)
*   **Responsibility:**
    *   Pure Data Access Object (DAO).
    *   Handle **CRUD** operations (Create, Read, Update, Delete).
    *   Abstract the database technology (Drizzle).
*   **Rules:**
    *   ✅ **ONLY** interact with the database.
    *   ❌ **NO** business logic (e.g., "if user is premium...").
    *   ❌ **NO** side effects (sending emails, calling external APIs).
    *   ❌ **NEVER** call Services.

---

## Refactoring Checklist

When working on existing code, ensure it aligns with this structure:

1.  **Identify the "Fat":** Look for Routers containing DB queries or logic, or Repositories sending emails.
2.  **Extract to Service:** Move the logic into a dedicated file in `api/src/service/`.
3.  **Simplify Repository:** Ensure the Repository only does DB operations.
4.  **Slim the Router:** Update the Router to simply call the new Service function.

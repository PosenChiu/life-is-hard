# Project Blueprint for AI Agents

**Purpose**: This README serves as the single source of truth for AI-driven code development on the project. It consolidates the project’s goals, structure, development principles, and guidelines into one clear document. By following this blueprint, an AI (or any developer) can quickly understand the project structure, adhere to development rules (including functional programming practices), respect authorization boundaries, and effectively use the reference files (`CODE.md` and `LAYOUT.md`) to maintain and extend the backend system.

______________________________________________________________________

## 1. Project Goal

The goal of this project is to **maintain and expand a Golang-based backend API system**. This system is a RESTful API server written in Go, primarily using the Echo framework, with a PostgreSQL database and a Redis cache. Current features include user management and authentication (e.g. user CRUD operations, JWT-based login, OAuth client management, health check endpoint, etc.). The project should continue to grow with new features and improvements while maintaining reliability and code quality.

Key objectives of the project include:

- **Stability**: Ensure existing API functionality (user authentication, authorization, data operations) remains stable and well-tested as new features are added.
- **Extensibility**: Implement new requirements promptly and cleanly, integrating them into the existing architecture without hacks or shortcuts.
- **Consistency**: Follow established patterns and practices so that new code blends seamlessly with the old (in terms of style, structure, and conventions).
- **Documentation**: Keep the code and API documentation (Swagger) up-to-date as the system evolves, so that both the AI and human developers can easily understand the API behavior and usage.

This README outlines the rules and guidelines that the AI agent (acting as a developer) must follow to achieve these goals. By adhering to these guidelines, the AI can autonomously implement new features or changes in the project without needing step-by-step human oversight.

______________________________________________________________________

## 2. Development Principles

To maintain a high-quality codebase, the following development principles and best practices must be followed. These principles emphasize a **functional programming** style and clean architecture, ensuring the code remains modular, testable, and easy to reason about:

- **Functional Programming Approach**: Favor a functional programming style wherever possible. This means using pure functions and avoiding side effects and global mutable state. Dependencies (like database connections, caches, etc.) should be passed as arguments rather than stored as global variables or singletons. Functions should return data (and errors) instead of modifying shared state. This approach makes the code easier to test and reason about. For example, database operations are implemented as functions that accept a context and a DB connection pool as parameters, rather than methods on an object with internal state.

- **Separation of Concerns (Layered Architecture)**: Adhere strictly to the project’s layered structure. Each layer has a single responsibility, and code should be placed in the appropriate folder (see **Project Structure** below). Business logic must be separated from HTTP handlers, data access separated from business logic, etc. In practice:

  - **Handlers** (in `internal/handler/`): Contain no business logic. They are responsible for parsing input (e.g., request JSON or form data), calling the appropriate service functions, and formatting the output (HTTP response). Handlers also handle request validation and request/response binding. They include Swagger annotations for API documentation but should delegate actual work to the service layer.
  - **Services** (in `internal/service/`): Implement business logic and orchestrate repository calls. A service function will typically call one or more repository functions, apply business rules, and return results to the handler. Services act as an intermediary between handlers and repositories when business processes are non-trivial. (For simple data operations, handlers might call repositories directly, but ideally complex logic resides in services.)
  - **Repositories** (in `internal/repository/`): Encapsulate data access logic. This is where we interact with the PostgreSQL database or Redis cache. Repository functions perform CRUD operations (Create, Read, Update, Delete) and other queries. They should **wrap errors** with context (e.g., using `fmt.Errorf("OperationName: %w", err)`) so that lower-level errors carry meaning up the stack without exposing internal details. Repositories must not contain business decision logic — they purely retrieve or manipulate data.
  - **Models** (in `internal/model/`): Define the data structures (structs) that map to database tables or represent core entities. These structs use struct tags for database mapping (e.g. `db:""` for SQL mapping) and JSON serialization (`json:""` for API responses). Models should be simple data containers with little or no methods, behaving like typed records.
  - **DTOs (Data Transfer Objects)** (in `internal/dto/`): Define request and response schemas for the API. These may duplicate fields from models or combine multiple models for a given API. DTO structs include validation tags (using e.g. `validator` library) for request binding and have Swagger annotations for automatic documentation generation. This ensures the API documentation is always up-to-date with the code.
  - **Middleware** (in `internal/middleware/`): Provide cross-cutting concerns that apply to many routes (authentication checks, logging, rate limiting, etc.). Middleware functions should remain generic and reusable, not containing business logic but rather concerns like security and instrumentation.
  - **Router** (in `internal/router/`): Defines how routes (endpoints) are set up and mapped to handlers, and what middleware or security policies apply to each route group. This is where URL paths are constructed and grouped (for example, all user-related routes under `/api/users`). The router should configure the Echo server with all handlers and middleware in one place.

- **Functional Error Handling**: Always handle errors at each layer boundary. Functions should return errors rather than panicking. Use Go’s error wrapping to add context (as mentioned for repository functions). The goal is that when an error surfaces (for example, in an HTTP response), it is informative but not revealing sensitive internals. Propagate errors up and let the handler decide how to translate them into HTTP responses (often as JSON error messages defined in `internal/dto.HTTPError`). This principle ensures reliability and debuggability.

- **Immutability and Stateless Design**: Wherever possible, treat data as immutable. Rather than modifying an object in place, consider returning a new value. Avoid global state; for example, do not use package-level variables for database connections or configuration. Instead, pass required data through function parameters or context. This minimizes unintended interactions between parts of the code. The use of dependency injection (for instance, passing the DB pool and Redis client to handlers or services) is preferred over global singletons.

- **Consistency and Naming Conventions**: Follow the established naming conventions and coding style throughout the project. Use idiomatic Go style (meaning clear variable names, exported functions with comments, etc.). Some specific conventions in this project:

  - **Handler and Route Naming**: The subfolder name under `internal/handler/` should correspond to the first segment of the API path. For example, handlers in `internal/handler/users/` implement routes under `/api/users/...`, and `internal/handler/auth/` corresponds to routes under `/api/auth/...`. This makes the structure intuitive: one can find the handler for an endpoint by following the path naming.
  - **DTO Naming**: Request DTOs and response DTOs are typically named with `Request` or `Response` suffixes (e.g., `LoginRequest`, `UserResponse`) and often correspond to Swagger models.
  - **File Names**: Use lowercase with underscores for multi-word file names. For instance, `create_user.go` contains the handler for creating a user. Tests for a file `user.go` would be in `user_test.go` in the same directory.
  - **Database Migrations**: Migration files under `internal/db/migrations` are prefixed with an incremental number and descriptive name (e.g., `0001_create_users.up.sql`). The file naming and numbering should remain consistent to ensure they run in order.

- **Complete Testing**: Whenever a new piece of functionality is added at the repository or service layer, include unit tests for it. The repository layer in particular should have corresponding `_test.go` files that test database interactions (where feasible; possibly using a test database or mocking the database). Business logic in the service layer should also be covered by tests to verify that all rules work as expected. The testing principle ensures regressions are caught and that the AI’s changes do not break existing features.

- **Swagger Documentation**: All new endpoints must be documented via Swagger comments. The project uses `swaggo/swag` to generate API documentation from code annotations. For example, a handler function will have comment lines starting with `// @Summary`, `// @Description`, `// @Tags`, etc., describing the endpoint. The AI must continue this practice: any new handler should include proper annotations so that running `swag init` updates the API docs. This keeps documentation in sync with implementation.

- **License and External Code**: (Authorization Boundary) Only use external libraries that are approved for the project and compatible with its license (if any). Avoid copying any code from external sources unless it’s under a compatible license and necessary for the task. Generally, the AI should stick to writing original code or using the dependencies already present in `go.mod`. If a new dependency is truly needed, it should be added with caution and noted. The AI is **not authorized** to use any proprietary or unlicensed code. This principle ensures we respect all licensing constraints and project policies.

By following these principles, the AI will produce code that is maintainable, consistent with the project's style, and within the boundaries of what is allowed in this project’s context.

______________________________________________________________________

## 3. Project Structure and Key Artifacts

Understanding the project layout is crucial for navigating the code and placing new components in the correct location. The project is organized in a predictable manner. The following reference artifacts are provided to help the AI (or any developer) quickly find files and understand the existing implementation:

### 3.1 Artifacts Overview

| Artifact                 | Description                                                                                                                                                                                                                                                                                                                                                                                                    |
| ------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **LAYOUT.md**            | Project directory structure (all key files and folders).                                                                                                                                                                                                                                                                                                                                                       |
| **CODE.md**              | Aggregated code snippets and templates covering handlers, models, repositories, services, and migrations. This serves as a reference for how things are implemented.                                                                                                                                                                                                                                           |
| **cmd/service/main.go**  | Entrypoint of the application. Sets up the server, performs dependency injection (connecting to DB and Redis, initializing handlers/services), applies middleware, and starts the Echo server. Also registers Swagger (API documentation) setup.                                                                                                                                                               |
| **internal/db/**         | Database and cache initialization, plus migration files. Contains migration runner logic (`migrations.go`) and SQL files for each migration (`.up.sql` and `.down.sql` files), as well as any caching setup (Redis).                                                                                                                                                                                           |
| **internal/model/**      | Data models defining the structure of database tables and other core entities. These are Go structs tagged with `db` (for database fields) and `json` (for API responses) to map to the DB and API.                                                                                                                                                                                                            |
| **internal/repository/** | Data access layer with functions to interact with the database and cache. Contains SQL queries and Redis calls for CRUD operations. Errors are wrapped here and returned for handling by the service or handler. Each repository file should have a corresponding `_test.go` for unit tests.                                                                                                                   |
| **internal/service/**    | Business logic layer. Functions here implement higher-level operations by calling repository functions and applying business rules. This layer ensures that handlers remain thin. For example, authentication, password hashing, or complex multi-step operations would be in the service layer.                                                                                                               |
| **internal/dto/**        | Data Transfer Objects. Defines request structures (for binding and validation) and response structures (for consistent JSON output). Also contains error response formats (e.g., `HTTPError`) and includes Swagger annotations for documentation generation.                                                                                                                                                   |
| **internal/handler/**    | HTTP handler functions grouped by feature area or route prefix. Each subfolder corresponds to an API endpoint group (e.g., `users`, `auth`). Handlers parse incoming requests, call services or repositories, and write JSON responses. Swagger annotations are used here to document each endpoint’s behavior, parameters, and responses.                                                                     |
| **internal/middleware/** | Reusable middleware components for the Echo server. This includes things like `RequireAuth` (to enforce JWT authentication on routes), `RequireAdmin` (to enforce admin role), logging middleware, etc. Middleware functions are typically used when registering routes to apply cross-cutting concerns.                                                                                                       |
| **internal/router/**     | The router setup for the API. Contains the function that registers all routes and associates them with their handlers and any route-specific middleware. For instance, it sets up the `/api` group, then registers `/api/ping`, `/api/auth/login`, `/api/users` routes, etc., linking them to the correct handler functions and adding security (like attaching `RequireAuth` or `RequireAdmin` where needed). |
| **config.mk**            | Build and environment configuration file (Makefile format). Contains environment variable defaults (like `DATABASE_URL`, `REDIS_ADDR`, etc.) and can be used to set up the environment for running the service.                                                                                                                                                                                                |
| **go.mod**, **go.sum**   | Go module files listing dependencies. These track all external packages used (Echo, JWT library, validator, etc.) and ensure reproducible builds.                                                                                                                                                                                                                                                              |

This overview gives a high-level map of the project. Whenever the AI needs to find where something should go or where to look for existing logic, it can refer to the above artifact descriptions or check the actual content of **`LAYOUT.md`** and **`CODE.md`**:

- **`LAYOUT.md`** provides a tree view of the directory structure. It is useful for verifying if a file or module already exists or determining the correct path for a new file. For example, if adding a new feature "orders", the AI can check `LAYOUT.md` to see if an `internal/handler/orders/` directory (or similar model/repo) exists, or if it needs to create one.
- **`CODE.md`** contains aggregated code snippets from the project. The AI can search within this file to find examples of how certain operations are implemented. For instance, to see how a typical repository function is written or how error handling is done, the AI can find the relevant section in `CODE.md`. This helps in maintaining consistency (using similar patterns and not reinventing the wheel if a utility function exists). It effectively acts as in-project documentation for code implementation details.

### 3.2 Folder Responsibilities

To reiterate the responsibilities of each folder in the project (the clean architecture structure), here’s a summary with emphasis on what each part should and should not do:

- **`cmd/`**: Contains the application entry point. In our case, `cmd/service/main.go` configures the server. It loads configuration (from env vars or config files), initializes the database connection pool and Redis client, sets up an Echo instance, attaches middleware, registers all routes (by calling `internal/router.Setup`), and finally starts the HTTP server. This is the only place where all components come together (Dependency Injection happens here by passing `db` and `rdb` to handlers/services). There should be minimal logic here aside from orchestrating startup. If more complex initialization is needed, it can be broken out into functions, but generally `main.go` should remain concise.

- **`internal/db/`**: Handles database and cache interactions at a low level. It includes migration files (SQL scripts to set up the schema). The `migrations.go` file usually contains code to run all `.up.sql` files in order or roll them back, using a library like golang-migrate. There may also be helper functions to connect to the DB (`NewPool`) and ping the DB or to initialize the Redis client. This folder should not contain business logic, only database schema management and connectivity concerns.

- **`internal/model/`**: Defines the data models corresponding to database tables or other fundamental data structures. Each model is typically a Go struct with fields and appropriate tags. For example, a `User` model struct with fields `ID, Name, Email, PasswordHash, IsAdmin, CreatedAt`. Models might also define methods for convenience (like table name constants or minor helper methods), but heavy logic belongs in services or repositories. The model layer is pure data definition.

- **`internal/repository/`**: Data access layer. For each model (or related set of models), there is typically a repository file (e.g., `user.go` for user-related DB operations, `oauth_client.go` for OAuth client operations). Repositories contain functions that interact with the database (using SQL queries via the pgx pool) or the cache (Redis). All SQL statements live here (or in migrations for table creation). Repositories must handle errors from the database and wrap them with additional context messages. They then return either the requested data or an error. They **should not** make business decisions; for example, a repository function should fetch data, but deciding if a user is allowed to fetch that data would be a service or handler concern. Also, repository functions should be **unit tested**. For any new repository function, the AI should create a corresponding test in a `_test.go` file to validate expected behavior (using test database transactions or mocks as appropriate).

- **`internal/service/`**: Business logic layer. Not every feature will have a dedicated service function if the logic is simple, but any non-trivial operation should. Services typically call one or more repository functions and implement additional logic. For example, the authentication service might verify a password by fetching user data via the repository and then comparing hashes, or a user creation service might include sending a welcome email (if that were in scope). In our current context, services also contain utility logic like password hashing or token generation (`service.HashPassword`, `service.GenerateJWT` etc. might exist here). The service layer helps keep handlers thin and allows reuse of business logic (e.g., the same service function could be invoked by different handlers or maybe by scheduled tasks in the future). Services should also be covered by tests when they contain important logic.

- **`internal/dto/`**: Data Transfer Objects. This folder contains structures that define what a client sends and receives via the API. For example, `LoginRequest` (with fields for username and password), `UserResponse` (with fields a user should see, possibly excluding sensitive info), and error response formats. These DTOs often overlap with model fields but are distinct to allow decoupling of internal representation from external API. Each DTO can have validation rules using the `validator` library (e.g., `validate:"required"` tags) and JSON tags for proper serialization. Importantly, the DTO definitions are annotated for Swagger. Comments starting with `// swagger:...` above a struct or field will be used by `swaggo/swag` to generate API documentation. When the AI adds a new endpoint, it should ensure that any new request or response struct in this folder has the necessary documentation comments (see existing DTOs in `CODE.md` for examples).

- **`internal/handler/`**: HTTP Handlers divided by feature area. Each handler package corresponds to a route group. For example, `internal/handler/users/` contains handlers for user-related endpoints (`CreateUserHandler`, `GetUserHandler`, etc.), and `internal/handler/auth/` might contain `LoginHandler`, etc. A handler function typically looks like `func SomeHandler(dependencies) echo.HandlerFunc { return func(c echo.Context) error { ... } }`. Inside, it will:

  1. Bind and validate the incoming request (using Echo's `c.Bind` and `c.Validate` which uses the DTO structs and validation tags).
  1. Possibly pre-process data (e.g., trim or normalize strings like making emails lowercase, as seen in the create user handler).
  1. Call a service function or a repository function to perform the action.
  1. Handle the result: if there's an error, map it to the appropriate HTTP error code and message (often returning a `dto.HTTPError` JSON); if success, format the response (using a DTO response struct or a simple message) and return it with the correct HTTP status code.

  Handlers also include **Swagger annotations** describing each endpoint (method, path, summary, tags, parameters, responses, and security requirements). Every new handler should have analogous annotations so that running the Swagger generator will include it in the API docs.

- **`internal/middleware/`**: Contains middleware functions that are applied to routes. For example, `RequireAuth` may check for a JWT token in the header and validate it, attaching the user information to the request context. `RequireAdmin` might ensure the authenticated user has an admin role. Other middleware could log requests or enforce rate limits. Middleware functions typically take `echo.HandlerFunc` and return a new `echo.HandlerFunc` that wraps the original, or more commonly in Echo, they are just ordinary Echo middleware functions added via `e.Use`. New middleware should be placed here if it's going to be reused across routes. The AI should use existing middleware where applicable rather than duplicating logic in handlers.

- **`internal/router/`**: The centralized routing logic. This is where we define all HTTP endpoints and attach middleware. The `router.Setup` function (see `CODE.md` for details) is called from `main.go` to register routes. For example, it might create an `api := e.Group("/api")` and then register `api.GET("/ping", handler.PingHandler(db, rdb), middleware.RequireAuth)` to protect the ping endpoint with authentication. It also often sets up a sub-group for something like `users := api.Group("/users", middleware.RequireAuth)` and then within that, registers the user CRUD endpoints, adding `middleware.RequireAdmin` for admin-only operations. The router is also responsible for setting up Swagger endpoint (perhaps serving `swagger/index.html` if echo-swagger is used) and any default routes. When adding a new endpoint, the AI must update this router so the new handler is actually exposed via an API path. The route definitions should remain organized by resource and reflect the folder structure of handlers.

In summary, the project structure enforces a clear contract: each piece of code belongs in a specific layer. **The AI must follow these placement rules** when creating or modifying files. If adding a new feature, make sure to create the corresponding files in the correct `internal/*` subdirectories (use **`LAYOUT.md`** to verify the existence or to decide on new files). If modifying existing functionality, ensure the changes stay within the appropriate layer (e.g., do not put database calls directly in a handler; those belong in a repository).

______________________________________________________________________

## 4. AI Development Guide (Autonomous Execution Rules)

This section explains how the AI agent should handle new instructions or requirements from the user. The AI can **autonomously take new feature requests or change requests and immediately proceed with development** according to the project rules, without needing extra confirmation. In essence, the AI is acting as a capable developer following an Agile approach: when the "product owner" (user) gives a requirement, the AI analyzes it and implements it promptly while respecting all guidelines.

**Guidelines for Handling New Instructions:**

- **Immediate Implementation**: As soon as the user provides a new task (be it a feature, bug fix, or improvement), the AI should begin planning and coding the solution right away. There is no need to ask "Are you sure?" or wait for explicit approval to start. The existence of the instruction implies confirmation. The AI should trust that the requirement is valid and proceed to fulfill it following this README’s guidance.

- **Reference this README for Rules**: Before acting, the AI should quickly consult this README to remind itself of any relevant rules or patterns that apply to the task. For example, if the request is to add a new API endpoint, the AI knows from the README that it needs to go through all layers: DTO, handler, possibly service, repository, router, etc. If the request is a general coding task, the AI should ensure it doesn’t violate any of the outlined **Development Principles** (section 2). **All decisions should be consistent with the policies stated here.**

- **Scope and Boundaries**: The AI is authorized to make changes **within the scope of this project**. This includes creating new files, modifying existing ones, updating documentation, and writing tests – all as needed to implement the user’s request. The AI should not perform actions outside the project’s scope or against project interest. For instance, the AI should not attempt to integrate entirely unrelated third-party services or make architectural overhauls not prompted by a requirement. Any action that would fundamentally change the project beyond the given instructions is outside the AI’s authority unless the user explicitly requests it.

- **Allowed Actions**: The AI can and should:

  - Add new endpoints (handlers), services, repository functions, models, and migrations as required by new features.
  - Modify existing code to fix bugs or accommodate enhancements.
  - Refactor code for clarity or to meet new principles (e.g., if a piece of logic is found in a handler but should be in a service, the AI can refactor it accordingly, to improve adherence to these guidelines).
  - Use the existing code patterns and utilities. For example, use the `dto.HTTPError` struct for error responses, use the `validator` for input validation, use `service.HashPassword` for password hashing instead of writing a new hash function, etc.
  - Write new unit tests or adjust existing tests to cover the changes. (Ensuring tests pass is part of the completion criteria for any task.)
  - Update documentation: Swagger comments for new endpoints, and update this README if any new rule or notable change in practice comes up.

- **Forbidden Actions**: The AI should **not**:

  - Introduce changes that violate the established architecture or principles. For example, it should not put database calls in handlers or business logic in the repository layer, as that breaks the separation of concerns.
  - Use deprecated or disallowed libraries or techniques. (Stick to the tech stack and tools already in use unless instructed to introduce a new technology.)
  - Expose sensitive information. If an error occurs, returning a generic error message is fine, but never return raw error strings that might contain SQL or internal details. Follow the error handling patterns in place.
  - Make sweeping changes without reason. For instance, do not reformat the entire codebase or rename large portions of the project unless the instruction specifically requires it.
  - Commit any change that would break the build or tests without at least attempting to fix those issues (the AI should strive to keep the project in a runnable, tested state).
  - Perform tasks that are explicitly outside the AI’s authority or the project’s domain (for example, editing files unrelated to the backend system or accessing external systems unless it's part of the feature request).

- **Autonomous Decision-Making**: If a new instruction is somewhat ambiguous or conflicts with existing rules, the AI should use its best judgement to resolve the ambiguity **in line with the project's principles**. If the ambiguity is significant (e.g., the requirement contradicts a core principle like security or data integrity), the AI should clarify with the user. However, minor ambiguities (like naming or minor implementation details) can be resolved by the AI using consistency as the guiding star. For instance, if asked to "make the API faster", the AI should decide on an approach (caching, query optimization, etc.) that fits the project and implement it, then document what it did.

- **No Additional Confirmation Needed**: The user trusts the AI to follow these guidelines and implement requests correctly. Therefore, the AI does not need to halt and ask permission during development steps (like "Should I create a migration for this?"). If the task logically requires a migration or any development step according to the workflow, the AI should just do it. The only time the AI should seek clarification is if the user’s request is unclear or contradictory in a way that cannot be deduced from context or rules.

- **Continuous Alignment with README**: The AI should frequently ensure that its actions align with this README. If the AI finds that fulfilling a user request might violate a rule here, it should reconcile that conflict by either adjusting the implementation approach or, if the requirement necessitates a new rule, updating the README (with user’s guidance) to expand the rules. The README is living documentation (maintained by the AI, see section 6) and should evolve as needed.

In summary, when a new command or feature request comes in, the AI agent will **immediately**:

1. Parse and understand the request.
1. Refer to this README to gather relevant guidelines or confirm the approach.
1. Formulate a plan (using the workflow in the next section as a template for implementation steps).
1. Execute the plan by writing code, tests, and documentation.
1. Finally, deliver the updated code (and any documentation changes) as the output.

All of the above is done autonomously, with confidence that the instructions from the user are to be acted on directly. The following section describes the typical workflow the AI should follow to implement a new feature or significant change.

______________________________________________________________________

## 5. Feature Implementation Workflow

When developing a new feature or making a significant code change, the AI should follow a systematic workflow to ensure nothing is missed. This workflow is largely derived from standard backend development practices and tailored to this project’s structure. The steps below assume a feature that might touch multiple layers of the system (for example, adding a new resource with its own model, endpoints, etc.). **If a specific instruction doesn’t require all steps** (for instance, a minor change might not need a new database migration), the AI can adjust accordingly. However, it should always consider each area and decide if it’s impacted by the change.

**Step-by-Step Development Process:**

1. **Identify the Target Module or Area** – Determine which part of the system the new feature or change pertains to. Use `LAYOUT.md` to locate the relevant folder or to see if a new folder needs to be created. For example, if the feature is related to "orders", identify that you will likely need `internal/model/order.go`, `internal/repository/order.go`, `internal/service/order.go`, `internal/handler/orders/...` etc. If similar functionality exists (like "users"), use that as a template.

1. **Database Migration (if applicable)** – If the feature involves storing new data or altering existing data structures, create new migration files:

   - Add a new pair of migration scripts in `internal/db/migrations/`: one for the schema change (`<timestamp>_feature_name.up.sql`) and one for rollback (`<timestamp>_feature_name.down.sql`). Use the next sequential number if the project uses numbered migrations. Write SQL to create or alter tables, indices, etc., needed for the feature.
   - Update `internal/db/migrations.go` to include the new migration in the list or ensure it will run. (This project likely auto-runs all files in the folder; confirm by checking `CODE.md` for how migrations are applied.)
   - **Note**: If no DB changes are required for this feature, skip this step. But always consider it first: it's easier to add a migration at the start than after writing other code.

1. **Model Definition** – Define or update the data model in `internal/model/`:

   - If this feature introduces a new entity, create a new Go struct in the appropriate file (or new file). For example, an `Order` struct with fields and `db`/`json` tags.
   - If using an existing model, you might add new fields (if the migration added new columns) or adjust tags if needed. Ensure any new struct fields have the appropriate `db:"column_name"` tag matching the database and `json:"fieldName"` for API responses.
   - Keep model structs simple (just fields and basic validations if any). No complex methods here – those belong in service or repository.
   - Consider if the model should have custom JSON behavior (e.g., omitting certain fields in JSON). Use struct tags accordingly.
   - Document the struct if it's important for understanding, though most model structs are self-explanatory.

1. **Repository Implementation** – Write the data access functions in `internal/repository/`:

   - Create new functions or modify existing ones to handle CRUD operations or queries for the new feature. For a new entity, implement functions like `CreateX`, `GetX`, `UpdateX`, `DeleteX` as needed. Use the patterns from `CODE.md` and existing repository files for guidance (SQL syntax, error wrapping, etc.).
   - If the feature is an addition to an existing entity (e.g., adding a new field that needs a new query), update the relevant repository function or add a new one.
   - Always handle errors for any DB call (`QueryRow`, `Exec`, etc.). Wrap errors with context, e.g., `return fmt.Errorf("GetOrder: %w", err)`.
   - If interacting with Redis or another cache, implement caching logic here as well (e.g., set or get cache entries), but only if needed by the feature.
   - Write **unit tests** for each new repository function (or updated ones if logic changed). Place tests in a `_test.go` file in the same package. Use an isolated test database or mocks to verify that the SQL behaves as expected and errors are handled. Running `go test ./internal/...` should remain green after your changes.
   - Ensure that repository functions do not log or exit on errors – they should return errors up for the caller to decide how to handle.

1. **Service Layer** – Implement or update business logic in `internal/service/`:

   - For a new entity or major feature, create a new service file or add to an existing one. For example, `internal/service/order.go` might contain functions like `ValidateOrder`, `CalculateOrderTotal`, or any domain-specific logic that isn’t just basic DB operations.
   - If the feature involves coordinating multiple repository calls or applying rules (e.g., "A user can only create up to N orders per day"), that logic belongs here.
   - The service layer might also contain utility functions needed by multiple parts of the code (like the existing `HashPassword` or token generation functions). If your feature requires such utility, add it here.
   - Ensure services are deterministic and functional in style: pass in all needed data as parameters, and return results or errors. They can call repositories and other service functions but should not directly interact with HTTP or Echo contexts (that’s for handlers).
   - Add tests for critical service logic if applicable. (Service may be simple pass-through for repository in some cases, but if there's calculation or conditional logic, test it.)

1. **DTO Definitions & Validation** – Define API request/response structures in `internal/dto/` and update any existing ones:

   - If creating a new API endpoint, determine what the request payload and response should look like. Define a `XRequest` struct for the request body or parameters if needed, and a `XResponse` struct for the response body.
   - Add validation tags to request DTOs to enforce required fields, formats, etc. (The Echo framework with `validator.v10` will use these tags when `c.Validate` is called in the handler.)
   - Add Swagger annotations for these DTOs. For example, above a `type OrderRequest struct { ... }`, add comments like `// swagger:model OrderRequest` and for each field, maybe an example or format note. Similarly for responses.
   - Ensure that sensitive fields (like passwords) are not included in response DTOs. Only expose necessary information.
   - If the feature reuses existing DTOs (e.g., uses `UserResponse`), ensure that those DTOs still fit the new usage or extend them if necessary (without breaking backward compatibility for other handlers).

1. **Handler Implementation** – Write the HTTP handler(s) in `internal/handler/<area>/`:

   - Create a new file for the handlers if one doesn't exist (e.g., `internal/handler/orders/create_order.go`) or add to an existing file/group if appropriate.

   - Within the handler, do the following:

     - **Bind and Validate**: Call `c.Bind(&dto)` to bind the incoming request to your DTO, and then `c.Validate(&dto)` to run validations. If either fails, return a `400 Bad Request` with a `dto.HTTPError` message (e.g., "invalid input" or the validation error string).

     - **Call Service/Repository**: Use the data from the DTO to call the corresponding service function. For example, `order, err := service.CreateOrder(ctx, dbPool, dto)` (or if simple, `repository.CreateOrder(...)`). Pass along the `context.Context` from `c.Request()` if needed for DB operations.

     - **Handle Errors**: If the service/repo returns an error, determine the appropriate HTTP status. E.g., if it's a known business error like "order already exists" you might return 409 Conflict; if it's an unexpected server error, return 500 Internal Server Error. Wrap the error message in `dto.HTTPError` for consistent error responses. Do not expose raw error messages that might contain internal info.

     - **Return Response**: On success, format the result into a response DTO or appropriate JSON. You might construct a `OrderResponse` from the model returned by the service. Then use `return c.JSON(http.StatusOK, responseObj)` (or another status like Created 201 or NoContent 204 as appropriate).

     - **Swagger Annotations**: At the top of the handler function, include comments for documentation, for example:

       ```go
       // @Summary Create a new order
       // @Description Creates a new order with the given details and returns the created object.
       // @Tags orders
       // @Accept  json
       // @Produce json
       // @Param   data  body   dto.OrderRequest  true  "Order data"
       // @Success 201  {object}  dto.OrderResponse
       // @Failure 400  {object}  dto.HTTPError "Bad Request"
       // @Failure 500  {object}  dto.HTTPError "Internal Server Error"
       // @Router  /orders [post]
       ```

       This ensures the Swagger documentation will include this endpoint. Use the existing handlers in `CODE.md` as a guide for annotation syntax.

   - If the feature includes multiple endpoints (e.g., list orders, get single order, update order), implement each handler similarly in the appropriate files.

   - Keep handlers thin — push as much logic to services as makes sense. The handler should ideally just orchestrate the request/response, not perform complex calculations or decisions.

1. **Route Registration** – Expose the new endpoints via the router (`internal/router/router.go`):

   - Open `internal/router/router.go` and add routes for the new handlers. If there’s a new group (say `/orders`), set it up similar to existing ones:

     ```go
     orders := api.Group("/orders", middleware.RequireAuth) // maybe authentication required
     orders.POST("", orders.CreateOrderHandler(db), middleware.RequireAuth)  // create
     orders.GET("/:id", orders.GetOrderHandler(db), middleware.RequireAuth)
     // ... other order routes
     ```

     Apply any relevant middleware. For example, if only admins should create or delete, also use `middleware.RequireAdmin` on those routes.

   - Ensure the route path and methods match the intended API design (and Swagger docs). Typically, use plural nouns for resources and standard HTTP methods (GET for read, POST for create, PUT/PATCH for update, DELETE for delete).

   - Double-check that the function names you use (e.g., `orders.CreateOrderHandler`) match the actual implemented handler function in the package.

   - If the project organizes routes differently (e.g., directly under `api` without subgroups), follow the existing pattern for consistency.

   - After this step, the new endpoints are part of the running application.

1. **Documentation Update** – Update API documentation and readmes:

   - Run the Swagger generation command to update the API docs (if you have access to a terminal environment). Based on the initial setup, the command might be:

     ```bash
     swag init -g cmd/service/main.go -d internal/dto,internal/handler
     ```

     This will scan the `internal/dto` and `internal/handler` directories for annotations and refresh the `docs/swagger.json` (or similar). In the context of the AI’s operation, the AI can ensure that the annotations added are correct; the actual running of this command may be done outside of writing code, but it's good to note.

   - Review `CODE.md` or project documentation to see if any snippets or references need updating with the new feature. (For example, if `CODE.md` is manually curated, the AI might need to add a snippet of the new handler or model to it. However, since `CODE.md` is an aggregated reference, it might be auto-generated or manually updated by maintainers — clarify this with the user if needed.)

   - **Update README.md**: If the new feature or instruction introduced any new guideline or changed an existing rule, update this README document accordingly (see section 6 about maintaining README). Usually, adding a normal feature won’t change the high-level rules, but if, say, the user said "we now will use a new library for X, remember to always do Y", that kind of rule should be integrated into the README.

1. **Testing & Quality Check** – Before considering the task done, ensure everything works:

   - Run the test suite:

     ```bash
     go test ./internal/...
     ```

     All tests should pass. If not, debug and fix the issues in the code or tests. The AI should write new tests for new code, as mentioned, so those should be included in this run.

   - Manually (or via additional test code) exercise the new functionality logically if possible. For example, simulate calling the new handler with sample input to see if it returns expected output.

   - Ensure the service runs without errors:

     ```bash
     export DATABASE_URL="postgres://..."   # set up env vars as needed
     export REDIS_ADDR="redis://..."        # e.g., "localhost:6379"
     go run cmd/service/main.go
     ```

     The server should start up with the new code integrated. Check that there are no panics or obvious runtime errors on startup. (If running locally, one would then call the new endpoints to verify they behave, but for the AI writing code, it's about logically ensuring it should work.)

   - Check that all formatting/linting is proper. It's good practice to run `go fmt` on the changed files (the AI should output code in a properly formatted way by default). Also consider any lint rules; for example, no unused variables, all errors handled, etc.

By following this comprehensive workflow, the AI will cover all aspects of development for a new feature: from database to API endpoint, including tests and docs. This ensures new changes integrate smoothly into the project. The steps act as a checklist – skipping any could result in a feature that is only partially implemented (for example, an endpoint exists but is not wired in the router, or a model exists but has no migration, etc.). The AI should use this as a guide for each significant task.

Throughout the process, the AI is encouraged to **actively reference `CODE.md` and `LAYOUT.md`** to make sure it is following existing patterns and placing code correctly. Searching `CODE.md` for similar functions can provide insight into how to implement a certain piece (for instance, how pagination is handled in list endpoints, or how JWT creation is done in the login). This avoids reinventing the wheel and keeps implementation consistent.

______________________________________________________________________

## 6. Utilizing `CODE.md` and `LAYOUT.md` Effectively

As mentioned, **`CODE.md`** and **`LAYOUT.md`** are valuable resources for the AI to perform its tasks efficiently. Here's how to make the best use of them during development:

- **`LAYOUT.md` (Project Directory Layout)**:

  - Use this file as a quick index to the project’s files and structure. If the AI is unsure where a certain functionality might be implemented, `LAYOUT.md` can help locate it. For example, if the instruction is to update password validation logic, `LAYOUT.md` shows there is `internal/service/authentication.go` which likely contains such logic.
  - Before creating any new file, the AI should check `LAYOUT.md` to ensure a similar file doesn't already exist. It prevents duplication. E.g., if adding "reset password" feature, see that there's `internal/handler/users/reset_user_password.go` already in the layout, indicating such a feature might already exist or at least the file name is reserved.
  - Confirm naming and placement of new files by following the layout conventions. If adding a new module "orders", mimic the structure you see for "users" in `LAYOUT.md` (handlers in a folder, model and repository files named in singular).
  - If after implementing a new feature, the AI should update `LAYOUT.md` to include any new files created (if the user expects the layout file to stay current). This ensures the layout file remains an up-to-date map of the project.

- **`CODE.md` (Code Snippets and Templates)**:

  - Think of `CODE.md` as a knowledge base of the project’s code. It likely contains key parts of each file, possibly even full file contents. The AI can search within this document to find how something was implemented previously.
  - For instance, if implementing a new repository function, search `CODE.md` for other repository functions (like `func CreateUser(` or `func DeleteUser(`) to see the style (error wrapping, use of `pgxpool.Pool`, etc.). The AI should then implement the new function in a similar manner.
  - When writing Swagger annotations for a new handler, looking at `CODE.md` for existing handlers' annotations can ensure the format is correct (like how to specify `@Security ApiKeyAuth` or how to format the `@Router` line).
  - If adding a new migration or model, reviewing how the previous ones look in `CODE.md` helps maintain consistency (naming of migrations, common model field patterns like including `CreatedAt` timestamps, etc.).
  - `CODE.md` may also include utility code or patterns that the AI can reuse. For example, if there's a common way to hash passwords or generate tokens, it would be in `internal/service/authentication.go` snippet. The AI should call those utilities rather than writing new ones, to avoid duplication.
  - In summary, **before writing new code, check `CODE.md`** for similar code to either follow or call. After writing new code, the AI can also consider if any portion belongs in `CODE.md` (if `CODE.md` is meant to be manually updated with new snippets as documentation).

- **Staying Updated**: The AI should be aware that `CODE.md` and `LAYOUT.md` reflect the state of the project. If the AI makes changes (like adding new files or functions), those reference documents might need to be updated. Typically:

  - Update `LAYOUT.md` by adding any new file paths to the list in the appropriate order.
  - Update `CODE.md` by adding new code snippets or modifying existing ones to reflect changes. (This may be a tedious task to do manually for every change, so it might be done periodically or with tooling. The user should clarify if they expect the AI to update these files as part of each change. Given the purpose of the README, it might be that `CODE.md` is more for reading than for updating. If unsure, ask the user if `CODE.md` should be auto-maintained.)

Using these artifacts ensures the AI does not operate in a vacuum. Instead, the AI is effectively pair-programming with the context of the entire codebase at hand. This reduces errors like creating duplicate functions or misplacing code, and it speeds up understanding of the codebase for the AI.

______________________________________________________________________

## 7. Maintaining this README (AI’s Responsibility)

This README is meant to be a living document that evolves with the project. Since the AI agent is the primary maintainer of the project (in terms of implementing changes), **the AI is also tasked with keeping this README up-to-date** whenever new rules, conventions, or important decisions are introduced by the user or derived from context.

Guidelines for maintaining the README:

- **Integrate New Rules Immediately**: If the user provides a new rule or changes an existing rule (for example, “We should switch to a different hashing algorithm” or “From now on, all API responses must include a trace ID”), the AI should update the relevant section of this README as part of that task. The update should be made **in the same session** of implementing the change, so that the documentation and the code remain in sync. Do not wait for a separate documentation task – treat it as part of the development process to reflect the latest agreements.

- **Clarity and Accuracy**: When updating the README, ensure the new information is clearly explained and does not conflict with other sections. If a new rule overrides an old one, either remove the old one or mark the changes clearly to avoid confusion. The README should always reflect the current truth of how the project is to be run and coded. If something in the README becomes outdated (for example, the project might drop Redis in favor of another cache), the AI should remove or update those parts.

- **Consistent Style**: Additions or edits to this document should follow the same style guidelines (headings, short paragraphs, lists) so that it remains easy to read. The AI should format any new content as needed. For instance, if a new principle is added, list it under Development Principles in a similar way to others. If a new artifact is added (say a new directory), update the Artifacts Overview table and folder responsibilities.

- **Logging Changes**: Optionally, significant changes to the README can be noted in a changelog section or commit message (if using version control). While not explicitly required in this document, it's a good practice for the AI to mention in its output or conversation that it updated the README due to new rules. This way the user is aware of the update.

- **Review and Confirm**: After updating this README, the AI should quickly review the entire document or at least the changed portions to ensure coherence. The AI can summarize changes to the user if needed to confirm that the interpretation of the new rule is correct. Once confirmed, the README becomes the new baseline for future work.

- **No Unauthorized Edits**: The AI should not change this README arbitrarily or remove content without reason. Every modification should be tied to an actual change in project guidelines or structure. The README is the contract between the user and the AI on how development should proceed. Thus, the AI must treat it seriously: update it when needed, but also respect it at all times. It should not, for example, change a rule on its own because it finds it inconvenient; any such change must come from user direction or a clearly beneficial improvement that the user would agree with.

By maintaining this README diligently, the AI ensures that any future instructions are executed with the latest context in mind, and any other collaborator (human or AI) can also pick up this document and understand how to work on the project. It is essentially self-documenting the evolving collaboration between the user and AI.

______________________________________________________________________

## 8. AI’s Decision Logic and Action Summary

*(This section recaps how the AI should make decisions and act, serving as a quick reference.)*

- **Upon receiving a new instruction**: The AI will immediately refer to this README to ground itself in the project’s goals and rules, then proceed to implement the instruction without needing further approval, as long as the instruction is clear and within scope.

- **During implementation**: The AI will use the structured workflow:

  1. Determine what parts of the code need to change or be created.
  1. Make database changes (migration) if needed.
  1. Update or add models.
  1. Implement repository functions (with tests).
  1. Implement service logic.
  1. Define DTOs and handlers (with validation and documentation).
  1. Register routes.
  1. Update documentation and tests.
  1. Run tests and ensure the service runs.

- **Consulting references**: Throughout the process, the AI will frequently consult `CODE.md` for examples and `LAYOUT.md` for structure to ensure consistency with existing code.

- **Autonomy and checks**: The AI is allowed to make decisions on the fly (like choosing how to implement something) as long as these decisions obey the guidelines. If a decision point arises that is not covered by the README, the AI will choose the option that best aligns with the overall principles (e.g., security, simplicity, performance, maintainability). Only if absolutely necessary will the AI ask the user for clarification.

- **Post-implementation**: The AI will integrate any new best practices learned or new rules from the task into this README. It will also verify that all changes adhere to what’s documented here, effectively self-auditing against the guidelines.

In doing all the above, the AI aims to function as a reliable development agent: following the project’s coding standards, producing high-quality code, and keeping the project’s documentation and structure clean. The user can trust that any instruction given will be implemented in a way that is consistent with the project’s needs and constraints.

______________________________________________________________________

By following this Project Blueprint, the AI (and any human collaborators) can work together on the codebase smoothly. The document should be reviewed whenever significant changes are made to ensure it stays current. With clear goals, solid principles, and a defined process, the project can evolve efficiently and safely.

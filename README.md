# 📝 NoteApp — REST API for Managing Notes

**NoteApp** is a production-ready Go (Golang) backend for managing personal notes with JWT authentication, PostgreSQL, clean architecture, and full DevOps setup.

Built with:
- **Gin** (HTTP framework)
- **PostgreSQL** (database)
- **JWT + Refresh tokens** (authentication)
- **Clean Architecture** (separation of layers)
- **Structured logging** (`zap`)
- **Docker & Makefile** (local dev)
- **GitHub Actions CI/CD**
- **Migrations** (`golang-migrate`)
- **Unit & integration tests**

---

## 🚀 Features

- ✅ User registration, login, logout, password change
- ✅ CRUD for notes with filtering (`done`/`not done`) and pagination
- ✅ JWT authentication with secure refresh token rotation
- ✅ Input validation with custom error messages
- ✅ Structured logging (JSON/console) with levels
- ✅ Graceful shutdown
- ✅ Automated testing with isolated test DB
- ✅ CI/CD pipeline (GitHub Actions)
- ✅ Docker-based local environment
- ✅ Configurable via YAML
- ✅ Migrations for schema versioning

---

## 🛠️ Tech Stack

| Layer         | Technology                     |
|--------------|--------------------------------|
| Language     | Go 1.24+                       |
| Framework    | [Gin](https://gin-gonic.com/)  |
| Database     | PostgreSQL                     |
| Auth         | JWT (HS256)                    |
| Logging      | `zap` (Uber)                   |
| Validation   | `validator/v10`                |
| Password     | `bcrypt`                       |
| Config       | YAML                           |
| CI           | GitHub Actions                 |
| Migrations   | `golang-migrate/migrate`       |

---

## 📦 Prerequisites

- [Go 1.24+](https://go.dev/dl/)
- [PostgreSQL](https://www.postgresql.org/download/)
- [Docker](https://www.docker.com/) (optional, for dev/test DB)
- [`migrate`](https://github.com/golang-migrate/migrate) CLI
- `make` (for convenience)

---

## ⚙️ Configuration

Copy and edit the config:

```bash
cp configs/config.yaml
```

**Edit configs/config.yaml with your DB and auth settings.**

**Example:**
```yaml
server:
  port: ":8080"
  read_timeout: 10s
  write_timeout: 10s
  shutdown_timeout: 15s

db:
  host: "localhost"
  port: 5432
  user: "root"
  password: "px40w1!oEn"
  name: "notesApp"
  sslmode: "disable"

auth:
  access_token_ttl: "15m"
  refresh_token_ttl: "720h" # 30 days

logger:
  environment: "development"
  level: "debug"
  encoding: "console"
  output_paths: ["stdout"]
  error_output_paths: ["stderr"]
```

**Make .env file**
**Example**
```.env
CONFIG_PATH=example
CONFIG_NAME=example

CONTAINER_NAME=example

DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=example
DB_USER=example
DB_PASSWORD=example
DB_SSLMODE=disable

AUTH_JTT_SECRET=example
```

---

## 🐳 Development Setup (Docker + Makefile)

Start the app with **Dockerized PostgreSQL** and **auto-migrations**:

```bash
make start
```

This will:

1. Start PostgreSQL in Docker
2. Create the database
3. Run migrations
4. Launch the server

Stop everything:
```bash
make stop
```

---

## 🧪 Testing
Run full test suite (with isolated test DB):

```bash
make test-start
```

This:

1. Starts a separate test DB in Docker
2. Creates test_notes_app
3. Runs migrations
4. Generates mocks
5. Runs go test
6. Cleans up

---


## 🌐 API Endpoints
**Auth**
| METHOD        | ENDPOINT                       | DESCRIPTION        |
|-------------- |--------------------------------|--------------------|
|POST           |`/api/auth/sign-up`             |Register            |
|POST           |`/api/auth/sign-in`             |Login               |
|GET            |`/api/auth/logout`              |Logout              |
|GET            |`/api/auth/refresh`             |Refresh access token|

**Profile**
| Method | Endpoint               | Description         |
|--------|------------------------|---------------------|
| GET    | `/api/profile`         | Get profile         |
| PUT    | `/api/profile`         | Update profile      |
| PUT    | `/api/profile/pass`    | Change password     |
| DELETE | `/api/profile`         | Delete account      |


**Notes**
| Method | Endpoint                  | Description                          |
|--------|---------------------------|--------------------------------------|
| POST   | `/api/notes`              | Create note                          |
| GET    | `/api/notes`              | List notes (pagination, "done")      |
| GET    | `/api/notes/:note_id`     | Get note                             |
| PUT    | `/api/notes/:note_id`     | Update note                          |
| DELETE | `/api/notes/:note_id`     | Delete note                          |

---

## 🧰 Makefile Commands

| Command           | Description                                      |
|-------------------|--------------------------------------------------|
| `make start`      | Start app + DB + migrations + server             |
| `make stop`       | Stop DB and clean up                             |
| `make run`        | Run server only (DB must be up)                  |
| `make migrate-up` | Apply DB migrations                              |
| `make migrate-down` | Roll back migrations                           |
| `make test-start` | Full test cycle (DB, migrate, test, cleanup)     |
| `make mock`       | Generate mocks for interfaces                    |

---

🔄 CI/CD (GitHub Actions)
The `.github/workflows/ci.yml` runs on every `push`/`pull_request` to `main`/`master`:

✅ Checkout
✅ Setup Go
✅ Download deps
✅ Run tests
✅ Check formatting (gofmt)

---

## 🧪 Example Requests

**Sign Up**
```bash
curl -X POST http://localhost:8080/api/auth/sign-up \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "email": "alice@example.com",
    "password": "secure123"
  }'
```

**Sign In**
```bash
curl -X POST http://localhost:8080/api/auth/sign-in \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "password": "secure123"
  }'
```

---

### 📄 License
MIT License – see LICENSE for details.

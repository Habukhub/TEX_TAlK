# TEX Backend

Go backend scaffold สำหรับ `TEX/backend` ที่วางโครงแบบ Clean Architecture ด้วย Fiber, GORM และ Atlas migrations โดยในโปรเจคนี้ใช้ environment แค่ `local` และ `prod`

---

## สารบัญ

1. [Quick Start](#quick-start)
2. [Project Architecture](#project-architecture)
3. [Technology Stack](#technology-stack)
4. [Project Structure](#project-structure)
5. [Directory Responsibilities](#directory-responsibilities)
6. [Feature Development Workflow](#feature-development-workflow)
7. [Current Scaffold Scope](#current-scaffold-scope)
8. [Configuration Reference](#configuration-reference)
9. [Database & Migration Layout](#database--migration-layout)
10. [Logging Layout](#logging-layout)

---

## Quick Start

### Prerequisites

```bash
go mod tidy
curl -sSf https://atlasgo.sh | sh
```

### Run locally

```bash
docker compose up -d postgres
go run ./src/cmd/api
```

หรือใช้ Makefile:

```bash
# macOS / Linux
make tidy
make db-up
make dev
```

```bash
# Windows
make tidy
make db-up-win
make dev-win
```

Default endpoints:

```bash
GET /healthz
```

---

## Project Architecture

### Clean Architecture Layers

โปรเจคนี้แยก responsibility เป็นชั้น ๆ เพื่อให้เพิ่ม feature, test, และเปลี่ยน implementation ได้ง่าย

```text
HTTP Request
    |
    v
handler/                     รับ request, parse input, คืน response
    |
    v
usecase/                     รวม business rules และ flow ของงาน
    |
    v
infrastructure/repository/   เข้าถึง database หรือ external systems
    |
    v
domain/                      entities และข้อมูลหลักของระบบ
```

### Layer Mapping

| Layer | Folder | หน้าที่ |
| --- | --- | --- |
| HTTP / Delivery | `src/internal/handler` | รับ request, map input/output, เรียก usecase |
| Application / Business | `src/internal/usecase` | รวม business logic และกำหนด contract สำหรับ data access |
| Data Access | `src/internal/infrastructure/repository` | implementation ที่คุยกับ GORM/PostgreSQL |
| Domain | `src/internal/domain` | นิยาม entity/model หลักของระบบ |
| Cross-cutting | `configs`, `logger`, `middleware`, `util`, `constant` | config, logging, helper, constants |

### Design Intent

- `domain/` เก็บ model กลางของระบบ
- `dto/` ใช้สำหรับ request/response object
- `usecase/` เป็นจุดรวม business logic
- `repository/` รับผิดชอบ data access
- `router/` ทำหน้าที่ประกาศ route และผูก route กับ handler

### Current Implementation Pattern

ตอนนี้ scaffold คงไว้แค่ flow กลางของระบบ:

```text
router -> handler/healthcheck
```

ส่วน `domain`, `usecase`, `repository` และ module feature อื่น ๆ ตั้งใจเว้นให้เพิ่มเองตาม schema และ business flow ของ TEX

---

## Technology Stack

| Layer | Technology | ใช้สำหรับ |
| --- | --- | --- |
| Backend | Go | ภาษาหลักของ service |
| HTTP Framework | Fiber v3 | routing, middleware, request/response |
| ORM | GORM | เชื่อมต่อและ query PostgreSQL |
| Database | PostgreSQL 18 | relational database สำหรับ local stack |
| Migration | Atlas | schema loader และ SQL migrations |
| Container | Docker Compose | รัน local infrastructure |
| Logging | `log/slog` JSON logger | structured logs |
| Log Shipping Layout | Fluent Bit / OpenSearch | โครงสร้างสำหรับต่อยอด centralized logging |

---

## Project Structure

```text
TEX/backend/
├── src/
│   ├── atlas/
│   │   └── loader/                    # Atlas schema loader จาก GORM models
│   ├── cmd/
│   │   └── api/main.go                # entry point ของ API server
│   └── internal/
│       ├── configs/                   # โหลด config จาก file/env
│       ├── constant/                  # constants และ path definitions
│       ├── di/                        # reserved สำหรับ dependency injection container
│       ├── domain/                    # GORM entities / business models (add when schema is defined)
│       ├── dto/                       # request/response DTOs
│       ├── handler/                   # HTTP handlers
│       ├── infrastructure/
│       │   ├── database/              # PostgreSQL connection setup
│       │   ├── middleware/            # HTTP middleware implementations
│       │   ├── repository/            # data access implementations (add when schema is defined)
│       │   └── router/                # route registration
│       ├── logger/                    # structured logger + trace helpers
│       ├── middleware/                # business middleware (add when needed)
│       ├── testutils/                 # test mocks/helpers (add when needed)
│       ├── usecase/                   # business logic per module (add when needed)
│       └── util/                      # helper utilities
├── configuration/                     # local config files และ template credentials
├── docs/                              # docs/swagger placeholder
├── migrations/
│   ├── local/
│   └── prod/
├── opensearch/
│   └── fluent-bit/                    # local log shipping config
├── docker-compose.yml                 # local postgres + logging stack
├── atlas.hcl                          # Atlas environments and migration config
├── Makefile                           # common developer commands
├── VERSION                            # application version
└── go.mod / go.sum                    # Go module dependencies
```

---

## Directory Responsibilities

### Root Level

| Path | ใช้ทำอะไร |
| --- | --- |
| `go.mod` / `go.sum` | ระบุ Go module และ dependencies |
| `Makefile` | รวมคำสั่งพื้นฐาน เช่น `tidy`, `db-up`, `dev`, `migrate-*` |
| `docker-compose.yml` | รัน local PostgreSQL และ logging-related services |
| `atlas.hcl` | บอก Atlas ว่า schema source มาจาก GORM loader และ migration แต่ละ env อยู่ที่ไหน |
| `configuration/` | เก็บค่าตั้งต้นของ application สำหรับ local |
| `migrations/` | เก็บ SQL migrations แยกตาม `local` และ `prod` |
| `opensearch/fluent-bit/` | เก็บ config สำหรับ log collector/local log pipeline |
| `docs/` | ที่วาง swagger/generated docs ในอนาคต |

### `src/cmd`

| Path | ใช้ทำอะไร |
| --- | --- |
| `src/cmd/api/main.go` | boot application, load config, init logger, connect DB, register routes, start Fiber server |

### `src/atlas`

| Path | ใช้ทำอะไร |
| --- | --- |
| `src/atlas/loader` | โหลด GORM models แล้วแปลงเป็น schema input สำหรับ Atlas diff/apply |

### `src/internal/configs`

| Path | ใช้ทำอะไร |
| --- | --- |
| `configs/` | โหลดค่าจาก `configuration/config.json` และ environment variables พร้อม default values |

### `src/internal/constant`

| Path | ใช้ทำอะไร |
| --- | --- |
| `constant/` | รวม API paths, config filenames, timezone constants และชื่อ env บางตัว |

### `src/internal/domain`

| Path | ใช้ทำอะไร |
| --- | --- |
| `domain/` | สำหรับนิยาม entity/model ของระบบเมื่อ schema ชัดเจนแล้ว |

### `src/internal/dto`

| Path | ใช้ทำอะไร |
| --- | --- |
| `dto/` | นิยาม request/response object ที่ไม่ควรผูกตรงกับ entity |

### `src/internal/handler`

| Path | ใช้ทำอะไร |
| --- | --- |
| `handler/healthcheck` | health check endpoint |
| `handler/handlers.go` | รวม handler กลางที่ router ใช้งาน |

### `src/internal/usecase`

| Path | ใช้ทำอะไร |
| --- | --- |
| `usecase/` | สำหรับ business logic ของแต่ละ module เมื่อเริ่ม implement จริง |

### `src/internal/infrastructure`

| Path | ใช้ทำอะไร |
| --- | --- |
| `database/` | สร้าง connection ไป PostgreSQL และตั้งค่า connection pool |
| `repository/` | สำหรับ implementation ของ data access เมื่อ schema พร้อม |
| `router/` | ประกาศ route กลางของระบบ เช่น `/healthz` |
| `middleware/` | access log middleware และ HTTP cross-cutting concerns |

### `src/internal/logger`

| Path | ใช้ทำอะไร |
| --- | --- |
| `logger/` | สร้าง JSON logger และ helper สำหรับ trace id/context |

### `src/internal/middleware`

| Path | ใช้ทำอะไร |
| --- | --- |
| `middleware/` | สำหรับ middleware เชิง business เมื่อมี requirement เพิ่ม |

### `src/internal/testutils`

| Path | ใช้ทำอะไร |
| --- | --- |
| `testutils/` | ที่รวม mock, fixture, helper สำหรับ unit/integration tests ในอนาคต |

### `src/internal/util`

| Path | ใช้ทำอะไร |
| --- | --- |
| `util/` | utility functions กลางที่ reusable ได้หลาย package |

---

## Feature Development Workflow

### Step-by-Step Implementation

เวลาจะเพิ่ม API ใหม่ ให้ไล่ทำตามลำดับนี้:

1. **Define Feature Scope**
   ระบุก่อนว่า endpoint นี้ทำอะไร ใช้ method อะไร รับ request แบบไหน และคืน response แบบไหน

2. **Domain (ถ้ามี table/model ใหม่)**
   สร้าง GORM struct ใน `src/internal/domain/`
   ใช้ขั้นตอนนี้เมื่อ feature นั้นมี schema ของตัวเองจริง ๆ

3. **Migration**
   ถ้ามี model ใหม่ ให้ลงทะเบียน model ใน `src/atlas/loader/main.go`
   จากนั้นรัน:

   ```bash
   make migrate-diff name=create_feature_tables
   make migrate-apply
   ```

4. **DTO**
   สร้าง request/response struct ใน `src/internal/dto/`
   เช่น create request, update request, list response, detail response

5. **Repository**
   ถ้า feature ต้องคุยกับ database ให้สร้าง implementation ใน `src/internal/infrastructure/repository/`
   จุดนี้ควรรับผิดชอบแค่ query/insert/update/delete และไม่ใส่ business logic หลัก

6. **UseCase**
   สร้าง business logic ใน `src/internal/usecase/`
   usecase ควรเป็นตัว orchestrate flow ของงาน เช่น validate rule, เรียก repository, transform response

7. **Handler**
   สร้าง handler ใน `src/internal/handler/`
   หน้าที่คือ parse request, bind dto, เรียก usecase, และคืน HTTP response

8. **Register Handler**
   เพิ่ม handler ตัวใหม่เข้าไปใน `src/internal/handler/handlers.go`
   เพื่อให้ router ดึงไปใช้จากจุดรวมกลางได้

9. **Router**
   ลงทะเบียน route ใน `src/internal/infrastructure/router/router.go`
   เช่น `GET /api/v1/example`, `POST /api/v1/example`

10. **Middleware (ถ้าจำเป็น)**
    ถ้า endpoint ต้องมี auth, logging เฉพาะ, validation เพิ่ม หรือ request context บางอย่าง ให้เพิ่มใน `src/internal/middleware/` หรือ `src/internal/infrastructure/middleware/`

11. **Testing**
    ถ้ามี logic สำคัญ ให้เพิ่ม test helper ใน `src/internal/testutils/`
    แล้วเขียน unit test ให้กับ usecase หรือส่วนที่มี business rule

### Typical API Flow

ตัวอย่างเวลาจะเพิ่ม API ใหม่ เช่น `POST /api/v1/example`

```text
1. dto/               สร้าง ExampleCreateRequest และ ExampleResponse
2. domain/            สร้าง Example model ถ้ามี schema ใหม่
3. repository/        สร้าง method CreateExample
4. usecase/           สร้าง CreateExampleUsecase
5. handler/           สร้าง CreateExample handler
6. handlers.go        register handler
7. router.go          เพิ่ม POST /api/v1/example
8. migrations/        สร้างและ apply migration ถ้ามี table ใหม่
```

### Rule of Thumb

- ถ้าเป็นเรื่องโครงสร้างข้อมูล ให้เริ่มที่ `domain` และ `migrations`
- ถ้าเป็นเรื่องรูปแบบ input/output ให้เริ่มที่ `dto`
- ถ้าเป็นกติกาธุรกิจ ให้ลงที่ `usecase`
- ถ้าเป็นการคุยกับ DB ให้ลงที่ `repository`
- ถ้าเป็นการรับ/ตอบ HTTP ให้ลงที่ `handler` และ `router`

---

## Current Scaffold Scope

สิ่งที่มี implementation จริงในตอนนี้:

- boot server ผ่าน Fiber
- config loading จาก file + env
- JSON access logging
- PostgreSQL connection setup
- Atlas schema loader placeholder
- `GET /healthz`

สิ่งที่มีแค่ structure placeholder ตอนนี้:

- `di/`
- `domain/`, `usecase/`, `repository/`, `testutils/`
- migrations จริงของแต่ละ environment

จุดประสงค์คือให้ทีมเริ่ม feature ใหม่บนโครงนี้ได้เลย โดยไม่ต้องย้อนมาจัด tree ใหม่ภายหลัง

---

## Configuration Reference

config อ่านจาก `configuration/config.json` และ environment variables โดย env จะ override file values

### App & Server

| Variable | Default | Description |
| --- | --- | --- |
| `APP_ENV` | `local` | environment name (`local` หรือ `prod`) |
| `PORT` | `8080` | port ของ API server |
| `SERVICE_NAME` | `tex-backend` | service name สำหรับ logging |
| `APP_TIMEZONE` | `Asia/Bangkok` | timezone หลักของ application |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:4200,http://localhost:3000` | allowed origins สำหรับ CORS |

### Database

| Variable | Default | Description |
| --- | --- | --- |
| `DB_DRIVER` | `postgres` | database driver |
| `DB_HOST` | `localhost` | database host |
| `DB_PORT` | `6632` | database port |
| `DB_USER` | `admin` | database username |
| `DB_PASSWORD` | `password` | database password |
| `DB_NAME` | `tex_db` | database name |
| `DB_SSLMODE` | `disable` | postgres sslmode |
| `DB_RETRY_BACKOFF` | `5s` | reserved สำหรับ retry/backoff policy |

---

## Database & Migration Layout

- `atlas.hcl` ใช้ `src/atlas/loader` เป็น schema source
- migrations แยกเป็น `local` และ `prod`
- ตอนนี้โฟลเดอร์ migration ยังว่าง intentionally เพื่อรอ schema จริงของ TEX

คำสั่งพื้นฐาน:

```bash
make migrate-diff name=create_initial_tables
make migrate-apply
make migrate-status
```

---

## Logging Layout

โครง logging ถูกเตรียมไว้ 2 ชั้น:

- ใน code ตอนนี้ใช้ JSON logger ผ่าน `log/slog`
- ใน infra มี `opensearch/fluent-bit` และ service ใน `docker-compose.yml` สำหรับต่อยอด centralized logging

ปัจจุบัน app ยังไม่ได้ผูก integration เชิงลึกกับ OpenSearch โดยตรง แต่โครงสร้างไฟล์และ container layout ถูกเตรียมไว้แล้ว

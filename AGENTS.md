# PROJECT KNOWLEDGE BASE

**Generated:** 2026-05-11
**Commit:** Latest
**Branch:** refactor/v2

## OVERVIEW
Go backend + embedded static UI for password/account management and note taking. SQLite storage, Gorilla mux HTTP API, vanilla JS frontend served via embed.FS.

## STRUCTURE
```
.
├── cmd/server/           # Server entry point
│   ├── main.go
│   ├── static/           # Embedded frontend (HTML/JS/CSS)
│   │   ├── index.html
│   │   ├── css/style.css
│   │   └── js/app.js
│   └── migrations/       # Database migrations (auto-run on startup)
├── internal/
│   ├── config/           # Configuration management
│   ├── database/         # Database connection & migrations
│   ├── handler/          # HTTP handlers
│   ├── service/          # Business logic
│   ├── repository/       # Data access layer
│   ├── model/            # Data models
│   └── storage/          # File storage
├── dist/                 # Build output (not committed)
├── docs/                 # Documentation
├── test/                 # Test data directory (not committed)
└── temp/                 # Temporary files (not committed)
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Add API endpoint | internal/handler/*.go | Follow existing handler pattern |
| Change DB schema | cmd/server/migrations/*.sql | Add migration file, auto-runs on startup |
| Modify frontend | cmd/server/static/js/app.js | Main frontend logic |
| Modify UI HTML | cmd/server/static/index.html | Modal and page structures |
| Modify styles | cmd/server/static/css/style.css | Custom CSS variables for theming |
| Add model | internal/model/model.go | Add struct with db/json tags |
| Add repository | internal/repository/*.go | Follow existing SQL const pattern |
| Add service | internal/service/*.go | Business logic layer |

## DATABASE MIGRATIONS

**Location**: `cmd/server/migrations/`

SQL files in this directory are automatically executed on server startup:
- Files ending with `.up.sql` are applied in alphabetical order
- Migration state tracked in `schema_migrations` table
- Only runs once per migration version
- **No `.down.sql` needed** - forward-only migrations

## TESTING GUIDELINES

### Test Directory Structure
- **Test files**: Create under `test/` directory with date-time based subdirectories
- **Format**: `test/YYYYMMDDHHMMSS/` (年月日时分秒) for test sessions
- **Example**: `test/20260202093045/` for testing on Feb 2, 2026 at 09:30:45

### Running Tests
```bash
# Create dated test directory with timestamp
mkdir -p test/$(date +%Y%m%d%H%M%S)

# Run with test database and storage paths
./dist/passworder --db test/20260202093045/test.db --storage test/20260202093045/storage

# Or for specific test scenarios
./dist/passworder --db test/qa-$(date +%Y%m%d%H%M%S).db --storage test/qa-storage-$(date +%Y%m%d%H%M%S)
```

### DO NOT
- Create `test_*.db` or `test_*` directories in project root
- Commit test databases or test data to git
- Mix test data with production data

## CODE MAP

| Symbol | Type | Location | Role |
|--------|------|----------|------|
| main | Function | cmd/server/main.go | Entry point: flags, config, DB, server |
| Router | Struct | internal/handler/router.go | Route registration and middleware |
| AccountHandler | Struct | internal/handler/account.go | Account CRUD + reminder sync |
| NoteAttachmentHandler | Struct | internal/handler/note_attachment.go | Multi-file attachment upload/download/preview |
| PersonalFileHandler | Struct | internal/handler/personal_file.go | Note CRUD + file preview |
| ReminderHandler | Struct | internal/handler/reminder.go | Reminder management + send-due |
| AccountService | Struct | internal/service/account.go | Account business logic |
| NoteAttachmentService | Struct | internal/service/note_attachment.go | Attachment file handling |
| ReminderService | Struct | internal/service/reminder.go | Reminder sync + SMTP sending |
| SMTPSender | Struct | internal/service/smtp_sender.go | Email sending via SMTP |
| Config | Struct | internal/config/config.go | CLI + env config loader |

## CONVENTIONS
- Go 1.20; module name `passworder`
- SQL migrations as numbered `.up.sql` files in `cmd/server/migrations/`
- JSON responses built via `model.Response` struct (Type/Message/Data)
- Frontend uses vanilla JS with MDUI-style CSS
- Chinese UI strings throughout
- `panic()` used for fatal startup errors only
- File storage uses `{timestamp}_{originalName}` format

## ANTI-PATTERNS (THIS PROJECT)
- **No input validation layers**: HTTP handlers read raw JSON into structs without validation
- **No structured logging**: Basic log package to stdout
- **Global mutable state**: Some package-level vars in handlers

## UNIQUE STYLES
- `//go:embed static/*` + `//go:embed all:migrations/*.sql` — assets compiled into binary
- Dynamic CDN library loading for Vditor, JSZip, mammoth.js, SheetJS
- Client-side file preview (images, archives, documents, PDF) without server conversion
- Multi-file upload with append-mode file selection
- UTF-8 encoded Content-Disposition for Chinese filenames

## COMMANDS
```bash
# Run
go run ./cmd/server

# Build (Linux/macOS, smaller binary)
go build -trimpath -ldflags="-s -w" -o dist/passworder ./cmd/server

# Build (Windows 64-bit, no CGO, smaller binary)
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o dist/passworder-windows-amd64.exe ./cmd/server

# Test
go test ./...

# With options (long flags use --, short flags use -)
./dist/passworder --host 0.0.0.0 --port 8080 --db ./password.db --storage ./storage
./dist/passworder -v
```

## NOTES
- Database auto-migrates on startup using embedded migration files
- SQLite driver requires CGO; cross-compile needs `CC` set
- Static files served from embedded FS at root path `/`
- All API routes prefixed with `/api`
- Session token stored in localStorage, passed via Authorization header

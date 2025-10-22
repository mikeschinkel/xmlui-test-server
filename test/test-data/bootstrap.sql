-- bootstrap.sql
-- Purpose: Comprehensive test schema for API integration testing with ALL PathVars parameter types
-- This schema supports testing every permutation from xmluisvr/pathvars/ADR_PATHVARS.md

PRAGMA foreign_keys = ON;

BEGIN;

-- =========================
-- USERS TABLE
-- =========================
CREATE TABLE IF NOT EXISTS users (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid        TEXT NOT NULL UNIQUE,                -- for UUID parameter testing
    email       TEXT NOT NULL UNIQUE,               -- for string parameter testing
    name        TEXT NOT NULL,                       -- for string length constraints
    slug        TEXT NOT NULL UNIQUE,               -- for slug parameter testing
    score       INTEGER NOT NULL DEFAULT 0,         -- for int range constraints
    rating      REAL NOT NULL DEFAULT 0.0,          -- for real/decimal parameter testing
    active      INTEGER NOT NULL DEFAULT 1,         -- for boolean parameter testing (SQLite uses INTEGER)
    created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    birth_date  TEXT,                               -- for date parameter testing
    login_time  TEXT                                -- for time parameter testing
);

-- Test users with specific data for parameter validation
INSERT OR IGNORE INTO users (id, uuid, email, name, slug, score, rating, active, birth_date, login_time) VALUES
    (1, '550e8400-e29b-41d4-a716-446655440001', 'alice@example.com', 'Alice Carter', 'alice-carter', 85, 4.5, 1, '1990-05-15', '09:30:45'),
    (2, '550e8400-e29b-41d4-a716-446655440002', 'bob@example.com', 'Bob Nguyen', 'bob-nguyen', 92, 4.8, 1, '1985-12-03', '14:22:10'),
    (3, '550e8400-e29b-41d4-a716-446655440003', 'carol@example.com', 'Carol Diaz', 'carol-diaz', 78, 3.9, 0, '1995-07-22', '11:45:30'),
    -- Additional test users for specific test cases
    (4, '550e8400-e29b-41d4-a716-446655440000', 'dana@example.com', 'Dana Kim', 'dana-kim', 88, 4.2, 1, '1990-01-15', '10:15:00'),
    (5, '550e8400-e29b-41d4-a716-446655440005', 'evan@example.com', 'Evan Jones', 'evan-jones', 73, 3.5, 0, '1992-08-20', '16:30:25');

-- =========================
-- PROJECTS TABLE
-- =========================
CREATE TABLE IF NOT EXISTS projects (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,                       -- for string parameter testing
    slug        TEXT NOT NULL,                       -- for slug validation
    status      TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','archived','draft')), -- for enum constraints
    priority    INTEGER NOT NULL DEFAULT 3 CHECK (priority BETWEEN 1 AND 5), -- for range constraints
    budget      REAL NOT NULL DEFAULT 0.0,          -- for decimal parameter testing
    created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    due_date    TEXT,                               -- for date parameter testing
    UNIQUE (owner_id, name)
);

-- Test projects
INSERT OR IGNORE INTO projects (owner_id, name, slug, status, priority, budget, due_date) VALUES
    (1, 'Website Redesign', 'website-redesign', 'active', 4, 15000.50, '2025-10-15'),
    (1, 'Mobile App MVP', 'mobile-app-mvp', 'active', 5, 25000.75, '2025-11-30'),
    (2, 'Data Migration', 'data-migration', 'draft', 2, 8000.25, '2025-09-01'),
    (3, 'API Documentation', 'api-docs', 'archived', 1, 3000.00, '2025-08-15'),
    (1, 'Test Project Alpha', 'test-project-alpha', 'active', 3, 12000.00, '2025-12-01'),
    (2, 'Legacy Project Archive', 'legacy-project-archive', 'archived', 2, 8500.00, '2025-07-01');

-- =========================
-- TASKS TABLE
-- =========================
CREATE TABLE IF NOT EXISTS tasks (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id   INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    assignee_id  INTEGER REFERENCES users(id) ON DELETE SET NULL,
    title        TEXT NOT NULL,
    details      TEXT,
    status       TEXT NOT NULL DEFAULT 'todo' CHECK (status IN ('todo','doing','done')),
    priority     INTEGER NOT NULL DEFAULT 3 CHECK (priority BETWEEN 1 AND 5),
    estimate     REAL NOT NULL DEFAULT 0.0,          -- for decimal testing
    due_date     TEXT,
    created_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    completed_at TEXT,
    UNIQUE (project_id, title)
);

-- Test tasks
INSERT OR IGNORE INTO tasks (project_id, assignee_id, title, status, priority, estimate, due_date) VALUES
    (1, 1, 'Audit current site', 'doing', 2, 8.5, '2025-10-01'),
    (1, 2, 'Design homepage', 'todo', 4, 16.0, '2025-10-10'),
    (1, 3, 'Implement responsive layout', 'todo', 3, 24.5, '2025-10-15'),
    (1, 1, 'Fix urgent security issue', 'todo', 5, 4.0, '2025-09-25'),
    (2, 2, 'Auth flow', 'doing', 5, 12.0, '2025-10-05'),
    (2, 1, 'API client', 'todo', 4, 20.0, '2025-10-08'),
    (2, 3, 'Push notifications', 'todo', 3, 18.5, '2025-10-18');

-- =========================
-- LOGS TABLE (for path* testing)
-- =========================
CREATE TABLE IF NOT EXISTS logs (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    level       TEXT NOT NULL CHECK (level IN ('DEBUG','INFO','WARN','ERROR')),
    message     TEXT NOT NULL,
    file_path   TEXT NOT NULL,                      -- for multi-segment path* parameter testing
    line_number INTEGER NOT NULL,
    created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    created_date TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d','now')), -- for date format testing
    created_time TEXT NOT NULL DEFAULT (strftime('%H:%M:%S','now'))  -- for time format testing
);

-- Test logs with various file paths for multi-segment testing
INSERT OR IGNORE INTO logs (level, message, file_path, line_number) VALUES
    ('INFO', 'Server started', 'src/main.go', 42),
    ('DEBUG', 'Database connected', 'src/db/connection.go', 15),
    ('WARN', 'Slow query detected', 'src/handlers/api.go', 127),
    ('ERROR', 'Connection failed', 'src/external/service.go', 89),
    ('INFO', 'Request processed', 'src/middleware/logging.go', 203);

-- =========================
-- MEASUREMENTS TABLE (for real/decimal testing)
-- =========================
CREATE TABLE IF NOT EXISTS measurements (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    sensor_id   TEXT NOT NULL,                      -- alphanumeric parameter testing
    value       REAL NOT NULL,                      -- for real parameter testing with range constraints
    unit        TEXT NOT NULL,
    precision_val REAL NOT NULL,                    -- for decimal parameter testing
    timestamp   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

-- Test measurements
INSERT OR IGNORE INTO measurements (sensor_id, value, unit, precision_val) VALUES
    ('TEMP001', 23.5, 'celsius', 0.1),
    ('HUMID002', 65.8, 'percent', 0.5),
    ('PRESS003', 1013.25, 'hPa', 0.01),
    ('LIGHT004', 750.0, 'lux', 1.0),
    -- Additional test measurement for alphanumeric parameter testing
    ('ABC123', 42.0, 'units', 0.5);

-- =========================
-- ARCHIVE TABLE (for date format testing)
-- =========================
CREATE TABLE IF NOT EXISTS archive (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    archive_date TEXT NOT NULL,                     -- for various date format testing
    archive_time TEXT NOT NULL,                     -- for time format testing
    content     TEXT NOT NULL,
    format_type TEXT NOT NULL                       -- tracks which format was used
);

-- Test archive entries with different date formats
INSERT OR IGNORE INTO archive (archive_date, archive_time, content, format_type) VALUES
    ('2025-01-15', '09:30:45', 'January data backup', 'yyyy-mm-dd'),
    ('2025/02/20', '14:22:10', 'February reports', 'yyyy/mm/dd'),
    ('2025-03-25_11:45:30', '', 'March analysis', 'yyyy-mm-dd_hh:mm:ss'),
    ('2025/04/10/08:15:20', '', 'April migration', 'yyyy/mm/dd/hh:mm:ss');

COMMIT;

PRAGMA foreign_keys = OFF;

-- Quick verification queries (commented out for production use)
-- SELECT 'users' as table_name, count(*) as count FROM users
-- UNION ALL SELECT 'projects', count(*) FROM projects
-- UNION ALL SELECT 'tasks', count(*) FROM tasks
-- UNION ALL SELECT 'logs', count(*) FROM logs
-- UNION ALL SELECT 'measurements', count(*) FROM measurements
-- UNION ALL SELECT 'archive', count(*) FROM archive;
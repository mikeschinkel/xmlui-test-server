package dbqvars_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/xmlui-org/localdev/xmluisvr/dbqvars"
)

// noinspection SqlResolveForFile

func TestParseSQL(t *testing.T) {
	tests := []struct {
		name            string
		sql             dbqvars.SQLQuery
		formatParamFunc dbqvars.FormatParamFunc
		expected        dbqvars.ParsedSQL
		expectError     bool
		expectedError   error
	}{
		// Basic cases
		{
			name: "LIKE '{file_path}'",
			sql:  "SELECT id FROM logs WHERE file_path LIKE {file_path} || '%';",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.NewParsedSQL("SELECT id FROM logs WHERE file_path LIKE $1 || '%';", dbqvars.NewParameters("file_path")),
		},
		{
			name: "no placeholders",
			sql:  "SELECT * FROM users WHERE active = true",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.NewParsedSQL("SELECT * FROM users WHERE active = true", nil),
		},
		{
			name: "single placeholder",
			sql:  "SELECT * FROM users WHERE id = {id}",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.NewParsedSQL("SELECT * FROM users WHERE id = $1", dbqvars.NewParameters("id")),
		},
		{
			name: "multiple unique placeholders",
			sql:  "SELECT * FROM orders WHERE account_id={accountId} AND created_at>={since}",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.NewParsedSQL("SELECT * FROM orders WHERE account_id=$1 AND created_at>=$2", dbqvars.NewParameters("accountId", "since")),
		},
		{
			name: "duplicate placeholders",
			sql:  "SELECT * FROM orders WHERE created_at >= {since} AND updated_at >= {since}",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.NewParsedSQL("SELECT * FROM orders WHERE created_at >= $1 AND updated_at >= $1", dbqvars.NewParameters("since")),
		},
		// Dotted path placeholders (ADR-008)
		{
			name: "dotted path placeholders",
			sql:  "INSERT INTO events (user_id, payload) VALUES ({user.id}, {body.event})",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.NewParsedSQL("INSERT INTO events (user_id, payload) VALUES ($1, $2)", dbqvars.NewParameters("user.id", "body.event")),
		},
		{
			name: "array index placeholders",
			sql:  "SELECT * FROM products WHERE id = {items.0.id} AND sku = {items.0.sku}",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.NewParsedSQL("SELECT * FROM products WHERE id = $1 AND sku = $2", dbqvars.NewParameters("items.0.id", "items.0.sku")),
		},
		// Different database backends
		{
			name: "mysql format",
			sql:  "SELECT * FROM users WHERE id = {id}",
			formatParamFunc: func(int) string {
				return "?"
			},
			expected: dbqvars.NewParsedSQL("SELECT * FROM users WHERE id = ?", dbqvars.NewParameters("id")),
		},
		{
			name: "sql server format",
			sql:  "SELECT * FROM users WHERE id = {id}",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("@p%d", i)
			},
			expected: dbqvars.NewParsedSQL("SELECT * FROM users WHERE id = @p1", dbqvars.NewParameters("id")),
		},
		// String literal skipping
		{
			name: "placeholder in single quotes ignored",
			sql:  "SELECT * FROM users WHERE name = 'John {id} Doe' AND id = {id}",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.NewParsedSQL("SELECT * FROM users WHERE name = 'John {id} Doe' AND id = $1", dbqvars.NewParameters("id")),
		},
		{
			name: "placeholder in double quotes ignored",
			sql:  `SELECT * FROM users WHERE name = "John {id} Doe" AND id = {id}`,
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.NewParsedSQL(`SELECT * FROM users WHERE name = "John {id} Doe" AND id = $1`, dbqvars.NewParameters("id")),
		},
		// Comment skipping
		{
			name: "placeholder in line comment ignored",
			sql:  "SELECT * FROM users -- WHERE id = {id}\nWHERE active = true AND id = {id}",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.NewParsedSQL("SELECT * FROM users -- WHERE id = {id}\nWHERE active = true AND id = $1", dbqvars.NewParameters("id")),
		},
		{
			name: "placeholder in block comment ignored",
			sql:  "SELECT * FROM users /* WHERE id = {id} */ WHERE active = true AND id = {id}",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.NewParsedSQL("SELECT * FROM users /* WHERE id = {id} */ WHERE active = true AND id = $1", dbqvars.NewParameters("id")),
		},
		// Whitespace handling
		{
			name: "placeholder with whitespace",
			sql:  "SELECT * FROM users WHERE id = { id }",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.NewParsedSQL("SELECT * FROM users WHERE id = $1", dbqvars.NewParameters("id")),
		},
		// Error cases
		{
			name:            "nil format function",
			sql:             "SELECT * FROM users WHERE id = {id}",
			formatParamFunc: nil,
			expectError:     true,
			expectedError:   dbqvars.ErrFormatParamFuncRequired,
		},
		{
			name: "unclosed placeholder",
			sql:  "SELECT * FROM users WHERE id = {id",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expectError:   true,
			expectedError: dbqvars.ErrUnclosedPlaceholder,
		},
		{
			name: "empty placeholder name",
			sql:  "SELECT * FROM users WHERE id = {}",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expectError:   true,
			expectedError: dbqvars.ErrInvalidPlaceholderName,
		},
		{
			name: "invalid placeholder name",
			sql:  "SELECT * FROM users WHERE id = {123invalid}",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expectError:   true,
			expectedError: dbqvars.ErrInvalidPlaceholderName,
		},
		{
			name: "placeholder with invalid characters",
			sql:  "SELECT * FROM users WHERE id = {id-with-dashes}",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expectError:   true,
			expectedError: dbqvars.ErrInvalidPlaceholderName,
		},
		// Complex realistic examples
		{
			name: "complex query with multiple features",
			sql: `SELECT u.*, p.title
FROM users u
LEFT JOIN posts p ON u.id = p.user_id
WHERE u.created_at >= {filters.since}
  AND u.status IN ('active', 'pending')
  AND (u.name LIKE {search.name} OR u.email LIKE {search.email})
  AND u.org_id = {auth.org_id}
ORDER BY u.created_at DESC`,
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.NewParsedSQL(
				`SELECT u.*, p.title
FROM users u
LEFT JOIN posts p ON u.id = p.user_id
WHERE u.created_at >= $1
  AND u.status IN ('active', 'pending')
  AND (u.name LIKE $2 OR u.email LIKE $3)
  AND u.org_id = $4
ORDER BY u.created_at DESC`,
				dbqvars.NewParameters("filters.since", "search.name", "search.email", "auth.org_id")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := dbqvars.ParseSQL(tt.sql, tt.formatParamFunc)

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error but got none")
				}
				if tt.expectedError != nil && !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.SQL != tt.expected.SQL {
				t.Errorf("SQL mismatch:\nexpected: %q\nactual:   %q", tt.expected.SQL, result.SQL)
			}

			// Compare only the essential fields (Name and Index)
			if len(result.Parameters()) != len(tt.expected.Parameters()) {
				t.Errorf("Parameters length mismatch: expected %d, got %d", len(tt.expected.Parameters()), len(result.Parameters()))
			} else {
				for i, expectedParam := range tt.expected.Parameters() {
					actualParam := result.Parameters()[i]
					if actualParam != expectedParam {
						t.Errorf("Param[%d] Name mismatch: expected %q, got %q", i, expectedParam, actualParam)
					}
				}
			}
		})
	}
}

func TestParseSQL_EdgeCases(t *testing.T) {
	tests := []struct {
		name            string
		sql             dbqvars.SQLQuery
		formatParamFunc dbqvars.FormatParamFunc
		expected        dbqvars.ParsedSQL
	}{
		{
			name: "escaped single quotes",
			sql:  "SELECT * FROM users WHERE name = 'O''Brien {id}' AND id = {id}",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.NewParsedSQL("SELECT * FROM users WHERE name = 'O''Brien {id}' AND id = $1", dbqvars.NewParameters("id")),
		},
		{
			name: "backtick identifiers",
			sql:  "SELECT * FROM `users` WHERE `user-id` = {id}",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.NewParsedSQL("SELECT * FROM `users` WHERE `user-id` = $1", dbqvars.NewParameters("id")),
		},
		{
			name: "bracket identifiers",
			sql:  "SELECT * FROM [users] WHERE [user-id] = {id}",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.NewParsedSQL("SELECT * FROM [users] WHERE [user-id] = $1", dbqvars.NewParameters("id")),
		},
		{
			name: "hash comment",
			sql:  "SELECT * FROM users # WHERE id = {id}\nWHERE active = true AND id = {id}",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.NewParsedSQL(
				"SELECT * FROM users # WHERE id = {id}\nWHERE active = true AND id = $1",
				dbqvars.NewParameters("id"),
			),
		},
		{
			name: "postgresql dollar quoting",
			sql:  "SELECT * FROM users WHERE desc = $tag${id} not a placeholder$tag$ AND id = {id}",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.NewParsedSQL(
				"SELECT * FROM users WHERE desc = $tag${id} not a placeholder$tag$ AND id = $1",
				dbqvars.NewParameters("id"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := dbqvars.ParseSQL(tt.sql, tt.formatParamFunc)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.SQL != tt.expected.SQL {
				t.Errorf("SQL mismatch:\nexpected: %q\nactual:   %q", tt.expected.SQL, result.SQL)
			}

			// Compare only the essential fields (Name and Index)
			if len(result.Parameters()) != len(tt.expected.Parameters()) {
				t.Errorf("Parameters length mismatch: expected %d, got %d", len(tt.expected.Parameters()), len(result.Parameters()))
			} else {
				for i, expectedParam := range tt.expected.Parameters() {
					actualParam := result.Parameters()[i]
					if actualParam != expectedParam {
						t.Errorf("Param[%d] Name mismatch: expected %q, got %q", i, expectedParam, actualParam)
					}
				}
			}
		})
	}
}

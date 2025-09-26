package dbqvars_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/dbqvars"
)

func TestParseSQL(t *testing.T) {
	tests := []struct {
		name            string
		sql             string
		formatParamFunc dbqvars.FormatParamFunc
		expected        dbqvars.ParsedSQL
		expectError     bool
		expectedError   error
	}{
		// Basic cases
		{
			name: "no placeholders",
			sql:  "SELECT * FROM users WHERE active = true",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.ParsedSQL{
				SQL:    "SELECT * FROM users WHERE active = true",
				Params: []dbqvars.SQLParam{},
			},
		},
		{
			name: "single placeholder",
			sql:  "SELECT * FROM users WHERE id = {id}",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.ParsedSQL{
				SQL: "SELECT * FROM users WHERE id = $1",
				Params: []dbqvars.SQLParam{
					{Name: "id", Index: 1},
				},
			},
		},
		{
			name: "multiple unique placeholders",
			sql:  "SELECT * FROM orders WHERE account_id = {accountId} AND created_at >= {since}",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.ParsedSQL{
				SQL: "SELECT * FROM orders WHERE account_id = $1 AND created_at >= $2",
				Params: []dbqvars.SQLParam{
					{Name: "accountId", Index: 1},
					{Name: "since", Index: 2},
				},
			},
		},
		{
			name: "duplicate placeholders",
			sql:  "SELECT * FROM orders WHERE created_at >= {since} AND updated_at >= {since}",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.ParsedSQL{
				SQL: "SELECT * FROM orders WHERE created_at >= $1 AND updated_at >= $1",
				Params: []dbqvars.SQLParam{
					{Name: "since", Index: 1},
				},
			},
		},
		// Dotted path placeholders (ADR-008)
		{
			name: "dotted path placeholders",
			sql:  "INSERT INTO events (user_id, payload) VALUES ({user.id}, {body.event})",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.ParsedSQL{
				SQL: "INSERT INTO events (user_id, payload) VALUES ($1, $2)",
				Params: []dbqvars.SQLParam{
					{Name: "user.id", Index: 1},
					{Name: "body.event", Index: 2},
				},
			},
		},
		{
			name: "array index placeholders",
			sql:  "SELECT * FROM products WHERE id = {items[0].id} AND sku = {items[0].sku}",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.ParsedSQL{
				SQL: "SELECT * FROM products WHERE id = $1 AND sku = $2",
				Params: []dbqvars.SQLParam{
					{Name: "items[0].id", Index: 1},
					{Name: "items[0].sku", Index: 2},
				},
			},
		},
		// Different database backends
		{
			name: "mysql format",
			sql:  "SELECT * FROM users WHERE id = {id}",
			formatParamFunc: func(int) string {
				return "?"
			},
			expected: dbqvars.ParsedSQL{
				SQL: "SELECT * FROM users WHERE id = ?",
				Params: []dbqvars.SQLParam{
					{Name: "id", Index: 1},
				},
			},
		},
		{
			name: "sql server format",
			sql:  "SELECT * FROM users WHERE id = {id}",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("@p%d", i)
			},
			expected: dbqvars.ParsedSQL{
				SQL: "SELECT * FROM users WHERE id = @p1",
				Params: []dbqvars.SQLParam{
					{Name: "id", Index: 1},
				},
			},
		},
		// String literal skipping
		{
			name: "placeholder in single quotes ignored",
			sql:  "SELECT * FROM users WHERE name = 'John {id} Doe' AND id = {id}",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.ParsedSQL{
				SQL: "SELECT * FROM users WHERE name = 'John {id} Doe' AND id = $1",
				Params: []dbqvars.SQLParam{
					{Name: "id", Index: 1},
				},
			},
		},
		{
			name: "placeholder in double quotes ignored",
			sql:  `SELECT * FROM users WHERE name = "John {id} Doe" AND id = {id}`,
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.ParsedSQL{
				SQL: `SELECT * FROM users WHERE name = "John {id} Doe" AND id = $1`,
				Params: []dbqvars.SQLParam{
					{Name: "id", Index: 1},
				},
			},
		},
		// Comment skipping
		{
			name: "placeholder in line comment ignored",
			sql:  "SELECT * FROM users -- WHERE id = {id}\nWHERE active = true AND id = {id}",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.ParsedSQL{
				SQL: "SELECT * FROM users -- WHERE id = {id}\nWHERE active = true AND id = $1",
				Params: []dbqvars.SQLParam{
					{Name: "id", Index: 1},
				},
			},
		},
		{
			name: "placeholder in block comment ignored",
			sql:  "SELECT * FROM users /* WHERE id = {id} */ WHERE active = true AND id = {id}",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.ParsedSQL{
				SQL: "SELECT * FROM users /* WHERE id = {id} */ WHERE active = true AND id = $1",
				Params: []dbqvars.SQLParam{
					{Name: "id", Index: 1},
				},
			},
		},
		// Whitespace handling
		{
			name: "placeholder with whitespace",
			sql:  "SELECT * FROM users WHERE id = { id }",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.ParsedSQL{
				SQL: "SELECT * FROM users WHERE id = $1",
				Params: []dbqvars.SQLParam{
					{Name: "id", Index: 1},
				},
			},
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
			expected: dbqvars.ParsedSQL{
				SQL: `SELECT u.*, p.title
FROM users u
LEFT JOIN posts p ON u.id = p.user_id
WHERE u.created_at >= $1
  AND u.status IN ('active', 'pending')
  AND (u.name LIKE $2 OR u.email LIKE $3)
  AND u.org_id = $4
ORDER BY u.created_at DESC`,
				Params: []dbqvars.SQLParam{
					{Name: "filters.since", Index: 1},
					{Name: "search.name", Index: 2},
					{Name: "search.email", Index: 3},
					{Name: "auth.org_id", Index: 4},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := dbqvars.ParseSQL(tt.sql, dbqvars.ParseSQLArgs{
				FormatParamFunc: tt.formatParamFunc,
			})

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error but got none")
				}
				if tt.expectedError != nil && !strings.Contains(err.Error(), tt.expectedError.Error()) {
					t.Errorf("expected error containing %q, got %q", tt.expectedError.Error(), err.Error())
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
			if len(result.Params) != len(tt.expected.Params) {
				t.Errorf("Params length mismatch: expected %d, got %d", len(tt.expected.Params), len(result.Params))
			} else {
				for i, expectedParam := range tt.expected.Params {
					actualParam := result.Params[i]
					if actualParam.Name != expectedParam.Name {
						t.Errorf("Param[%d] Name mismatch: expected %q, got %q", i, expectedParam.Name, actualParam.Name)
					}
					if actualParam.Index != expectedParam.Index {
						t.Errorf("Param[%d] Index mismatch: expected %d, got %d", i, expectedParam.Index, actualParam.Index)
					}
				}
			}
		})
	}
}

func TestParseSQL_EdgeCases(t *testing.T) {
	tests := []struct {
		name            string
		sql             string
		formatParamFunc dbqvars.FormatParamFunc
		expected        dbqvars.ParsedSQL
	}{
		{
			name: "escaped single quotes",
			sql:  "SELECT * FROM users WHERE name = 'O''Brien {id}' AND id = {id}",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.ParsedSQL{
				SQL: "SELECT * FROM users WHERE name = 'O''Brien {id}' AND id = $1",
				Params: []dbqvars.SQLParam{
					{Name: "id", Index: 1},
				},
			},
		},
		{
			name: "backtick identifiers",
			sql:  "SELECT * FROM `users` WHERE `user-id` = {id}",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.ParsedSQL{
				SQL: "SELECT * FROM `users` WHERE `user-id` = $1",
				Params: []dbqvars.SQLParam{
					{Name: "id", Index: 1},
				},
			},
		},
		{
			name: "bracket identifiers",
			sql:  "SELECT * FROM [users] WHERE [user-id] = {id}",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.ParsedSQL{
				SQL: "SELECT * FROM [users] WHERE [user-id] = $1",
				Params: []dbqvars.SQLParam{
					{Name: "id", Index: 1},
				},
			},
		},
		{
			name: "hash comment",
			sql:  "SELECT * FROM users # WHERE id = {id}\nWHERE active = true AND id = {id}",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.ParsedSQL{
				SQL: "SELECT * FROM users # WHERE id = {id}\nWHERE active = true AND id = $1",
				Params: []dbqvars.SQLParam{
					{Name: "id", Index: 1},
				},
			},
		},
		{
			name: "postgresql dollar quoting",
			sql:  "SELECT * FROM users WHERE desc = $tag${id} not a placeholder$tag$ AND id = {id}",
			formatParamFunc: func(i int) string {
				return fmt.Sprintf("$%d", i)
			},
			expected: dbqvars.ParsedSQL{
				SQL: "SELECT * FROM users WHERE desc = $tag${id} not a placeholder$tag$ AND id = $1",
				Params: []dbqvars.SQLParam{
					{Name: "id", Index: 1},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := dbqvars.ParseSQL(tt.sql, dbqvars.ParseSQLArgs{
				FormatParamFunc: tt.formatParamFunc,
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.SQL != tt.expected.SQL {
				t.Errorf("SQL mismatch:\nexpected: %q\nactual:   %q", tt.expected.SQL, result.SQL)
			}

			// Compare only the essential fields (Name and Index)
			if len(result.Params) != len(tt.expected.Params) {
				t.Errorf("Params length mismatch: expected %d, got %d", len(tt.expected.Params), len(result.Params))
			} else {
				for i, expectedParam := range tt.expected.Params {
					actualParam := result.Params[i]
					if actualParam.Name != expectedParam.Name {
						t.Errorf("Param[%d] Name mismatch: expected %q, got %q", i, expectedParam.Name, actualParam.Name)
					}
					if actualParam.Index != expectedParam.Index {
						t.Errorf("Param[%d] Index mismatch: expected %d, got %d", i, expectedParam.Index, actualParam.Index)
					}
				}
			}
		})
	}
}

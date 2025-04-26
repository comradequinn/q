package schema

import (
	"testing"
)

func TestBuild(t *testing.T) {
	testCases := []struct {
		name        string
		definition  string
		expected    string
		expectError bool
	}{
		{
			name:       "Single full, valid definitions",
			definition: "id:integer:User.ID, with some punctation!",
			expected:   `{"properties":{"id":{"description":"User.ID, with some punctation!","type":"integer"}},"type":"object"}`,
		},
		{
			name:       "Single full, valid definition as array",
			definition: "[]id:integer:User.ID, with some punctation!",
			expected:   `{"items":{"properties":{"id":{"description":"User.ID, with some punctation!","type":"integer"}},"type":"object"},"type":"array"}`,
		},
		{
			name:       "Multiple full, valid definitions",
			definition: "id:integer:User ID|name:string:User name|email:string:User email address",
			expected:   `{"properties":{"email":{"description":"User email address","type":"string"},"id":{"description":"User ID","type":"integer"},"name":{"description":"User name","type":"string"}},"type":"object"}`,
		},
		{
			name:       "Single partial, valid definitions",
			definition: "id:integer",
			expected:   `{"properties":{"id":{"description":"","type":"integer"}},"type":"object"}`,
		},
		{
			name:       "Multiple partial, valid definitions",
			definition: "id:integer|name:string|email:string",
			expected:   `{"properties":{"email":{"description":"","type":"string"},"id":{"description":"","type":"integer"},"name":{"description":"","type":"string"}},"type":"object"}`,
		},
		{
			name:       "Multiple partial and full, valid definitions",
			definition: "id:integer|name:string:User name|email:string",
			expected:   `{"properties":{"email":{"description":"","type":"string"},"id":{"description":"","type":"integer"},"name":{"description":"User name","type":"string"}},"type":"object"}`,
		},
		{
			name:        "Single, invalid definition",
			definition:  "id/integer",
			expectError: true,
		},
		{
			name:        "Mixed, valid and invalid definition",
			definition:  "id:integer|name:string:User name:invalid|email:string",
			expectError: true,
		},
	}

	assert := func(t *testing.T, condition bool, format string, v ...any) {
		if !condition {
			t.Fatalf(format, v...)
		}
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			json, err := Build(tc.definition)

			assert(t, err == nil != tc.expectError, "error expectation was %v. got %v", tc.expectError, err)
			assert(t, string(json) == tc.expected, "expected json '%v'. got '%v'", tc.expected, string(json))
		})
	}
}

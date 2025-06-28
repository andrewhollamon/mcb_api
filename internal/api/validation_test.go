package api

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

func TestValidationCheckboxNumber(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		checkboxNbr    string
		expectedResult int
		expectError    bool
		errorContains  string
	}{
		{
			name:           "Valid number - minimum bound",
			checkboxNbr:    "1",
			expectedResult: 1,
			expectError:    false,
		},
		{
			name:           "Valid number - maximum bound",
			checkboxNbr:    "1000000",
			expectedResult: 1000000,
			expectError:    false,
		},
		{
			name:           "Valid number - middle range",
			checkboxNbr:    "500000",
			expectedResult: 500000,
			expectError:    false,
		},
		{
			name:           "Empty parameter",
			checkboxNbr:    "",
			expectedResult: 0,
			expectError:    true,
			errorContains:  "Missing checkbox number parameter",
		},
		{
			name:           "Blank parameter with spaces",
			checkboxNbr:    "   ",
			expectedResult: 0,
			expectError:    true,
			errorContains:  "Missing checkbox number parameter",
		},
		{
			name:           "Zero - below minimum",
			checkboxNbr:    "0",
			expectedResult: 0,
			expectError:    true,
			errorContains:  "out of range",
		},
		{
			name:           "Negative number",
			checkboxNbr:    "-1",
			expectedResult: 0,
			expectError:    true,
			errorContains:  "out of range",
		},
		{
			name:           "Above maximum bound",
			checkboxNbr:    "1000001",
			expectedResult: 0,
			expectError:    true,
			errorContains:  "out of range",
		},
		{
			name:           "Non-integer string",
			checkboxNbr:    "abc",
			expectedResult: 0,
			expectError:    true,
			errorContains:  "not a valid integer",
		},
		{
			name:           "Decimal number",
			checkboxNbr:    "123.45",
			expectedResult: 0,
			expectError:    true,
			errorContains:  "not a valid integer",
		},
		{
			name:           "Number with whitespace - should be trimmed",
			checkboxNbr:    "  123  ",
			expectedResult: 123,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// mock gin context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/", nil)
			c.Params = gin.Params{
				{Key: "checkboxNbr", Value: tt.checkboxNbr},
			}

			result, err := validateCheckboxNumber(c)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Equal(t, tt.expectedResult, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestValidateUserUuid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		userUuid      string
		expectError   bool
		errorContains string
		expectedUuid  uuid.UUID
	}{
		{
			name:         "Valid UUIDv7",
			userUuid:     "01234567-89ab-7def-89ab-123456789abc",
			expectError:  false,
			expectedUuid: uuid.MustParse("01234567-89ab-7def-89ab-123456789abc"),
		},
		{
			name:         "Valid UUIDv4",
			userUuid:     "550e8400-e29b-41d4-a716-446655440000",
			expectError:  false,
			expectedUuid: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		},
		{
			name:         "Valid UUID with uppercase letters",
			userUuid:     "550E8400-E29B-41D4-A716-446655440000",
			expectError:  false,
			expectedUuid: uuid.MustParse("550E8400-E29B-41D4-A716-446655440000"),
		},
		{
			name:          "Empty parameter",
			userUuid:      "",
			expectError:   true,
			errorContains: "Missing user UUID parameter",
			expectedUuid:  uuid.Nil,
		},
		{
			name:          "Whitespace only parameter",
			userUuid:      "   ",
			expectError:   true,
			errorContains: "Missing user UUID parameter",
			expectedUuid:  uuid.Nil,
		},
		{
			name:          "Tab and space parameter",
			userUuid:      "\t  \n",
			expectError:   true,
			errorContains: "Missing user UUID parameter",
			expectedUuid:  uuid.Nil,
		},
		{
			name:          "Too short - 35 characters",
			userUuid:      "550e8400-e29b-41d4-a716-44665544000",
			expectError:   true,
			errorContains: "not a valid UUID",
			expectedUuid:  uuid.Nil,
		},
		{
			name:          "Too long - 37 characters",
			userUuid:      "550e8400-e29b-41d4-a716-4466554400000",
			expectError:   true,
			errorContains: "not a valid UUID",
			expectedUuid:  uuid.Nil,
		},
		{
			name:         "Missing dashes",
			userUuid:     "550e8400e29b41d4a716446655440000",
			expectError:  false,
			expectedUuid: uuid.MustParse("550e8400e29b41d4a716446655440000"),
		},
		{
			name:          "Wrong dash positions",
			userUuid:      "550e840-0e29b-41d4-a716-446655440000",
			expectError:   true,
			errorContains: "not a valid UUID",
			expectedUuid:  uuid.Nil,
		},
		{
			name:          "Invalid characters - contains 'g'",
			userUuid:      "550e8400-e29b-41d4-a716-44665544000g",
			expectError:   true,
			errorContains: "not a valid UUID",
			expectedUuid:  uuid.Nil,
		},
		{
			name:          "Invalid characters - contains special chars",
			userUuid:      "550e8400-e29b-41d4-a716-44665544000!",
			expectError:   true,
			errorContains: "not a valid UUID",
			expectedUuid:  uuid.Nil,
		},
		{
			name:          "Invalid characters - contains space",
			userUuid:      "550e8400-e29b-41d4-a716-44665544 000",
			expectError:   true,
			errorContains: "not a valid UUID",
			expectedUuid:  uuid.Nil,
		},
		{
			name:         "All zeros",
			userUuid:     "00000000-0000-0000-0000-000000000000",
			expectError:  false,
			expectedUuid: uuid.MustParse("00000000-0000-0000-0000-000000000000"),
		},
		{
			name:         "All uppercase hex",
			userUuid:     "FFFFFFFF-FFFF-FFFF-FFFF-FFFFFFFFFFFF",
			expectError:  false,
			expectedUuid: uuid.MustParse("FFFFFFFF-FFFF-FFFF-FFFF-FFFFFFFFFFFF"),
		},
		{
			name:         "Valid UUID with surrounding whitespace - should be trimmed",
			userUuid:     "  550e8400-e29b-41d4-a716-446655440000  ",
			expectError:  false,
			expectedUuid: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		},
		{
			name:          "Only dashes",
			userUuid:      "--------",
			expectError:   true,
			errorContains: "not a valid UUID",
			expectedUuid:  uuid.Nil,
		},
		{
			name:          "Completely invalid format",
			userUuid:      "not-a-uuid-at-all",
			expectError:   true,
			errorContains: "not a valid UUID",
			expectedUuid:  uuid.Nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// mock gin context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/", nil)
			c.Params = gin.Params{
				{Key: "userUuid", Value: tt.userUuid},
			}

			result, err := validateUserUUID(c)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Equal(t, tt.expectedUuid, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUuid, result)
			}
		})
	}
}

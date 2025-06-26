package api

import (
	"github.com/gin-gonic/gin"
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

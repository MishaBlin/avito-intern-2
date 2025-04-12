package dummyLogin

import (
	"avito-intern/internal/api/dto/request/authDto"
	"avito-intern/internal/api/dto/response"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDummyLoginHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestData    authDto.DummyLoginRequest
		invalidBody    bool
		expectedStatus int
		expectedResp   interface{}
	}{
		{
			name: "Successful dummy login - employee role",
			requestData: authDto.DummyLoginRequest{
				Role: "employee",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Successful dummy login - moderator role",
			requestData: authDto.DummyLoginRequest{
				Role: "moderator",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid role",
			requestData: authDto.DummyLoginRequest{
				Role: "invalid-role",
			},
			expectedStatus: http.StatusBadRequest,
			expectedResp:   response.ErrorResponse{Message: "Invalid role"},
		},
		{
			name:           "Invalid request body",
			invalidBody:    true,
			expectedStatus: http.StatusBadRequest,
			expectedResp:   response.ErrorResponse{Message: "Invalid request"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := New()

			var req *http.Request
			var err error

			if tt.invalidBody {
				req = httptest.NewRequest(http.MethodPost, "/dummy-login", strings.NewReader("invalid json"))
			} else {
				body, _ := json.Marshal(tt.requestData)
				req = httptest.NewRequest(http.MethodPost, "/dummy-login", bytes.NewReader(body))
			}

			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			require.Equal(t, tt.expectedStatus, w.Code)

			var resp interface{}
			body, err := io.ReadAll(w.Body)
			require.NoError(t, err)

			if tt.invalidBody || tt.name == "Invalid role" {
				var errorResp response.ErrorResponse
				err = json.Unmarshal(body, &errorResp)
				require.NoError(t, err)
				resp = errorResp
				require.Equal(t, tt.expectedResp, resp)
			} else if tt.expectedStatus == http.StatusOK {
				var tokenResp response.TokenResponse
				err = json.Unmarshal(body, &tokenResp)
				require.NoError(t, err)
				require.NotEmpty(t, tokenResp.Token, "Token should not be empty")
			}
		})
	}
}

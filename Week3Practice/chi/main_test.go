package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// executeRequest, creates a new ResponseRecorder
// then executes the request by calling ServeHTTP in the router
// after which the handler writes the response to the response recorder
// which we can then inspect.
func executeRequest(req *http.Request, s *Server) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	s.Router.ServeHTTP(rr, req)

	return rr
}

func expectEqual[T comparable](t *testing.T, expected T, actual T) {
	if expected != actual {
		t.Fatalf("Expected: %v\ndoes not match\nActual: %v",
			expected,
			actual)
	}
}

func TestHealthCheck(t *testing.T) {
	// Create a New Server Struct
	s := CreateNewServer()
	// Mount Handlers
	s.MountHandlers()

	// Create a New Request
	req, _ := http.NewRequest("GET", "/health", nil)

	// Execute Request
	response := executeRequest(req, s)

	// Test checks
	expectEqual(t, http.StatusOK, response.Code)

	expectedBody := `{"alive": true}`
	expectEqual(t, expectedBody, response.Body.String())
}

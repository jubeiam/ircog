// endpoints_test.go
package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func testControllerHelper(t *testing.T, req *http.Request, handler func(w http.ResponseWriter, r *http.Request), expected string) {
	rr := httptest.NewRecorder()
	handlerFunc := http.HandlerFunc(handler)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handlerFunc.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestPingHandler(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/any", nil)
	if err != nil {
		t.Fatal(err)
	}

	expected := `{"ping": "pong"}`
	testControllerHelper(t, req, pingController, expected)
}

func TestHelloHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/any?hello=leszek", nil)
	if err != nil {
		t.Fatal(err)
	}

	expected := `{"hello": "leszek"}`
	testControllerHelper(t, req, helloController, expected)
}

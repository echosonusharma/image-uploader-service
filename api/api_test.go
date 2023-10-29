package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPingHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	apiServer := &ApiServer{}

	handler := http.HandlerFunc(makeHTTPHandlerFunc(apiServer.PingHandler))
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resMsg Msg
	decodeErr := json.NewDecoder(rr.Body).Decode(&resMsg)
	if decodeErr != nil {
		t.Fatalf("Failed to decode response msg: %v", decodeErr)
	}

	expectedMsg := "🚀"

	if resMsg.Msg != expectedMsg {
		t.Errorf("Handler returned unexpected body: got %v want %v", resMsg.Msg, expectedMsg)
	}
}

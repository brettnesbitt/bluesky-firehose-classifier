package mlclassifier_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"stockseer.ai/blueksy-firehose/internal/apis/mlclassifier"
)

func TestClient_Classify_Success_httptest(t *testing.T) {

	// Create a mock HTTP server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Respond with a mock JSON response
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `[{"Label": "positive", "Score": 0.9}]`)
	}))
	defer testServer.Close()

	client := mlclassifier.NewClient(testServer.URL)
	requestData := mlclassifier.DataRequest{
		Items: []mlclassifier.DataRequestItem{
			{Text: "This is a positive text"},
		},
	}
	responseData, err := client.Classify(requestData)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(*responseData) != 1 {
		t.Errorf("Expected 1 response item, got %d", len(*responseData))
	}

	if (*responseData)[0].Label != "positive" {
		t.Errorf("Expected Label to be 'positive', got %s", (*responseData)[0].Label)
	}
	if (*responseData)[0].Score != 0.9 {
		t.Errorf("Expected Score to be 0.9, got %f", (*responseData)[0].Score)
	}
}

func TestClient_Classify_Error_Non200Status_httptest(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Internal Server Error")
	}))
	defer testServer.Close()

	client := mlclassifier.NewClient(testServer.URL)
	requestData := mlclassifier.DataRequest{
		Items: []mlclassifier.DataRequestItem{
			{Text: "This is a text"},
		},
	}
	_, err := client.Classify(requestData)

	if err == nil {
		t.Error("Expected error for non-200 status code")
	}
}

// Example using gomock (for more complex scenarios)
type MockHTTPClient struct {
	ctrl   *gomock.Controller
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func TestClient_Classify_Success(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")   // Important: Set Content-Type header
		fmt.Fprintln(w, `[{"label": "test", "score": 0.8}]`) // Correct JSON response
	}))
	defer testServer.Close()

	client := mlclassifier.NewClient(testServer.URL)
	requestData := mlclassifier.DataRequest{
		Items: []mlclassifier.DataRequestItem{
			{Text: "Test"},
		},
	}

	resp, err := client.Classify(requestData)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(*resp) != 1 {
		t.Errorf("Expected 1 response item, got %d", len(*resp))
	}

	if (*resp)[0].Label != "test" {
		t.Errorf("Expected label to be 'test', got %s", (*resp)[0].Label)
	}

	if (*resp)[0].Score != 0.8 {
		t.Errorf("Expected score to be 0.8, got %f", (*resp)[0].Score)
	}
}

func TestClient_Classify_Error_Marshal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Respond with a mock JSON response
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, ``)
	}))
	defer testServer.Close()

	client := mlclassifier.NewClient(testServer.URL)

	requestData := mlclassifier.DataRequest{
		Items: []mlclassifier.DataRequestItem{
			{Text: "This is a text"},
		},
	}
	_, err := client.Classify(requestData)

	if err == nil {
		t.Errorf("Expected error when marshaling request data")
	}
}

func TestClient_Classify_Error_Request(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mlclassifier.NewClient("")

	requestData := mlclassifier.DataRequest{
		Items: []mlclassifier.DataRequestItem{
			{Text: "This is a text"},
		},
	}
	_, err := client.Classify(requestData)

	if err == nil {
		t.Errorf("Expected error when creating request (empty baseURL)")
	}
}

func TestClient_Classify_Error_Non200Status(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Internal Server Error")
	}))

	defer testServer.Close()

	client := mlclassifier.NewClient(testServer.URL)

	requestData := mlclassifier.DataRequest{
		Items: []mlclassifier.DataRequestItem{
			{Text: "This is a text"},
		},
	}
	_, err := client.Classify(requestData)

	if err == nil {
		t.Errorf("Expected error for non-200 status code")
	}

	if !strings.HasPrefix(err.Error(), "unexpected status code: 500") {
		t.Errorf("Unexpected error message: %v", err)
	}
}

package mlclassifier_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"stockseer.ai/blueksy-firehose/internal/apis/mlclassifier"
)

type MockHTTPClient struct {
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
		w.Header().Set("Content-Type", "application/json")  // Important: Set Content-Type header
		fmt.Println(w, `[{"label": "test", "score": 0.8}]`) // Correct JSON response
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
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Respond with a mock JSON response
		w.Header().Set("Content-Type", "application/json")
		fmt.Println(w, ``)
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
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(w, "Internal Server Error")
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

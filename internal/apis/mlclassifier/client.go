package mlclassifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}
}

func (c *Client) Classify(requestData DataRequest) (*[]DataResponseItem, error) {
	url := fmt.Sprintf("%s/classify", c.BaseURL)

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("marshaling request data: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("reading response body: %w", err)
		}
		bodyString := string(bodyBytes)
		return nil, fmt.Errorf("unexpected status code: %d, response body: %s", resp.StatusCode, bodyString)
	}

	var dataResponse []DataResponseItem
	if err := json.NewDecoder(resp.Body).Decode(&dataResponse); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &dataResponse, nil
}

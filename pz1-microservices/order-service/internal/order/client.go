package order

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// UserServiceClient is a small HTTP client for user-service.
type UserServiceClient struct {
	baseURL string
	client  *http.Client
}

func NewUserServiceClient(baseURL string) *UserServiceClient {
	return &UserServiceClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func (c *UserServiceClient) GetUserByID(ctx context.Context, id int64) (UserDTO, error) {
	url := fmt.Sprintf("%s/users/%d", c.baseURL, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return UserDTO{}, fmt.Errorf("create request to user-service failed: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return UserDTO{}, fmt.Errorf("request to user-service failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return UserDTO{}, fmt.Errorf("user not found in user-service")
	}

	if resp.StatusCode != http.StatusOK {
		return UserDTO{}, fmt.Errorf("user-service returned status %d", resp.StatusCode)
	}

	var user UserDTO
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return UserDTO{}, fmt.Errorf("decode user response failed: %w", err)
	}

	return user, nil
}

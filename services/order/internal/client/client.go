package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kostayne/go-microservice/pkg/telemetry"
)

type UserClient struct {
	baseURL string
	http    *http.Client
}

func NewUserClient(baseURL string) *UserClient {
	return &UserClient{baseURL: baseURL, http: telemetry.HTTPClient(10 * time.Second)}
}

func (c *UserClient) ValidateUser(ctx context.Context, userID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/internal/validate/"+userID, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("user validation failed: %d", resp.StatusCode)
	}
	return nil
}

type RestaurantClient struct {
	baseURL string
	http    *http.Client
}

func NewRestaurantClient(baseURL string) *RestaurantClient {
	return &RestaurantClient{baseURL: baseURL, http: telemetry.HTTPClient(10 * time.Second)}
}

type menuItem struct {
	ID           string  `json:"id"`
	RestaurantID string  `json:"restaurant_id"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
}

type validateMenuResponse struct {
	Items []menuItem `json:"items"`
}

func (c *RestaurantClient) ValidateMenu(ctx context.Context, restaurantID string, itemIDs []string) ([]menuItem, error) {
	body, _ := json.Marshal(map[string]any{
		"restaurant_id": restaurantID,
		"item_ids":      itemIDs,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/menu/validate", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("menu validation failed: %d", resp.StatusCode)
	}

	var result validateMenuResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Items, nil
}

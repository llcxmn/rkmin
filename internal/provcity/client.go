package provcity

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	base string
	http *http.Client
}

func New(base string) *Client {
	return &Client{base: strings.TrimRight(base, "/"), http: &http.Client{Timeout: 10 * time.Second}}
}

func (c *Client) ListProvinces(search string, page, limit int) (any, error) {
	var data []map[string]any
	if err := c.get("/provinces.json", &data); err != nil {
		return nil, err
	}
	return filterPage(data, "name", search, page, limit), nil
}

func (c *Client) ListCities(provID string, search string, page, limit int) (any, error) {
	var data []map[string]any
	if err := c.get(fmt.Sprintf("/regencies/%s.json", url.PathEscape(provID)), &data); err != nil {
		return nil, err
	}
	return filterPage(data, "name", search, page, limit), nil
}

func (c *Client) DetailProvince(id string) (any, error) {
	var data map[string]any
	return data, c.get(fmt.Sprintf("/province/%s.json", url.PathEscape(id)), &data)
}

func (c *Client) DetailCity(id string) (any, error) {
	var data map[string]any
	return data, c.get(fmt.Sprintf("/regency/%s.json", url.PathEscape(id)), &data)
}

func (c *Client) get(path string, out any) error {
	resp, err := c.http.Get(c.base + path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("province city API returned %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func filterPage(rows []map[string]any, field, search string, page, limit int) []map[string]any {
	filtered := rows
	if search != "" {
		filtered = filtered[:0]
		for _, row := range rows {
			val, _ := row[field].(string)
			if strings.Contains(strings.ToLower(val), strings.ToLower(search)) {
				filtered = append(filtered, row)
			}
		}
	}
	if limit <= 0 {
		return filtered
	}
	if page <= 0 {
		page = 1
	}
	start := (page - 1) * limit
	if start >= len(filtered) {
		return []map[string]any{}
	}
	end := start + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[start:end]
}

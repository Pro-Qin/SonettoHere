// Package apiclient provides shared HTTP clients for third-party APIs used by
// built-in tools: UAPIS (weather, holidays, tarot, image understanding),
// 高德地图 (AMAP), and Todoist.
package apiclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Env vars for API keys.
const (
	EnvUapisKey     = "UAPIS_API_KEY"
	EnvAmapKey      = "AMAP_API_KEY"
	EnvTodoistToken = "TODOIST_API_TOKEN"
)

// UapisClient calls the UAPIS API (uapis.cn).
type UapisClient struct {
	baseURL string
	token   string
	http    *http.Client
}

// NewUapisClient creates a UapisClient from its env var.
func NewUapisClient() *UapisClient {
	token := os.Getenv(EnvUapisKey)
	return &UapisClient{
		baseURL: "https://uapis.cn",
		token:   token,
		http: &http.Client{
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:    5,
				IdleConnTimeout: 60 * time.Second,
			},
		},
	}
}

func (c *UapisClient) doGet(path string, params map[string]string) (map[string]any, error) {
	u, _ := url.Parse(c.baseURL + path)
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	req, _ := http.NewRequest("GET", u.String(), nil)
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("uapis request %s: %w", path, err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("uapis decode %s: %w", path, err)
	}
	return result, nil
}

// GetWeather calls the UAPIS weather endpoint.
// See: SonettoHere tools/network/tool_weather.py
func (c *UapisClient) GetWeather(city, adcode string, extended, forecast, hourly, minutely, indices bool, lang string) (map[string]any, error) {
	params := map[string]string{
		"extended": fmt.Sprintf("%v", extended),
		"forecast": fmt.Sprintf("%v", forecast),
		"hourly":   fmt.Sprintf("%v", hourly),
		"minutely": fmt.Sprintf("%v", minutely),
		"indices":  fmt.Sprintf("%v", indices),
		"lang":     lang,
	}
	if city != "" {
		params["city"] = city
	}
	if adcode != "" {
		params["adcode"] = adcode
	}
	return c.doGet("/misc/weather", params)
}

// GetHolidayCalendar calls the UAPIS holiday endpoint.
func (c *UapisClient) GetHolidayCalendar(date, month, year, timezone, holidayType string, includeNearby bool, nearbyLimit int) (map[string]any, error) {
	params := map[string]string{
		"timezone":         timezone,
		"holiday_type":     holidayType,
		"include_nearby":   fmt.Sprintf("%v", includeNearby),
		"nearby_limit":     fmt.Sprintf("%d", nearbyLimit),
		"exclude_past":     "true",
	}
	if date != "" {
		params["date"] = date
	}
	if month != "" {
		params["month"] = month
	}
	if year != "" {
		params["year"] = year
	}
	return c.doGet("/misc/holiday_calendar", params)
}

// GetTarot draws random tarot cards.
func (c *UapisClient) GetTarot(count int) (map[string]any, error) {
	return c.doGet("/misc/tarot", map[string]string{"count": fmt.Sprintf("%d", count)})
}

// GetAnswerBook returns a random answer.
func (c *UapisClient) GetAnswerBook(question string) (map[string]any, error) {
	return c.doGet("/misc/answer_book", map[string]string{"question": question})
}

// AmapClient calls the 高德地图 (AMAP) API.
type AmapClient struct {
	key  string
	http *http.Client
}

// NewAmapClient creates an AmapClient from its env var.
func NewAmapClient() *AmapClient {
	return &AmapClient{
		key: os.Getenv(EnvAmapKey),
		http: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:    5,
				IdleConnTimeout: 60 * time.Second,
			},
		},
	}
}

func (c *AmapClient) doGet(endpoint string, params map[string]string) (map[string]any, error) {
	if params == nil {
		params = map[string]string{}
	}
	params["key"] = c.key
	params["output"] = "json"

	u, _ := url.Parse("https://restapi.amap.com" + endpoint)
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	resp, err := c.http.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("amap request %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("amap decode %s: %w", endpoint, err)
	}
	return result, nil
}

// Geocode converts an address to GCJ-02 coordinates.
func (c *AmapClient) Geocode(address string) (map[string]any, error) {
	return c.doGet("/v3/geocode/geo", map[string]string{"address": address})
}

// ReGeocode converts GCJ-02 coordinates to an address.
func (c *AmapClient) ReGeocode(location string) (map[string]any, error) {
	return c.doGet("/v3/geocode/regeo", map[string]string{"location": location})
}

// NearbySearch searches for POIs near a location.
func (c *AmapClient) NearbySearch(location, keywords, types, radius, offset, page, extensions string) (map[string]any, error) {
	params := map[string]string{
		"location":   location,
		"radius":     radius,
		"offset":     offset,
		"page":       page,
		"extensions": extensions,
	}
	if keywords != "" {
		params["keywords"] = keywords
	}
	if types != "" {
		params["types"] = types
	}
	return c.doGet("/v3/place/around", params)
}

// FuzzySearch searches for POIs by keyword.
func (c *AmapClient) FuzzySearch(keywords, city, offset, page string) (map[string]any, error) {
	return c.doGet("/v3/place/text", map[string]string{
		"keywords": keywords,
		"city":     city,
		"offset":   offset,
		"page":     page,
	})
}

// TransitRoute plans a公交 route.
func (c *AmapClient) TransitRoute(origin, destination, city, strategy, date, time string) (map[string]any, error) {
	params := map[string]string{
		"origin":      origin,
		"destination": destination,
		"city":        city,
		"strategy":    strategy,
	}
	if date != "" {
		params["date"] = date
	}
	if time != "" {
		params["time"] = time
	}
	return c.doGet("/v3/direction/transit/integrated", params)
}

// CyclingRoute plans a cycling route.
func (c *AmapClient) CyclingRoute(origin, destination string) (map[string]any, error) {
	return c.doGet("/v4/direction/cycling", map[string]string{
		"origin":      origin,
		"destination": destination,
	})
}

// TodoistClient wraps the Todoist REST API.
type TodoistClient struct {
	token string
	http  *http.Client
}

// NewTodoistClient creates a TodoistClient from its env var.
func NewTodoistClient() *TodoistClient {
	return &TodoistClient{
		token: os.Getenv(EnvTodoistToken),
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *TodoistClient) doRequest(method, path string, body io.Reader) (map[string]any, error) {
	u := "https://api.todoist.com/rest/v2" + path
	req, _ := http.NewRequest(method, u, body)
	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("todoist %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("todoist decode %s: %w", path, err)
	}
	return result, nil
}

func (c *TodoistClient) doRequestArray(method, path string, body io.Reader) ([]any, error) {
	u := "https://api.todoist.com/rest/v2" + path
	req, _ := http.NewRequest(method, u, body)
	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("todoist %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	var result []any
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("todoist decode array %s: %w", path, err)
	}
	return result, nil
}

// GetProjects lists all projects.
func (c *TodoistClient) GetProjects() ([]any, error) {
	return c.doRequestArray("GET", "/projects", nil)
}

// AddTask creates a new task.
func (c *TodoistClient) AddTask(content, projectID, dueString string, priority int) (map[string]any, error) {
	body := map[string]any{
		"content":  content,
		"priority": priority,
	}
	if projectID != "" {
		body["project_id"] = projectID
	}
	if dueString != "" {
		body["due_string"] = dueString
	}
	b, _ := json.Marshal(body)
	return c.doRequest("POST", "/tasks", strings.NewReader(string(b)))
}

// GetTasks lists tasks (optionally filtered by project).
func (c *TodoistClient) GetTasks(projectID, filter string) ([]any, error) {
	params := url.Values{}
	if projectID != "" {
		params.Set("project_id", projectID)
	}
	if filter != "" {
		params.Set("filter", filter)
	}
	q := params.Encode()
	if q != "" {
		q = "?" + q
	}
	return c.doRequestArray("GET", "/tasks"+q, nil)
}

// CloseTask completes a task.
func (c *TodoistClient) CloseTask(taskID string) error {
	_, err := c.doRequest("POST", "/tasks/"+taskID+"/close", nil)
	return err
}

// DeleteTask deletes a task.
func (c *TodoistClient) DeleteTask(taskID string) error {
	_, err := c.doRequest("DELETE", "/tasks/"+taskID, nil)
	return err
}

// UpdateTask updates a task.
func (c *TodoistClient) UpdateTask(taskID, content, dueString string, priority int) (map[string]any, error) {
	body := map[string]any{}
	if content != "" {
		body["content"] = content
	}
	if dueString != "" {
		body["due_string"] = dueString
	}
	if priority > 0 {
		body["priority"] = priority
	}
	b, _ := json.Marshal(body)
	return c.doRequest("POST", "/tasks/"+taskID, strings.NewReader(string(b)))
}

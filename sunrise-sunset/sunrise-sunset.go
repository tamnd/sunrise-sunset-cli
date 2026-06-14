// Package sunrisesunset is the library behind the sunrise-sunset command line:
// the HTTP client, request shaping, and the typed data models for
// api.sunrise-sunset.org.
//
// The Client is the spine every command shares. It sets a real User-Agent,
// paces requests so a busy session stays polite, and retries the transient
// failures (429 and 5xx) that any public site throws under load.
package sunrisesunset

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DefaultUserAgent identifies the client to the API.
const DefaultUserAgent = "sunrise-sunset-cli/dev (+https://github.com/tamnd/sunrise-sunset-cli)"

// Host is the API hostname.
const Host = "api.sunrise-sunset.org"

// BaseURL is the root every request is built from.
const BaseURL = "https://" + Host

// Client talks to api.sunrise-sunset.org over HTTP.
type Client struct {
	HTTP      *http.Client
	UserAgent string
	// Rate is the minimum gap between requests. Zero means no pacing.
	Rate    time.Duration
	Retries int

	last time.Time
}

// NewClient returns a Client with sensible defaults: a 30s timeout, a 200ms
// minimum gap between requests, and five retries on transient errors.
func NewClient() *Client {
	return &Client{
		HTTP:      &http.Client{Timeout: 30 * time.Second},
		UserAgent: DefaultUserAgent,
		Rate:      200 * time.Millisecond,
		Retries:   5,
	}
}

// Get fetches url and returns the response body. It paces and retries
// according to the client's settings.
func (c *Client) Get(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, url)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", url, lastErr)
}

func (c *Client) do(ctx context.Context, url string) (body []byte, retry bool, err error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

// pace blocks until at least Rate has passed since the previous request.
func (c *Client) pace() {
	if c.Rate <= 0 {
		return
	}
	if wait := c.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}

// SolarInfo holds the sunrise/sunset data for a given location and date.
type SolarInfo struct {
	ID                    string  `kit:"id" json:"id"`
	Latitude              float64 `json:"latitude"`
	Longitude             float64 `json:"longitude"`
	Date                  string  `json:"date"`
	Sunrise               string  `json:"sunrise"`
	Sunset                string  `json:"sunset"`
	SolarNoon             string  `json:"solar_noon"`
	DayLength             string  `json:"day_length"`
	CivilTwilightBegin    string  `json:"civil_twilight_begin"`
	CivilTwilightEnd      string  `json:"civil_twilight_end"`
	NauticalTwilightBegin string  `json:"nautical_twilight_begin"`
	NauticalTwilightEnd   string  `json:"nautical_twilight_end"`
	AstroTwilightBegin    string  `json:"astro_twilight_begin"`
	AstroTwilightEnd      string  `json:"astro_twilight_end"`
	Timezone              string  `json:"timezone"`
}

// apiResponse mirrors the wire format from api.sunrise-sunset.org/json.
type apiResponse struct {
	Results struct {
		Sunrise                   string `json:"sunrise"`
		Sunset                    string `json:"sunset"`
		SolarNoon                 string `json:"solar_noon"`
		DayLength                 string `json:"day_length"`
		CivilTwilightBegin        string `json:"civil_twilight_begin"`
		CivilTwilightEnd          string `json:"civil_twilight_end"`
		NauticalTwilightBegin     string `json:"nautical_twilight_begin"`
		NauticalTwilightEnd       string `json:"nautical_twilight_end"`
		AstronomicalTwilightBegin string `json:"astronomical_twilight_begin"`
		AstronomicalTwilightEnd   string `json:"astronomical_twilight_end"`
	} `json:"results"`
	Status string `json:"status"`
	Tzid   string `json:"tzid"`
}

// GetSolar fetches solar data for the given coordinates and optional date.
// Pass an empty date string to get today's data in UTC.
func (c *Client) GetSolar(ctx context.Context, lat, lng float64, date string) (*SolarInfo, error) {
	url := fmt.Sprintf("%s/json?lat=%g&lng=%g", BaseURL, lat, lng)
	if date != "" {
		url += "&date=" + date
	}

	body, err := c.Get(ctx, url)
	if err != nil {
		return nil, err
	}

	var resp apiResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode solar response: %w", err)
	}
	if resp.Status != "OK" {
		return nil, fmt.Errorf("api error: status %q", resp.Status)
	}

	d := date
	if d == "" {
		d = "today"
	}

	return &SolarInfo{
		ID:                    fmt.Sprintf("%g,%g,%s", lat, lng, d),
		Latitude:              lat,
		Longitude:             lng,
		Date:                  d,
		Sunrise:               resp.Results.Sunrise,
		Sunset:                resp.Results.Sunset,
		SolarNoon:             resp.Results.SolarNoon,
		DayLength:             resp.Results.DayLength,
		CivilTwilightBegin:    resp.Results.CivilTwilightBegin,
		CivilTwilightEnd:      resp.Results.CivilTwilightEnd,
		NauticalTwilightBegin: resp.Results.NauticalTwilightBegin,
		NauticalTwilightEnd:   resp.Results.NauticalTwilightEnd,
		AstroTwilightBegin:    resp.Results.AstronomicalTwilightBegin,
		AstroTwilightEnd:      resp.Results.AstronomicalTwilightEnd,
		Timezone:              resp.Tzid,
	}, nil
}

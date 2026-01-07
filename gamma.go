package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	GammaAPIBase = "https://gamma-api.polymarket.com"
)

// GammaClient handles requests to the Polymarket Gamma API
type GammaClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewGammaClient creates a new Gamma API client
func NewGammaClient() *GammaClient {
	return &GammaClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: GammaAPIBase,
	}
}

// FetchMarkets retrieves active markets from the Gamma API
func (c *GammaClient) FetchMarkets(limit int, offset int) ([]Market, error) {
	params := url.Values{}
	params.Set("active", "true")
	params.Set("closed", "false")
	params.Set("limit", fmt.Sprint(limit))
	params.Set("offset", fmt.Sprint(offset))
	params.Set("order", "volume") // Sort by volume desc
	params.Set("ascending", "false")

	reqURL := fmt.Sprintf("%s/markets?%s", c.baseURL, params.Encode())
	
	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch markets: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse raw gamma markets
	var gammaMarkets []GammaMarket
	if err := json.Unmarshal(body, &gammaMarkets); err != nil {
		return nil, fmt.Errorf("failed to parse markets: %w", err)
	}

	// Convert to our Market type and filter expired
	markets := make([]Market, 0, len(gammaMarkets))
	now := time.Now()
	for _, gm := range gammaMarkets {
		market := c.convertGammaMarket(gm)
		
		// Skip if end date is in the past (allow 24h grace period for resolving markets)
		if !market.EndDate.IsZero() && market.EndDate.Before(now.AddDate(0, 0, -1)) {
			continue
		}

		markets = append(markets, market)
	}

	return markets, nil
}

// FetchMarketByID retrieves a specific market by its ID
func (c *GammaClient) FetchMarketByID(marketID string) (*Market, error) {
	reqURL := fmt.Sprintf("%s/markets/%s", c.baseURL, marketID)
	
	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch market: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var gm GammaMarket
	if err := json.Unmarshal(body, &gm); err != nil {
		return nil, fmt.Errorf("failed to parse market: %w", err)
	}

	market := c.convertGammaMarket(gm)
	return &market, nil
}

// convertGammaMarket converts a GammaMarket to our internal Market type
func (c *GammaClient) convertGammaMarket(gm GammaMarket) Market {
	market := Market{
		ID:          gm.ID,
		ConditionID: gm.ConditionID,
		Question:    gm.Question,
		Description: gm.Description,
		Volume:      gm.Volume,
		Liquidity:   gm.Liquidity,
		Active:      gm.Active && !gm.Closed,
	}

	// Parse outcomes JSON array
	if gm.Outcomes != "" {
		var outcomes []string
		if err := json.Unmarshal([]byte(gm.Outcomes), &outcomes); err == nil {
			market.Outcomes = outcomes
		}
	}

	// Parse outcome prices JSON array
	if gm.OutcomePrices != "" {
		var prices []string
		if err := json.Unmarshal([]byte(gm.OutcomePrices), &prices); err == nil {
			market.OutcomeTokens = prices
		}
	}

	// Parse end date
	if gm.EndDateISO != "" {
		// Try multiple formats
		formats := []string{
			time.RFC3339,
			"2006-01-02",
			"2006-01-02T15:04:05",
		}
		
		for _, format := range formats {
			if t, err := time.Parse(format, gm.EndDateISO); err == nil {
				market.EndDate = t
				break
			}
		}
	}

	return market
}

// SearchMarkets searches for markets matching a query using the public-search endpoint
func (c *GammaClient) SearchMarkets(query string, limit int) ([]Market, error) {
	params := url.Values{}
	params.Set("q", query) // Search query parameter

	reqURL := fmt.Sprintf("%s/public-search?%s", c.baseURL, params.Encode())

	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to search markets: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var searchResp SearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	// Extract markets from events
	markets := make([]Market, 0)
	now := time.Now()
	for _, event := range searchResp.Events {
		for _, gm := range event.Markets {
			// Only include active, non-closed markets
			if gm.Active && !gm.Closed {
				market := c.convertGammaMarket(gm)
				
				// Skip if end date is in the past
				if !market.EndDate.IsZero() && market.EndDate.Before(now.AddDate(0, 0, -1)) {
					continue
				}

				markets = append(markets, market)
			}
		}
		// Limit results
		if len(markets) >= limit {
			markets = markets[:limit]
			break
		}
	}

	return markets, nil
}

package main

import "time"

// Market represents a Polymarket prediction market
type Market struct {
	ID           string    `json:"id"`
	ConditionID  string    `json:"conditionId"`
	Question     string    `json:"question"`
	Description  string    `json:"description"`
	Outcomes     []string  `json:"outcomes"`
	OutcomeTokens []string `json:"outcomePrices"` // Token prices as strings (e.g., "0.65")
	EndDate      time.Time `json:"endDate"`
	Volume       string    `json:"volume"`
	Liquidity    string    `json:"liquidity"`
	Active       bool      `json:"active"`
}

// GammaMarket is the raw response from Gamma API (maps directly to their schema)
type GammaMarket struct {
	ID            string `json:"id"`
	ConditionID   string `json:"conditionId"`
	Question      string `json:"question"`
	Description   string `json:"description"`
	Outcomes      string `json:"outcomes"`      // JSON array as string, need to parse
	OutcomePrices string `json:"outcomePrices"` // JSON array as string
	EndDateISO    string `json:"endDateIso"`
	Volume        string `json:"volume"`
	Liquidity     string `json:"liquidity"`
	Active        bool   `json:"active"`
	Closed        bool   `json:"closed"`
}

// GammaEvent represents an event from the search API
type GammaEvent struct {
	ID      string        `json:"id"`
	Title   string        `json:"title"`
	Slug    string        `json:"slug"`
	Active  bool          `json:"active"`
	Closed  bool          `json:"closed"`
	Markets []GammaMarket `json:"markets"`
}

// SearchResponse represents the response from /public-search
type SearchResponse struct {
	Events []GammaEvent `json:"events"`
}

// Position represents a user's position on a market outcome
type Position struct {
	ID      string `json:"id"`
	User    string `json:"user"`
	Outcome string `json:"outcome"` // Outcome index or name
	Market  string `json:"market"`
	Size    string `json:"size"`
}

// OutcomeStats contains statistics for a single outcome
type OutcomeStats struct {
	Outcome      string  `json:"outcome"`
	OutcomeIndex int     `json:"outcomeIndex"`
	UserCount    int     `json:"userCount"`
	Percentage   float64 `json:"percentage"`
	Price        string  `json:"price"` // Current price/probability
}

// MarketStats contains aggregated statistics for a market
type MarketStats struct {
	MarketID       string         `json:"marketId"`
	Question       string         `json:"question"`
	TotalUsers     int            `json:"totalUsers"`
	OutcomeStats   []OutcomeStats `json:"outcomeStats"`
	PopularOutcome string         `json:"popularOutcome"` // Outcome with highest user percentage
	PopularPct     float64        `json:"popularPct"`
}

// APIResponse is a generic wrapper for API responses
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

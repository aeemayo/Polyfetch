package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// Polymarket PnL Subgraph endpoint on The Graph
	SubgraphURL = "https://api.thegraph.com/subgraphs/name/polymarket/pnl-subgraph"
	// Alternative endpoints to try if the main one fails
	SubgraphURLAlt = "https://api.thegraph.com/subgraphs/name/polymarket/polymarket-matic"
)

// SubgraphClient handles GraphQL queries to The Graph
type SubgraphClient struct {
	httpClient *http.Client
	endpoint   string
}

// GraphQLRequest represents a GraphQL query request
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

// PositionData is the subgraph position entity
type PositionData struct {
	ID          string `json:"id"`
	User        struct {
		ID string `json:"id"`
	} `json:"user"`
	Outcome     string `json:"outcome"`
	Market      struct {
		ID string `json:"id"`
	} `json:"market"`
	QuantityBought string `json:"quantityBought"`
	QuantitySold   string `json:"quantitySold"`
}

// NewSubgraphClient creates a new SubgraphClient
func NewSubgraphClient() *SubgraphClient {
	return &SubgraphClient{
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // Longer timeout for potentially large queries
		},
		endpoint: SubgraphURL,
	}
}

// Query executes a GraphQL query
func (c *SubgraphClient) Query(query string, variables map[string]interface{}) (json.RawMessage, error) {
	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	resp, err := c.httpClient.Post(c.endpoint, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var gqlResp GraphQLResponse
	if err := json.Unmarshal(body, &gqlResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL error: %s", gqlResp.Errors[0].Message)
	}

	return gqlResp.Data, nil
}

// GetMarketPositions fetches all positions for a specific market with pagination
func (c *SubgraphClient) GetMarketPositions(marketID string) ([]PositionData, error) {
	var allPositions []PositionData
	skip := 0
	limit := 1000 // The Graph max per query

	query := `
		query GetPositions($marketId: String!, $first: Int!, $skip: Int!) {
			positions(
				first: $first
				skip: $skip
				where: { market: $marketId }
			) {
				id
				user {
					id
				}
				outcome
				market {
					id
				}
				quantityBought
				quantitySold
			}
		}
	`

	for {
		variables := map[string]interface{}{
			"marketId": marketID,
			"first":    limit,
			"skip":     skip,
		}

		data, err := c.Query(query, variables)
		if err != nil {
			return nil, fmt.Errorf("failed to query positions (skip=%d): %w", skip, err)
		}

		var result struct {
			Positions []PositionData `json:"positions"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("failed to parse positions: %w", err)
		}

		allPositions = append(allPositions, result.Positions...)

		// Break if we got fewer results than the limit (no more pages)
		if len(result.Positions) < limit {
			break
		}
		skip += limit

		// Safety limit to prevent infinite loops
		if skip > 10000 {
			break
		}
	}

	return allPositions, nil
}

// ComputeMarketStats calculates statistics based on market prices (Probability)
// Since the user positions subgraph is deprecated, we prioritize price/consensus data.
func (c *SubgraphClient) ComputeMarketStats(market *Market) (*MarketStats, error) {
	// Directly use price-based stats as the primary source of truth
	return c.computeStatsFromPrices(market), nil
}

// computeStatsFromPrices creates stats based on market prices
func (c *SubgraphClient) computeStatsFromPrices(market *Market) *MarketStats {
	stats := &MarketStats{
		MarketID:   market.ID,
		Question:   market.Question,
		TotalUsers: 0, // Not available via public API
	}

	var maxPct float64
	var popularOutcome string

	for i, outcomeName := range market.Outcomes {
		price := "0"
		if i < len(market.OutcomeTokens) {
			price = market.OutcomeTokens[i]
		}

		var priceFloat float64
		fmt.Sscanf(price, "%f", &priceFloat)
		
		// Price represents the probability (0-1), convert to percentage
		pct := priceFloat * 100

		stats.OutcomeStats = append(stats.OutcomeStats, OutcomeStats{
			Outcome:      outcomeName,
			OutcomeIndex: i,
			UserCount:    0, // Not available
			Percentage:   pct,
			Price:        price,
		})

		if pct > maxPct {
			maxPct = pct
			popularOutcome = outcomeName
		}
	}

	stats.PopularOutcome = popularOutcome
	stats.PopularPct = maxPct

	return stats
}

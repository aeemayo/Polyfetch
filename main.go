package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// Server holds the application dependencies
type Server struct {
	gammaClient    *GammaClient
	subgraphClient *SubgraphClient
	router         *mux.Router
}

// NewServer creates a new server instance
func NewServer() *Server {
	s := &Server{
		gammaClient:    NewGammaClient(),
		subgraphClient: NewSubgraphClient(),
		router:         mux.NewRouter(),
	}
	s.setupRoutes()
	return s
}

// setupRoutes configures the API routes
func (s *Server) setupRoutes() {
	s.router.HandleFunc("/api/health", s.handleHealth).Methods("GET")
	s.router.HandleFunc("/api/markets", s.handleGetMarkets).Methods("GET")
	s.router.HandleFunc("/api/markets/search", s.handleSearchMarkets).Methods("GET")
	s.router.HandleFunc("/api/market/{id}", s.handleGetMarket).Methods("GET")
	s.router.HandleFunc("/api/market/{id}/stats", s.handleGetMarketStats).Methods("GET")
}

// handleHealth returns server health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    map[string]string{"status": "healthy"},
	})
}

// handleGetMarkets returns a list of active markets
func (s *Server) handleGetMarkets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse query params
	limit := 50
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	markets, err := s.gammaClient.FetchMarkets(limit, offset)
	if err != nil {
		log.Printf("Error fetching markets: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    markets,
	})
}

// handleSearchMarkets searches for markets matching a query
func (s *Server) handleSearchMarkets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query().Get("q")
	if query == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   "query parameter 'q' is required",
		})
		return
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 50 {
			limit = parsed
		}
	}

	markets, err := s.gammaClient.SearchMarkets(query, limit)
	if err != nil {
		log.Printf("Error searching markets: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    markets,
	})
}

// handleGetMarket returns details for a specific market
func (s *Server) handleGetMarket(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	vars := mux.Vars(r)
	marketID := vars["id"]

	market, err := s.gammaClient.FetchMarketByID(marketID)
	if err != nil {
		log.Printf("Error fetching market %s: %v", marketID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    market,
	})
}

// handleGetMarketStats returns betting statistics for a market
func (s *Server) handleGetMarketStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	vars := mux.Vars(r)
	marketID := vars["id"]

	// First fetch the market details
	market, err := s.gammaClient.FetchMarketByID(marketID)
	if err != nil {
		log.Printf("Error fetching market %s: %v", marketID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Compute statistics
	stats, err := s.subgraphClient.ComputeMarketStats(market)
	if err != nil {
		log.Printf("Error computing stats for market %s: %v", marketID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    stats,
	})
}

func main() {
	log.Println("Starting Polyfetch server...")
	
	server := NewServer()

	// Setup CORS for frontend access
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173", "*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(server.router)

	port := ":8080"
	log.Printf("Server listening on %s", port)
	log.Printf("API endpoints:")
	log.Printf("  GET /api/health - Health check")
	log.Printf("  GET /api/markets - List active markets")
	log.Printf("  GET /api/markets/search?q=query - Search markets")
	log.Printf("  GET /api/market/{id} - Get market details")
	log.Printf("  GET /api/market/{id}/stats - Get market betting stats")

	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

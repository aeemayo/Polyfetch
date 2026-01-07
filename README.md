# Polyfetch - Polymarket Analysis App

A real-time Polymarket analysis application that identifies outcomes with high percentages of users betting on them.

## Features

- 📊 **Real-time Market Analysis**: Fetch active markets from Polymarket's Gamma API
- 👥 **User Distribution Stats**: See how many unique users are betting on each outcome
- 🔍 **Search Markets**: Filter markets by keywords
- 📈 **Popularity Tracking**: Identify the most popular betting outcomes
- 🎨 **Premium Dark UI**: Modern, responsive design with smooth animations

## Tech Stack

- **Backend**: Go with gorilla/mux router
- **Frontend**: React + Vite with Material-UI
- **APIs**: 
  - Polymarket Gamma API (market data)
  - The Graph Subgraph (position/betting data)


## Getting Started

### Prerequisites

- Go 1.21+
- Node.js 18+
- npm

### Running the Backend

```bash
# From the Polyfetch directory
go run .
# Or build and run
go build -o polyfetch.exe .
./polyfetch.exe
```

The backend will start on `http://localhost:8080`

### Running the Frontend

```bash
cd frontend
npm install
npm run dev
```

The frontend will start on `http://localhost:5173`

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/health` | GET | Health check |
| `/api/markets` | GET | List active markets (params: `limit`, `offset`) |
| `/api/markets/search` | GET | Search markets (params: `q`, `limit`) |
| `/api/market/{id}` | GET | Get market details |
| `/api/market/{id}/stats` | GET | Get betting statistics for a market |

## Example Response


## Notes

- The subgraph queries may fall back to price-based statistics if position data is unavailable
- Market prices represent probabilities (0-1)
- High-volume markets may require pagination (currently limited to 10,000 positions)

## License

MIT

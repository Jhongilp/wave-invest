# Wave Invest - Real-Time Market Movers

A real-time stock market dashboard displaying the **Top 20 Gainers** with live price updates via WebSocket push architecture.

## Architecture

```
┌─────────────────┐      WebSocket       ┌─────────────────┐
│   React + Vite  │ ◄──────────────────► │  Go/Gin Server  │
│    (Frontend)   │      Push Updates    │    (Backend)    │
└─────────────────┘                      └────────┬────────┘
                                                  │
                                         Poll every 30s
                                                  │
                                                  ▼
                                         ┌─────────────────┐
                                         │   Polygon.io    │
                                         │   Snapshot API  │
                                         └─────────────────┘
```

## Tech Stack

| Layer | Technology |
|-------|------------|
| Frontend | React 18 + TypeScript + Vite + Tailwind CSS |
| Backend | Go + Gin + gorilla/websocket |
| Data Source | Polygon.io (Massive.com) Snapshot API |
| Communication | WebSocket (server → client push) |

## Quick Start

### Prerequisites
- Go 1.21+
- Node.js 18+
- Polygon.io API key (free tier works)

### 1. Start the Backend

```bash
cd server

# Copy environment config
cp .env.example .env

# Add your Polygon API key to .env (optional - runs in demo mode without it)
# POLYGON_API_KEY=your_key_here

# Run the server
go run cmd/api/main.go
```

Server starts at `http://localhost:8080`

### 2. Start the Frontend

```bash
cd client

# Install dependencies
npm install

# Start dev server
npm run dev
```

Frontend starts at `http://localhost:5173`

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/api/gainers` | GET | Get current top 20 gainers (REST fallback) |
| `/ws` | WS | WebSocket connection for real-time updates |

## WebSocket Message Format

```json
{
  "type": "gainers",
  "data": [
    {
      "ticker": "AAPL",
      "price": 185.23,
      "changePercent": 5.42,
      "change": 9.52,
      "volume": 45000000,
      "previousClose": 175.71,
      "open": 176.50,
      "high": 186.00,
      "low": 175.20,
      "vwap": 181.50,
      "updatedAt": 1709510400000000000
    }
  ],
  "timestamp": "2026-03-03T15:30:00Z"
}
```

## Demo Mode

If no `POLYGON_API_KEY` is set, the server runs in **demo mode** with simulated market data that updates every 30 seconds. This is useful for development and testing.

## Configuration

### Backend (`server/.env`)

| Variable | Default | Description |
|----------|---------|-------------|
| `POLYGON_API_KEY` | - | Your Polygon.io API key |
| `PORT` | 8080 | Server port |
| `POLL_INTERVAL_SECONDS` | 30 | How often to fetch from Polygon |
| `ALLOWED_ORIGINS` | localhost:5173 | CORS allowed origins |

### Frontend (`client/.env`)

| Variable | Default | Description |
|----------|---------|-------------|
| `VITE_WS_URL` | ws://localhost:8080/ws | WebSocket server URL |

## Features

- ✅ Real-time price updates via WebSocket
- ✅ Green/red flash animations on price changes
- ✅ Auto-reconnect with exponential backoff
- ✅ Loading skeletons while connecting
- ✅ Connection status indicator
- ✅ Demo mode for testing without API key
- ✅ Responsive design with Tailwind CSS

## Project Structure

```
wave_invest/
├── client/                    # React frontend
│   ├── src/
│   │   ├── components/        # UI components
│   │   ├── hooks/             # Custom React hooks
│   │   └── types/             # TypeScript types
│   └── package.json
│
└── server/                    # Go backend
    ├── cmd/api/               # Application entrypoint
    └── internal/
        ├── config/            # Configuration
        ├── handlers/          # HTTP & WebSocket handlers
        ├── models/            # Data models
        └── services/          # Business logic & Polygon client
```

## License

MIT

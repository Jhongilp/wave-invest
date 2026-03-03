# Wave Invest - AI Agent Context

## Project Overview

Real-time stock market dashboard displaying **Top 20 Gainers** with live WebSocket push updates. Built March 2026.

## Architecture

```
React (Vite) ◄──WebSocket──► Go/Gin Server ──Poll 30s──► Polygon.io API
```

**Push Model**: Server polls Polygon every 30s, calculates top gainers, broadcasts to all WebSocket clients. Clients never poll—they receive pushed updates.

## Tech Stack

| Component | Technology | Version |
|-----------|------------|---------|
| Frontend | React + TypeScript + Vite | React 18, Vite 7 |
| Styling | Tailwind CSS | v4 (via @tailwindcss/vite plugin) |
| Backend | Go + Gin | Go 1.21+ |
| WebSocket | gorilla/websocket | v1.5 |
| Market Data | polygon-io/client-go | v1.16 (deprecated, rebranded to Massive.com) |

## Directory Structure

```
wave_invest/
├── client/                     # React frontend
│   ├── src/
│   │   ├── components/
│   │   │   ├── GainersTable.tsx    # Main table component
│   │   │   ├── GainerRow.tsx       # Individual row with animations
│   │   │   └── ConnectionStatus.tsx # Live/disconnected indicator
│   │   ├── hooks/
│   │   │   └── useGainersSocket.ts # WebSocket connection hook
│   │   ├── types/
│   │   │   └── market.ts           # TypeScript interfaces
│   │   ├── App.tsx
│   │   └── index.css               # Tailwind imports + custom animations
│   ├── vite.config.ts
│   └── package.json
│
├── server/                     # Go backend
│   ├── cmd/api/
│   │   └── main.go                 # Entry point, router setup
│   └── internal/
│       ├── config/
│       │   └── config.go           # Viper-based config from env
│       ├── handlers/
│       │   ├── websocket.go        # Hub + Client, broadcast logic
│       │   └── gainers.go          # REST endpoint handler
│       ├── models/
│       │   └── market.go           # Gainer, TickerSnapshot structs
│       └── services/
│           ├── poller.go           # Background polling goroutine
│           └── polygon/
│               ├── client.go       # Polygon SDK wrapper
│               └── mock.go         # Demo mode mock data
│
├── README.md                   # User documentation
└── CLAUDE.md                   # This file (AI context)
```

## Key Files

| File | Purpose |
|------|---------|
| `server/cmd/api/main.go` | Server entrypoint, Gin routes, graceful shutdown |
| `server/internal/handlers/websocket.go` | WebSocket hub managing client connections |
| `server/internal/services/poller.go` | Background goroutine polling Polygon API |
| `client/src/hooks/useGainersSocket.ts` | WebSocket hook with auto-reconnect |
| `client/src/components/GainerRow.tsx` | Row with green/red flash animation on price change |

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/api/gainers` | GET | REST fallback for current gainers |
| `/ws` | WebSocket | Real-time updates (primary) |

## WebSocket Message Format

```json
{
  "type": "gainers",
  "data": [{ "ticker": "AAPL", "price": 185.23, "changePercent": 5.42, ... }],
  "timestamp": "2026-03-03T15:30:00Z"
}
```

## Environment Variables

### Backend (`server/.env`)
- `POLYGON_API_KEY` - Polygon.io API key (optional, runs demo mode without)
- `PORT` - Server port (default: 8080)
- `POLL_INTERVAL_SECONDS` - Polling frequency (default: 30)
- `ALLOWED_ORIGINS` - CORS origins (default: localhost:5173)

### Frontend (`client/.env`)
- `VITE_WS_URL` - WebSocket URL (default: ws://localhost:8080/ws)

## Development Commands

```bash
# Backend
cd server
go run cmd/api/main.go          # Start server
go build ./...                   # Verify compilation
go mod tidy                      # Sync dependencies

# Frontend
cd client
npm install                      # Install deps
npm run dev                      # Dev server (port 5173)
npm run build                    # Production build
```

## Important Patterns

### TypeScript Imports
Use `import type { X }` for type-only imports (verbatimModuleSyntax enabled):
```typescript
import type { Gainer } from '../types/market';
```

### WebSocket Hub Pattern (Go)
- `Hub` struct manages all clients via channels
- `register`/`unregister` channels for connection lifecycle
- `broadcast` channel for sending to all clients
- Thread-safe with `sync.RWMutex`

### Price Animation (React)
`GainerRow` tracks previous price in `useRef`, triggers CSS animation class on change:
```tsx
setAnimationClass(isUp ? 'flash-green' : 'flash-red');
```

## Known Issues / Notes

1. **Polygon SDK Deprecation**: `polygon-io/client-go` is deprecated in favor of `massive-com/client-go`, but the new package has import path issues as of March 2026. Continue using polygon-io for now.

2. **Demo Mode**: When `POLYGON_API_KEY` is empty, server generates mock data with simulated price fluctuations.

3. **Tailwind v4**: Uses new `@tailwindcss/vite` plugin, CSS import is just `@import "tailwindcss";`

4. **VWAP Field**: Polygon snapshot API doesn't include VWAP in DaySnapshot, field is set to 0.

## Common Tasks

### Add a new column to the table
1. Update `models/market.go` (Go struct)
2. Update `types/market.ts` (TS interface)
3. Update `GainerRow.tsx` (add td)
4. Update `GainersTable.tsx` (add th + skeleton td)

### Change polling interval
Set `POLL_INTERVAL_SECONDS` in `server/.env`

### Add new WebSocket message type
1. Define type in `models/market.go`
2. Handle in `poller.go` broadcast
3. Parse in `useGainersSocket.ts` onmessage handler

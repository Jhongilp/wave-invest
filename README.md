# 🌊 Wave Invest

A swing trading assistant that uses AI to generate trading plans for stocks.

## Tech Stack

- **Client**: React 19, Vite, TypeScript, Tailwind CSS
- **Server**: Go, Chi router
- **AI**: Gemini API
- **Broker**: eToro API

## Project Structure

```
wave_invest/
├── client/                 # React frontend
│   └── src/
│       ├── components/     # UI components
│       │   ├── Layout/     # Dashboard layout
│       │   ├── Watchlist/  # Ticker list sidebar
│       │   └── TradingPlan/# AI analysis display
│       ├── hooks/          # useWatchlist, useTradingPlan
│       ├── types/          # TypeScript interfaces
│       └── lib/            # API client
├── server/                 # Go backend
│   ├── cmd/api/            # Entry point
│   ├── internal/
│   │   ├── handlers/       # HTTP handlers
│   │   ├── models/         # Data structures
│   │   └── services/       # Business logic
│   ├── pkg/
│   │   ├── etoro/          # eToro API client
│   │   └── gemini/         # Gemini AI client
│   └── config/             # Configuration
└── PLAN.md                 # Development roadmap
```

## Getting Started

### Prerequisites

- Go 1.23+
- Node.js 20+
- eToro API credentials
- Gemini API key

### Server Setup

```bash
cd server
cp .env.example .env
# Edit .env with your API keys
go run ./cmd/api
```

Server runs on http://localhost:8080

### Client Setup

```bash
cd client
npm install
npm run dev
```

Client runs on http://localhost:5173

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/api/watchlist` | Get eToro watchlist |
| POST | `/api/analyze/:ticker` | Generate trading plan |
| POST | `/api/analyze/batch` | Analyze multiple tickers |
| GET | `/api/plan/:ticker` | Get cached trading plan |

## Features

- **Watchlist View**: Display tickers from eToro with price/change info
- **AI Trading Plans**: Gemini-powered analysis including:
  - Moving averages (MA20, MA50, MA200)
  - RSI indicator
  - Support/Resistance levels
  - Entry zones & stop loss
  - Price targets (PT1, PT2, PT3)
  - Risk/Reward ratio
  - Sentiment analysis

## Environment Variables

```
PORT=8080
ETORO_API_KEY=your_key
ETORO_API_SECRET=your_secret
GEMINI_API_KEY=your_key
```

## License

MIT

# 🌊 Wave Invest

A swing trading assistant that uses AI to generate trading plans and execute trades autonomously.

## Tech Stack

- **Client**: React 19, Vite, TypeScript, Tailwind CSS
- **Server**: Go, Chi router
- **AI**: Gemini API (gemini-3-flash-preview with Google Search grounding)
- **Broker**: eToro API
- **Database**: Cloud Firestore

## Project Structure

```
wave_invest/
├── client/                 # React frontend
│   └── src/
│       ├── components/
│       │   ├── Layout/         # Dashboard layout with navigation
│       │   ├── Watchlist/      # Ticker list sidebar
│       │   ├── TradingPlan/    # AI analysis display
│       │   ├── Opportunities/  # Scored opportunities view
│       │   ├── Portfolio/      # Portfolio & positions
│       │   └── Settings/       # Trading rules config
│       ├── hooks/              # React hooks
│       ├── types/              # TypeScript interfaces
│       └── lib/                # API client
├── server/                 # Go backend
│   ├── cmd/api/            # Entry point
│   ├── internal/
│   │   ├── handlers/       # HTTP handlers
│   │   ├── models/         # Data structures
│   │   └── services/       # Business logic
│   │       ├── analyzer.go         # Single ticker analysis
│   │       ├── scorer.go           # Opportunity scoring
│   │       ├── daily_analysis.go   # Batch analysis orchestration
│   │       ├── executor.go         # Trade execution
│   │       ├── firestore.go        # Firestore CRUD
│   │       └── position_sizer.go   # Position sizing
│   ├── pkg/
│   │   ├── etoro/          # eToro API client
│   │   ├── gemini/         # Gemini AI client
│   │   └── firestore/      # Firestore client
│   └── config/             # Configuration
├── PLAN.md                 # Development roadmap
└── DEVLOG.md               # Build log / journal
```

## Getting Started

### Prerequisites

- Go 1.23+
- Node.js 20+
- eToro API credentials
- Gemini API key
- Google Cloud project with Firestore enabled

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

### Phase 1: Analysis

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/api/watchlist` | Get eToro watchlist |
| POST | `/api/analyze/{ticker}` | Generate trading plan for ticker |
| POST | `/api/analyze/batch` | Analyze multiple tickers |
| GET | `/api/plan/{ticker}` | Get cached trading plan |

### Phase 2: Autonomous Trading

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/daily-analysis` | Run batch analysis on entire watchlist |
| GET | `/api/opportunities` | Get today's scored opportunities |
| GET | `/api/opportunities/{date}` | Get opportunities for specific date |
| GET | `/api/portfolio` | Get portfolio summary |
| POST | `/api/portfolio` | Create portfolio with initial budget |
| GET | `/api/settings` | Get trading rules |
| PUT | `/api/settings` | Update trading rules |
| POST | `/api/execute-trades` | Execute trades for top opportunities |
| GET | `/api/positions` | Get open positions |
| DELETE | `/api/positions/{id}` | Close a position |
| GET | `/api/transactions` | Get transaction history |

## Features

### Phase 1: Trading Plans ✅
- **Watchlist View**: Display tickers from eToro with price/change info
- **AI Trading Plans**: Gemini-powered analysis including:
  - Moving averages (MA20, MA50, MA200)
  - RSI indicator
  - Support/Resistance levels
  - Entry zones & stop loss
  - Price targets (PT1, PT2, PT3)
  - Risk/Reward ratio
  - Sentiment analysis

### Phase 2: Autonomous Trading ✅
- **Daily Analysis**: Batch analyze entire watchlist with one click
- **Scoring Algorithm**: Rank opportunities by trade quality
  - R/R ratio: 40% weight
  - Sentiment: 25% weight
  - RSI position: 20% weight
  - Bias alignment: 15% weight
- **Portfolio Management**: Track budget, positions, and P&L
- **Trading Rules**: Configurable parameters
  - Max position size (% of portfolio)
  - Max concurrent positions
  - Minimum score threshold
  - Daily loss limit
- **Auto Execution**: Execute top-scored trades on eToro
- **Firestore Persistence**: 7-day retention for analysis history

### Phase 3: Full Autonomy (Planned)
- Market scanning (no watchlist required)
- Scheduled auto-analysis (cron)
- Notification system
- Performance analytics & backtesting

## Environment Variables

```bash
# Server
PORT=8080
ETORO_API_KEY=your_etoro_key
ETORO_API_SECRET=your_etoro_secret
GEMINI_API_KEY=your_gemini_key

# Firestore
GOOGLE_CLOUD_PROJECT=your_project_id
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
```

## Scoring Algorithm

Opportunities are scored 0-100 based on:

| Factor | Weight | Description |
|--------|--------|-------------|
| R/R Ratio | 40% | Higher risk/reward = higher score (capped at 5:1) |
| Sentiment | 25% | Positive sentiment for longs, negative for shorts |
| RSI | 20% | Oversold for longs (<30), overbought for shorts (>70) |
| Bias | 15% | Clear bullish/bearish bias vs neutral |

**Filters**: Neutral bias, unfavorable RSI, and R/R < 1:1 are excluded.

## License

MIT

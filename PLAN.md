This will be a swing trading app

Stack:

- client: React, Vite, Tailwind
- server: Go, Gemini AI
- broker api: etoro api

---

## Phase 1: Setting Trading Plan (Detailed)

### 1.1 Server Setup (Go)
- [ ] Initialize Go module (`go mod init wave_invest`)
- [ ] Set up project structure:
  ```
  server/
    cmd/api/main.go          # Entry point
    internal/
      handlers/              # HTTP handlers
      services/              # Business logic
      models/                # Data structures
    pkg/
      etoro/                 # eToro API client
      gemini/                # Gemini AI client
    config/                  # Configuration management
  ```
- [ ] Create basic HTTP server with Chi or Gin router
- [ ] Set up environment config (API keys, ports)

### 1.2 eToro Integration
- [ ] Research eToro API authentication (OAuth2/API keys)
- [ ] Implement eToro client:
  - [ ] `GetWatchlist()` - Fetch user's watchlist tickers
  - [ ] `GetTickerData(symbol)` - Fetch price history, volume
- [ ] Add rate limiting and error handling

### 1.3 Gemini AI Integration 
- [ ] Set up Gemini API client (using "gemini-3-flash-preview" model)
- [ ] Design prompt engineering for swing trading analysis:
  - Technical indicators (MA, RSI, Volume Profile)
  - Key price levels (support/resistance)
  - Entry points & price targets
  - Risk/reward calculation
  - Sentiment analysis
- [ ] Create structured output format (JSON schema)
- [ ] Implement `GenerateTradingPlan(tickerData)` service

### 1.4 API Endpoints
- [ ] `GET /api/watchlist` - Fetch eToro watchlist
- [ ] `POST /api/analyze/:ticker` - Generate trading plan for single ticker
- [ ] `POST /api/analyze/batch` - Generate plans for multiple tickers
- [ ] `GET /api/plan/:ticker` - Get cached trading plan

### 1.5 Client UI
- [ ] **Layout**: Dark theme dashboard with sidebar navigation
- [ ] **Watchlist View**:
  - [ ] Display tickers from eToro watchlist
  - [ ] Show basic info (price, daily change %)
  - [ ] "Analyze" button per ticker
- [ ] **Trading Plan View**:
  - [ ] Display AI-generated analysis:
    - Moving Averages chart/indicators
    - RSI gauge
    - Volume profile visualization
    - Support/Resistance levels
    - Entry/Exit points with PT1, PT2, PT3
    - Risk/Reward ratio display
    - Sentiment indicator
  - [ ] Loading states for AI analysis
- [ ] **API Integration**:
  - [ ] Custom hooks: `useWatchlist()`, `useTradingPlan()`
  - [ ] Error handling and retry logic

### 1.6 Data Types (TypeScript)
```typescript
interface Ticker {
  symbol: string;
  name: string;
  price: number;
  change: number;
  changePercent: number;
}

interface TradingPlan {
  ticker: string;
  analyzedAt: string;
  technicals: {
    ma20: number;
    ma50: number;
    ma200: number;
    rsi: number;
    volumeProfile: VolumeLevel[];
  };
  levels: {
    support: number[];
    resistance: number[];
  };
  trade: {
    bias: 'bullish' | 'bearish' | 'neutral';
    entryZone: { low: number; high: number };
    stopLoss: number;
    targets: { pt1: number; pt2: number; pt3: number };
    riskRewardRatio: number;
  };
  sentiment: {
    score: number; // -100 to 100
    institutionalFlow: string;
    smartMoneyBets: string[];
  };
  summary: string;
}
```

### 1.7 Milestones
| Milestone | Deliverable | Est. Time |
|-----------|-------------|-----------|
| M1 | Go server boilerplate + health endpoint | 2 hrs |
| M2 | eToro API integration (watchlist) | 4 hrs |
| M3 | Gemini AI prompt + trading plan generation | 4 hrs |
| M4 | API endpoints complete | 2 hrs |
| M5 | Client watchlist UI | 3 hrs |
| M6 | Client trading plan view | 4 hrs |
| M7 | Integration & testing | 3 hrs |

**Total Estimated Time: ~22 hours**

---

## Phase 2: Trading (Future)

---

## Phase 2: Autonomous Trading with Prioritization (Detailed)

### Overview
Add a scoring system to rank watchlist tickers by trade quality, implement budget management with position sizing, and track all transactions in Firestore for autonomous trading execution.

### 2.1 Scoring & Prioritization Engine
- [ ] Create `ScoredOpportunity` model combining `TradingPlan` with computed score
- [ ] Implement scoring algorithm in `services/scorer.go`:
  - R/R ratio weight: 40%
  - Sentiment weight: 25%
  - RSI position weight: 20%
  - Bias alignment weight: 15%
  - Filter out neutral bias and unfavorable RSI (>70 for longs, <30 for shorts)
- [ ] Add `GET /api/opportunities` endpoint — analyzed tickers sorted by score
- [ ] Add `POST /api/watchlist/analyze-all` — batch analyze entire watchlist

### 2.2 Firestore Setup & Data Models
- [ ] Add Firestore Go SDK dependency and initialize client
- [ ] Create Firestore collections:
  - `portfolios/{userId}` — budget, availableBalance, settings
  - `transactions/{txId}` — ticker, type, price, quantity, timestamp, status
  - `positions/{posId}` — ticker, entryPrice, quantity, stopLoss, targets, status
- [ ] Implement `services/portfolio.go` — CRUD for portfolio/budget management
- [ ] Implement `services/transactions.go` — transaction logging and position tracking

### 2.3 Trading Rules Engine
- [ ] Create `services/rules.go` with configurable trading rules:
  - Max position size (% of portfolio)
  - Max concurrent positions
  - Min score threshold to trade
  - Daily loss limit
- [ ] Implement position sizing calculator (fixed % or Kelly Criterion)
- [ ] Add `POST /api/settings` endpoint for rule configuration

### 2.4 Trading Execution
- [ ] Create `services/executor.go` — orchestrates trade decisions
- [ ] Implement eToro trade execution: `OpenPosition()`, `ClosePosition()`
- [ ] Add `POST /api/trade/execute` — manual trigger
- [ ] Add `GET /api/positions` — current open positions with P&L

### 2.5 Client Dashboard
- [ ] `OpportunitiesView` — ranked opportunities with score badges
- [ ] `PortfolioView` — budget, positions, P&L summary
- [ ] `TransactionsView` — transaction history
- [ ] Hooks: `useOpportunities()`, `usePortfolio()`, `useTransactions()`

### 2.6 Milestones
| Milestone | Deliverable | Est. Time |
|-----------|-------------|-----------|
| M1 | Scoring algorithm + model | 3 hrs |
| M2 | Opportunities API | 2 hrs |
| M3 | Firestore setup | 3 hrs |
| M4 | Portfolio & transactions | 4 hrs |
| M5 | Trading rules engine | 3 hrs |
| M6 | Position sizing | 2 hrs |
| M7 | eToro execution | 4 hrs |
| M8-9 | Client dashboards | 8 hrs |
| M10 | Integration tests | 4 hrs |

**Total Estimated Time: ~33 hours**

### 2.7 Decisions
- **Firestore over Postgres** — Simpler setup, real-time listeners, serverless
- **Scoring weights** — R/R ratio highest (40%) as it measures trade quality
- **Position sizing** — Fixed 2-5% per trade initially

---

## Phase 3: Full Autonomy (Future)

- Market scanning instead of predefined watchlist
- Notification system (webhook/email alerts)
- Scheduled auto-analysis (cron)
- Performance analytics and backtesting

## Phase 3: More Autonomy (Future)

- Instead of using predefined tickers from watchlist, allow the app to analyze the whole market and determine what tickers to trade
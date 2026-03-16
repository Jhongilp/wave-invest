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

- there will be a initial budget
- Once the trading plan is created, the app could trade using Etoro
- Use Postgres database to store trading transactions and tickers

## Phase 3: More Autonomy (Future)

- Instead of using predefined tickers from watchlist, allow the app to analyze the whole market and determine what tickers to trade
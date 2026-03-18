# Wave Invest - Build Log 🌊

A public development journal for Wave Invest, a personal swing trading assistant.

---

## Day 1 - March 12, 2026

**Published on X:**
> Day 1 of building Wave Invest 🌊
>
> A personal swing trading tool that:
> → Pulls my eToro watchlist
> → Runs AI analysis (Gemini)
> → Generates entry zones, stop losses, price targets
>
> Just got the skeleton up. React frontend, Go backend.
>
> Building in public. For myself, but maybe useful to others?

### What I Built

**Server (Go)**
- Set up project structure with Chi router
- Created API endpoints:
  - `GET /health` - Health check
  - `GET /api/watchlist` - Fetch eToro watchlist
  - `POST /api/analyze/:ticker` - Generate AI trading plan
  - `POST /api/analyze/batch` - Batch analysis
  - `GET /api/plan/:ticker` - Get cached plan
- Stubbed out eToro and Gemini API clients with mock data
- Added CORS support for local development

**Client (React + Vite + Tailwind)**
- Dark theme dashboard layout
- Watchlist sidebar showing tickers with price/change
- Trading plan view displaying:
  - AI summary
  - Moving averages (MA20, MA50, MA200)
  - RSI gauge
  - Support/Resistance levels
  - Trade setup (entry zone, stop loss, R/R ratio)
  - Price targets (PT1, PT2, PT3)
  - Sentiment analysis
- Custom hooks: `useWatchlist()`, `useTradingPlan()`

### Tech Stack
- React 19, Vite, TypeScript, Tailwind CSS
- Go 1.23, Chi router
- Gemini AI (for analysis)
- eToro API (for watchlist/data)

### Next Up
- [ ] Implement actual eToro API authentication
- [ ] Connect Gemini API with proper prompt engineering
- [ ] Add loading states and error handling polish
- [ ] Test end-to-end flow with real data

---

## Day 2 - March 18, 2026

**Published on X:**
> Day 2 of Wave Invest 🌊
>
> Added autonomous trading capabilities:
> → Daily batch analysis of entire watchlist
> → Scoring algorithm to rank opportunities (R/R, sentiment, RSI, bias)
> → Firestore for persistence
> → Auto-execution on eToro
>
> The bot can now trade without me clicking buttons 🤖

### What I Built

**Server (Go) - Phase 2: Autonomous Trading**

*Firestore Integration*
- Added Cloud Firestore SDK
- Created `pkg/firestore/client.go` - singleton Firestore client
- Created collections: portfolios, positions, transactions, daily_analysis

*New Models*
- `Portfolio` - budget, balance, trading rules
- `Position` - open/closed positions with entry/exit prices
- `Transaction` - BUY/SELL transaction history
- `DailyAnalysis` - daily analysis records with scores
- `ScoredOpportunity` - trading plan + score + breakdown
- `TradingRules` - configurable trading parameters

*Scoring Algorithm* (`services/scorer.go`)
- Weights: R/R ratio (40%), Sentiment (25%), RSI (20%), Bias (15%)
- Filters out: neutral bias, unfavorable RSI, low R/R (<1:1)
- Returns opportunities sorted by score descending

*Services*
- `PortfolioService` - portfolio CRUD operations
- `PositionService` - position management
- `TransactionService` - transaction logging
- `AnalysisService` - daily analysis storage with 7-day cleanup
- `DailyAnalysisOrchestrator` - orchestrates full workflow
- `Executor` - trade execution with position sizing
- `PositionSizer` - calculates position sizes based on rules

*New API Endpoints*
- `POST /api/daily-analysis` - run batch analysis
- `GET /api/opportunities` - today's scored opportunities
- `GET /api/opportunities/{date}` - historical opportunities
- `GET /api/portfolio` - portfolio summary
- `POST /api/portfolio` - create portfolio
- `GET /api/settings` - get trading rules
- `PUT /api/settings` - update trading rules
- `POST /api/execute-trades` - execute top opportunities
- `GET /api/positions` - open positions
- `DELETE /api/positions/{id}` - close position
- `GET /api/transactions` - transaction history

*eToro Client Extensions*
- `OpenPosition()` - place trades (stubbed for now)
- `ClosePosition()` - close positions
- `GetOpenPositions()` - fetch positions

**Client (React)**

*New Components*
- `OpportunitiesView` - display scored opportunities with "Run Daily Analysis" and "Execute Trades" buttons
- `PortfolioView` - portfolio summary, open positions, P&L
- `SettingsView` - configure trading rules (sliders)

*New Hooks*
- `useOpportunities()` - daily analysis + execution
- `usePortfolio()` - portfolio management
- `useSettings()` - trading rules CRUD

*Navigation*
- Updated `DashboardLayout` with sidebar nav
- Views: Watchlist | Opportunities | Portfolio | Settings

### Architecture Decisions
- **Manual trigger** for daily analysis (no cron)
- **Full auto execution** - system trades autonomously
- **7-day retention** - Firestore cleans up old analysis
- **Firestore > Postgres** - simpler setup, real-time capable

### Next Up
- [ ] Implement actual eToro trading API (currently stubbed)
- [ ] Add eToro sandbox/paper trading support
- [ ] Transaction history view
- [ ] Real-time position P&L updates
- [ ] Backtesting with historical data

---

## Template for Future Entries

```
## Day X - [Date]

### Progress
- What I built/fixed/improved

### Challenges
- Problems I ran into

### Learnings
- What I figured out

### Next Up
- [ ] What's coming next
```

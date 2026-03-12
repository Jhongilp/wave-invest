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

import { useState } from 'react';
import { DashboardLayout, Watchlist, TradingPlanView } from './components';
import { useWatchlist, useTradingPlan } from './hooks';

function App() {
  const [selectedTicker, setSelectedTicker] = useState<string | undefined>();
  const { tickers, loading: watchlistLoading, error: watchlistError } = useWatchlist();
  const { plan, loading: planLoading, error: planError, analyze } = useTradingPlan();

  const handleAnalyze = (ticker: string) => {
    setSelectedTicker(ticker);
    analyze(ticker);
  };

  return (
    <DashboardLayout
      sidebar={
        <Watchlist
          tickers={tickers}
          loading={watchlistLoading}
          error={watchlistError}
          onAnalyze={handleAnalyze}
          selectedTicker={selectedTicker}
        />
      }
    >
      <TradingPlanView
        plan={plan}
        loading={planLoading}
        error={planError}
      />
    </DashboardLayout>
  );
}

export default App;

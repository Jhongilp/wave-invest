import { useState } from 'react';
import { 
  DashboardLayout, 
  Watchlist, 
  TradingPlanView,
  OpportunitiesView,
  PortfolioView,
  SettingsView,
} from './components';
import type { ViewType } from './components/Layout/DashboardLayout';
import { useWatchlist, useTradingPlan } from './hooks';

function App() {
  const [currentView, setCurrentView] = useState<ViewType>('watchlist');
  const [selectedTicker, setSelectedTicker] = useState<string | undefined>();
  const { tickers, loading: watchlistLoading, error: watchlistError } = useWatchlist();
  const { plan, loading: planLoading, error: planError, analyze } = useTradingPlan();

  const handleAnalyze = (ticker: string) => {
    setSelectedTicker(ticker);
    analyze(ticker);
  };

  const handleSelectFromOpportunities = (ticker: string) => {
    setSelectedTicker(ticker);
    analyze(ticker);
    setCurrentView('watchlist');
  };

  const renderContent = () => {
    switch (currentView) {
      case 'opportunities':
        return <OpportunitiesView onSelectTicker={handleSelectFromOpportunities} />;
      case 'portfolio':
        return <PortfolioView />;
      case 'settings':
        return <SettingsView />;
      case 'watchlist':
      default:
        return (
          <TradingPlanView
            plan={plan}
            loading={planLoading}
            error={planError}
          />
        );
    }
  };

  return (
    <DashboardLayout
      currentView={currentView}
      onViewChange={setCurrentView}
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
      {renderContent()}
    </DashboardLayout>
  );
}

export default App;

import { useGainersSocket } from './hooks/useGainersSocket';
import { GainersTable } from './components/GainersTable';
import { ConnectionStatus } from './components/ConnectionStatus';

function App() {
  const { gainers, isConnected, error, lastUpdate, retry } = useGainersSocket();

  return (
    <div className="min-h-screen bg-gray-950 text-white">
      <div className="max-w-7xl mx-auto px-4 py-8">
        {/* Header */}
        <header className="mb-8">
          <h1 className="text-3xl font-bold text-white mb-2">
            📈 Top 20 Market Gainers
          </h1>
          <p className="text-gray-400">
            Real-time tracking of today's top performing stocks
          </p>
        </header>

        {/* Connection Status */}
        <ConnectionStatus
          isConnected={isConnected}
          error={error}
          lastUpdate={lastUpdate}
          onRetry={retry}
        />

        {/* Main Table */}
        <GainersTable 
          gainers={gainers} 
          isLoading={gainers.length === 0 && isConnected && !error} 
        />

        {/* Footer */}
        <footer className="mt-8 text-center text-gray-500 text-sm">
          <p>
            Data provided by{' '}
            <a
              href="https://polygon.io"
              target="_blank"
              rel="noopener noreferrer"
              className="text-blue-400 hover:text-blue-300 transition-colors"
            >
              Polygon.io
            </a>
            {' '}• Updates every 30 seconds
          </p>
        </footer>
      </div>
    </div>
  );
}

export default App;

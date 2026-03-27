import type { ReactNode } from 'react';
import { useTradingMode } from '../../hooks';

export type ViewType = 'watchlist' | 'opportunities' | 'portfolio' | 'settings';

interface NavItemProps {
  icon: string;
  label: string;
  active: boolean;
  onClick: () => void;
}

function NavItem({ icon, label, active, onClick }: NavItemProps) {
  return (
    <button
      onClick={onClick}
      className={`w-full flex items-center gap-3 px-4 py-3 text-left transition-colors ${
        active
          ? 'bg-blue-600/20 text-blue-400 border-r-2 border-blue-400'
          : 'text-gray-400 hover:bg-gray-800 hover:text-white'
      }`}
    >
      <span className="text-lg">{icon}</span>
      <span className="font-medium">{label}</span>
    </button>
  );
}

function TradingModeToggle() {
  const { isDemo, updating, toggleMode } = useTradingMode();

  return (
    <div className="flex items-center gap-2">
      <span className={`text-xs font-medium ${isDemo ? 'text-gray-400' : 'text-green-400'}`}>
        {isDemo ? 'DEMO' : 'REAL'}
      </span>
      <button
        onClick={() => toggleMode()}
        disabled={updating}
        className={`relative w-11 h-6 rounded-full transition-colors ${
          isDemo ? 'bg-gray-600' : 'bg-green-600'
        } ${updating ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}`}
        title={`Switch to ${isDemo ? 'real' : 'demo'} trading`}
      >
        <span
          className={`absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full transition-transform ${
            isDemo ? 'translate-x-0' : 'translate-x-5'
          }`}
        />
      </button>
      {!isDemo && (
        <span className="text-xs text-yellow-500 font-medium animate-pulse">⚠️ LIVE</span>
      )}
    </div>
  );
}

interface DashboardLayoutProps {
  sidebar: ReactNode;
  children: ReactNode;
  currentView: ViewType;
  onViewChange: (view: ViewType) => void;
}

export function DashboardLayout({ sidebar, children, currentView, onViewChange }: DashboardLayoutProps) {
  return (
    <div className="min-h-screen bg-gray-950 text-white flex">
      {/* Sidebar */}
      <aside className="w-80 bg-gray-900 border-r border-gray-800 flex flex-col">
        {/* Logo & Mode Toggle */}
        <div className="p-4 border-b border-gray-800">
          <div className="flex items-center justify-between">
            <h1 className="text-xl font-bold text-white flex items-center gap-2">
              <span className="text-2xl">🌊</span>
              Wave Invest
            </h1>
          </div>
          <div className="mt-3">
            <TradingModeToggle />
          </div>
        </div>

        {/* Navigation */}
        <nav className="border-b border-gray-800">
          <NavItem
            icon="📋"
            label="Watchlist"
            active={currentView === 'watchlist'}
            onClick={() => onViewChange('watchlist')}
          />
          <NavItem
            icon="🎯"
            label="Opportunities"
            active={currentView === 'opportunities'}
            onClick={() => onViewChange('opportunities')}
          />
          <NavItem
            icon="💼"
            label="Portfolio"
            active={currentView === 'portfolio'}
            onClick={() => onViewChange('portfolio')}
          />
          <NavItem
            icon="⚙️"
            label="Settings"
            active={currentView === 'settings'}
            onClick={() => onViewChange('settings')}
          />
        </nav>

        {/* Sidebar Content (Watchlist when on watchlist view) */}
        {currentView === 'watchlist' && (
          <div className="flex-1 overflow-auto">
            {sidebar}
          </div>
        )}

        {/* Spacer for non-watchlist views */}
        {currentView !== 'watchlist' && <div className="flex-1" />}

        {/* Footer */}
        <div className="p-4 border-t border-gray-800 text-xs text-gray-500">
          <p>Swing Trading Assistant</p>
          <p>Powered by Gemini AI</p>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 flex flex-col overflow-hidden">
        {children}
      </main>
    </div>
  );
}

import { useState } from 'react';
import { usePortfolio } from '../../hooks';
import type { Position } from '../../types';

function formatCurrency(value: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(value);
}

function formatPercent(value: number): string {
  return `${(value * 100).toFixed(1)}%`;
}

interface PositionRowProps {
  position: Position;
  onClose: (positionId: string) => void;
}

function PositionRow({ position, onClose }: PositionRowProps) {
  const positionValue = position.entryPrice * position.quantity;
  
  return (
    <tr className="border-b border-gray-700">
      <td className="py-3 px-4 text-white font-medium">{position.ticker}</td>
      <td className="py-3 px-4 text-gray-300">{formatCurrency(position.entryPrice)}</td>
      <td className="py-3 px-4 text-gray-300">{position.quantity.toFixed(2)}</td>
      <td className="py-3 px-4 text-gray-300">{formatCurrency(positionValue)}</td>
      <td className="py-3 px-4 text-red-400">{formatCurrency(position.stopLoss)}</td>
      <td className="py-3 px-4 text-green-400">{formatCurrency(position.takeProfit)}</td>
      <td className="py-3 px-4">
        <button
          onClick={() => onClose(position.id)}
          className="text-red-400 hover:text-red-300 text-sm"
        >
          Close
        </button>
      </td>
    </tr>
  );
}

export function PortfolioView() {
  const {
    portfolio,
    positions,
    loading,
    error,
    createPortfolio,
    closePosition,
    fetchPortfolio,
  } = usePortfolio();

  const [budgetInput, setBudgetInput] = useState('');
  const [showCreateForm, setShowCreateForm] = useState(false);

  const handleCreatePortfolio = async () => {
    const budget = parseFloat(budgetInput);
    if (budget > 0) {
      await createPortfolio(budget);
      setShowCreateForm(false);
      setBudgetInput('');
    }
  };

  if (loading && !portfolio) {
    return (
      <div className="p-6 flex items-center justify-center">
        <p className="text-gray-400">Loading portfolio...</p>
      </div>
    );
  }

  if (!portfolio?.budget && !showCreateForm) {
    return (
      <div className="p-6">
        <div className="text-center py-12">
          <h2 className="text-xl font-bold text-white mb-4">No Portfolio Found</h2>
          <p className="text-gray-400 mb-6">Create a portfolio to start trading</p>
          <button
            onClick={() => setShowCreateForm(true)}
            className="bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded-lg font-medium"
          >
            Create Portfolio
          </button>
        </div>
      </div>
    );
  }

  if (showCreateForm) {
    return (
      <div className="p-6">
        <div className="max-w-md mx-auto bg-gray-800 rounded-lg p-6">
          <h2 className="text-xl font-bold text-white mb-4">Create Portfolio</h2>
          <div className="mb-4">
            <label className="block text-gray-400 text-sm mb-2">Initial Budget</label>
            <input
              type="number"
              value={budgetInput}
              onChange={(e) => setBudgetInput(e.target.value)}
              placeholder="10000"
              className="w-full bg-gray-700 text-white px-4 py-2 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div className="flex gap-3">
            <button
              onClick={handleCreatePortfolio}
              className="flex-1 bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg font-medium"
            >
              Create
            </button>
            <button
              onClick={() => setShowCreateForm(false)}
              className="px-4 py-2 text-gray-400 hover:text-white"
            >
              Cancel
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-white">Portfolio</h1>
        <button
          onClick={() => fetchPortfolio()}
          className="text-gray-400 hover:text-white text-sm"
        >
          Refresh
        </button>
      </div>

      {error && (
        <div className="bg-red-900/50 border border-red-500 text-red-200 px-4 py-3 rounded-lg mb-6">
          {error}
        </div>
      )}

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-8">
        <div className="bg-gray-800 rounded-lg p-4">
          <p className="text-gray-400 text-sm">Total Budget</p>
          <p className="text-2xl font-bold text-white">{formatCurrency(portfolio?.budget || 0)}</p>
        </div>
        <div className="bg-gray-800 rounded-lg p-4">
          <p className="text-gray-400 text-sm">Available Balance</p>
          <p className="text-2xl font-bold text-white">{formatCurrency(portfolio?.availableBalance || 0)}</p>
        </div>
        <div className="bg-gray-800 rounded-lg p-4">
          <p className="text-gray-400 text-sm">Invested</p>
          <p className="text-2xl font-bold text-white">{formatCurrency(portfolio?.investedAmount || 0)}</p>
        </div>
        <div className="bg-gray-800 rounded-lg p-4">
          <p className="text-gray-400 text-sm">Daily P&L</p>
          <p className={`text-2xl font-bold ${(portfolio?.dailyPnl || 0) >= 0 ? 'text-green-400' : 'text-red-400'}`}>
            {formatCurrency(portfolio?.dailyPnl || 0)}
          </p>
        </div>
      </div>

      {/* Trading Rules Summary */}
      {portfolio?.rules && (
        <div className="bg-gray-800 rounded-lg p-4 mb-8">
          <h3 className="text-lg font-medium text-white mb-3">Trading Rules</h3>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
            <div>
              <p className="text-gray-400">Max Position %</p>
              <p className="text-white">{formatPercent(portfolio.rules.maxPositionPercent)}</p>
            </div>
            <div>
              <p className="text-gray-400">Max Positions</p>
              <p className="text-white">{portfolio.rules.maxConcurrentPositions}</p>
            </div>
            <div>
              <p className="text-gray-400">Min Score</p>
              <p className="text-white">{portfolio.rules.minScoreThreshold}</p>
            </div>
            <div>
              <p className="text-gray-400">Daily Loss Limit</p>
              <p className="text-white">{formatPercent(portfolio.rules.dailyLossLimit)}</p>
            </div>
          </div>
        </div>
      )}

      {/* Open Positions */}
      <div className="bg-gray-800 rounded-lg overflow-hidden">
        <div className="px-4 py-3 border-b border-gray-700">
          <h3 className="text-lg font-medium text-white">
            Open Positions ({positions.length})
          </h3>
        </div>
        {positions.length === 0 ? (
          <div className="px-4 py-8 text-center text-gray-400">
            No open positions
          </div>
        ) : (
          <table className="w-full">
            <thead className="bg-gray-750">
              <tr className="text-left text-gray-400 text-sm">
                <th className="py-2 px-4">Ticker</th>
                <th className="py-2 px-4">Entry Price</th>
                <th className="py-2 px-4">Quantity</th>
                <th className="py-2 px-4">Value</th>
                <th className="py-2 px-4">Stop Loss</th>
                <th className="py-2 px-4">Take Profit</th>
                <th className="py-2 px-4">Action</th>
              </tr>
            </thead>
            <tbody>
              {positions.map((position) => (
                <PositionRow
                  key={position.id}
                  position={position}
                  onClose={closePosition}
                />
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}

import type { Gainer } from '../types/market';
import { GainerRow } from './GainerRow';

interface GainersTableProps {
  gainers: Gainer[];
  isLoading: boolean;
}

function SkeletonRow({ index }: { index: number }) {
  const bgColor = index % 2 === 0 ? 'bg-gray-800/50' : 'bg-gray-900/50';
  
  return (
    <tr className={bgColor}>
      <td className="px-4 py-3">
        <div className="skeleton h-5 w-6 rounded" />
      </td>
      <td className="px-4 py-3">
        <div className="skeleton h-5 w-16 rounded" />
      </td>
      <td className="px-4 py-3">
        <div className="skeleton h-5 w-20 rounded ml-auto" />
      </td>
      <td className="px-4 py-3">
        <div className="skeleton h-5 w-16 rounded ml-auto" />
      </td>
      <td className="px-4 py-3">
        <div className="skeleton h-5 w-16 rounded ml-auto" />
      </td>
      <td className="px-4 py-3">
        <div className="skeleton h-5 w-20 rounded ml-auto" />
      </td>
      <td className="px-4 py-3">
        <div className="skeleton h-5 w-20 rounded ml-auto" />
      </td>
    </tr>
  );
}

export function GainersTable({ gainers, isLoading }: GainersTableProps) {
  return (
    <div className="overflow-x-auto rounded-lg border border-gray-700 shadow-xl">
      <table className="w-full text-sm">
        <thead className="bg-gray-800 text-gray-300 uppercase tracking-wider text-xs">
          <tr>
            <th className="px-4 py-4 text-left font-semibold">#</th>
            <th className="px-4 py-4 text-left font-semibold">Ticker</th>
            <th className="px-4 py-4 text-right font-semibold">Price</th>
            <th className="px-4 py-4 text-right font-semibold">Change %</th>
            <th className="px-4 py-4 text-right font-semibold">Change $</th>
            <th className="px-4 py-4 text-right font-semibold">Volume</th>
            <th className="px-4 py-4 text-right font-semibold">Prev Close</th>
          </tr>
        </thead>
        <tbody>
          {isLoading ? (
            // Skeleton loading state
            Array.from({ length: 10 }).map((_, index) => (
              <SkeletonRow key={index} index={index} />
            ))
          ) : gainers.length > 0 ? (
            // Actual data
            gainers.map((gainer, index) => (
              <GainerRow key={gainer.ticker} gainer={gainer} rank={index + 1} />
            ))
          ) : (
            // Empty state
            <tr>
              <td colSpan={7} className="px-4 py-12 text-center text-gray-400">
                No gainers data available. Waiting for market data...
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  );
}

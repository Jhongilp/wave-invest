import { memo, useRef, useEffect, useState } from 'react';
import type { Gainer } from '../types/market';

interface GainerRowProps {
  gainer: Gainer;
  rank: number;
}

function formatVolume(volume: number): string {
  if (volume >= 1_000_000_000) {
    return `${(volume / 1_000_000_000).toFixed(2)}B`;
  }
  if (volume >= 1_000_000) {
    return `${(volume / 1_000_000).toFixed(2)}M`;
  }
  if (volume >= 1_000) {
    return `${(volume / 1_000).toFixed(2)}K`;
  }
  return volume.toString();
}

function formatPrice(price: number): string {
  return price.toLocaleString('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
}

function formatPercent(percent: number): string {
  return `${percent >= 0 ? '+' : ''}${percent.toFixed(2)}%`;
}

export const GainerRow = memo(function GainerRow({ gainer, rank }: GainerRowProps) {
  const prevPriceRef = useRef<number>(gainer.price);
  const [animationClass, setAnimationClass] = useState<string>('');

  useEffect(() => {
    if (prevPriceRef.current !== gainer.price) {
      const isUp = gainer.price > prevPriceRef.current;
      setAnimationClass(isUp ? 'flash-green' : 'flash-red');
      prevPriceRef.current = gainer.price;

      // Remove animation class after animation completes
      const timeout = setTimeout(() => {
        setAnimationClass('');
      }, 500);

      return () => clearTimeout(timeout);
    }
  }, [gainer.price]);

  const isPositive = gainer.changePercent >= 0;
  const changeColor = isPositive ? 'text-green-500' : 'text-red-500';
  const bgColor = rank % 2 === 0 ? 'bg-gray-800/50' : 'bg-gray-900/50';

  return (
    <tr className={`${bgColor} ${animationClass} transition-colors duration-200 hover:bg-gray-700/50`}>
      <td className="px-4 py-3 text-gray-400 font-medium">
        {rank}
      </td>
      <td className="px-4 py-3">
        <span className="font-bold text-white">{gainer.ticker}</span>
      </td>
      <td className="px-4 py-3 text-right font-mono text-white">
        {formatPrice(gainer.price)}
      </td>
      <td className={`px-4 py-3 text-right font-mono font-semibold ${changeColor}`}>
        {formatPercent(gainer.changePercent)}
      </td>
      <td className={`px-4 py-3 text-right font-mono ${changeColor}`}>
        {isPositive ? '+' : ''}{formatPrice(gainer.change)}
      </td>
      <td className="px-4 py-3 text-right font-mono text-gray-300">
        {formatVolume(gainer.volume)}
      </td>
      <td className="px-4 py-3 text-right font-mono text-gray-400">
        {formatPrice(gainer.previousClose)}
      </td>
    </tr>
  );
});

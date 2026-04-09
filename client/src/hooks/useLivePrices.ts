import { useState, useEffect, useCallback, useRef } from 'react';
import type { LivePrice, LivePriceMap, WSMessage } from '../types';

// Derive WebSocket URL from API URL in production, fallback to localhost in development
function getWebSocketUrl(): string {
  if (import.meta.env.VITE_WS_URL) {
    return import.meta.env.VITE_WS_URL;
  }
  const apiUrl = import.meta.env.VITE_API_URL;
  if (apiUrl) {
    // Convert https:// to wss:// or http:// to ws://
    return apiUrl.replace(/^https:/, 'wss:').replace(/^http:/, 'ws:') + '/ws';
  }
  return 'ws://localhost:8080/ws';
}

const WS_URL = getWebSocketUrl();
const RECONNECT_DELAY = 3000;
const MAX_RECONNECT_ATTEMPTS = 5;

interface UseLivePricesOptions {
  symbols?: string[];
  autoConnect?: boolean;
}

interface UseLivePricesReturn {
  prices: LivePriceMap;
  isConnected: boolean;
  error: string | null;
  subscribe: (symbols: string[]) => void;
  unsubscribe: () => void;
  getPrice: (symbol: string) => LivePrice | undefined;
}

export function useLivePrices(options: UseLivePricesOptions = {}): UseLivePricesReturn {
  const { symbols = [], autoConnect = true } = options;
  
  const [prices, setPrices] = useState<LivePriceMap>({});
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectAttempts = useRef(0);
  const reconnectTimeout = useRef<ReturnType<typeof setTimeout> | null>(null);
  const subscribedSymbols = useRef<string[]>([]);

  // Connect to WebSocket
  const connect = useCallback(() => {
    // Prevent duplicate connections - check both OPEN and CONNECTING states
    const currentState = wsRef.current?.readyState;
    if (currentState === WebSocket.OPEN || currentState === WebSocket.CONNECTING) {
      return;
    }

    try {
      const ws = new WebSocket(WS_URL);
      wsRef.current = ws;

      ws.onopen = () => {
        setIsConnected(true);
        setError(null);
        reconnectAttempts.current = 0;
        
        // Resubscribe if we have symbols
        if (subscribedSymbols.current.length > 0) {
          ws.send(JSON.stringify({
            action: 'subscribe',
            symbols: subscribedSymbols.current
          }));
        }
      };

      ws.onmessage = (event) => {
        try {
          const msg: WSMessage = JSON.parse(event.data);
          
          if (msg.type === 'price' && msg.symbol) {
            // Validate prices - skip updates without valid price data
            const bid = msg.bid ?? 0;
            const ask = msg.ask ?? 0;
            const last = msg.last ?? 0;
            
            // Skip if all prices are zero or missing
            if (bid === 0 && ask === 0 && last === 0) {
              return;
            }
            
            const livePrice: LivePrice = {
              symbol: msg.symbol,
              bid,
              ask,
              last,
              timestamp: msg.timestamp || new Date().toISOString()
            };
            
            setPrices(prev => ({
              ...prev,
              [msg.symbol!]: livePrice
            }));
          } else if (msg.type === 'error') {
            setError(msg.message || 'Unknown error');
          }
        } catch {
          console.error('Failed to parse WebSocket message');
        }
      };

      ws.onerror = () => {
        setError('WebSocket connection error');
      };
    } catch {
      setError('Failed to connect to price server');
    }
  }, []);

  // Handle reconnection on close - using effect to avoid stale closure
  useEffect(() => {
    const handleClose = () => {
      setIsConnected(false);
      wsRef.current = null;

      // Attempt reconnection
      if (reconnectAttempts.current < MAX_RECONNECT_ATTEMPTS) {
        reconnectAttempts.current++;
        reconnectTimeout.current = setTimeout(() => {
          connect();
        }, RECONNECT_DELAY * reconnectAttempts.current);
      } else {
        setError('Connection lost. Please refresh the page.');
      }
    };

    const ws = wsRef.current;
    if (ws) {
      ws.onclose = handleClose;
    }
  }, [connect]);

  // Subscribe to symbols
  const subscribe = useCallback((newSymbols: string[]) => {
    if (newSymbols.length === 0) return;
    
    subscribedSymbols.current = newSymbols;
    
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({
        action: 'subscribe',
        symbols: newSymbols
      }));
    } else {
      // Connect if not connected
      connect();
    }
  }, [connect]);

  // Unsubscribe from all symbols
  const unsubscribe = useCallback(() => {
    subscribedSymbols.current = [];
    
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({
        action: 'unsubscribe',
        symbols: []
      }));
    }
    
    setPrices({});
  }, []);

  // Get price for a specific symbol
  const getPrice = useCallback((symbol: string): LivePrice | undefined => {
    return prices[symbol];
  }, [prices]);

  // Auto-connect and subscribe on mount
  useEffect(() => {
    let mounted = true;
    
    if (autoConnect && symbols.length > 0) {
      subscribedSymbols.current = symbols;
      // Use setTimeout to defer connection to next tick, avoiding sync setState warning
      const connectTimer = setTimeout(() => {
        if (mounted) {
          connect();
        }
      }, 0);
      
      return () => {
        mounted = false;
        clearTimeout(connectTimer);
        if (reconnectTimeout.current) {
          clearTimeout(reconnectTimeout.current);
        }
        if (wsRef.current) {
          wsRef.current.close();
          wsRef.current = null;
        }
      };
    }

    return () => {
      mounted = false;
      if (reconnectTimeout.current) {
        clearTimeout(reconnectTimeout.current);
      }
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, [autoConnect, connect, symbols]);

  // Update subscription when symbols change
  useEffect(() => {
    if (symbols.length > 0 && wsRef.current?.readyState === WebSocket.OPEN) {
      // Defer to avoid sync setState warning
      const timer = setTimeout(() => {
        if (wsRef.current?.readyState === WebSocket.OPEN) {
          wsRef.current.send(JSON.stringify({
            action: 'subscribe',
            symbols
          }));
        }
      }, 0);
      return () => clearTimeout(timer);
    }
  }, [symbols]);

  return {
    prices,
    isConnected,
    error,
    subscribe,
    unsubscribe,
    getPrice
  };
}

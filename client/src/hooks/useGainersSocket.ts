import { useState, useEffect, useRef, useCallback } from 'react';
import type { Gainer, WebSocketMessage, ConnectionState } from '../types/market';

const WS_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws';
const MAX_RECONNECT_ATTEMPTS = 10;
const INITIAL_RECONNECT_DELAY = 1000;

export function useGainersSocket() {
  const [gainers, setGainers] = useState<Gainer[]>([]);
  const [connectionState, setConnectionState] = useState<ConnectionState>({
    isConnected: false,
    error: null,
    reconnectAttempt: 0,
  });
  const [lastUpdate, setLastUpdate] = useState<Date | null>(null);

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<number | null>(null);
  const reconnectAttemptRef = useRef(0);

  const connect = useCallback(() => {
    // Clean up existing connection
    if (wsRef.current) {
      wsRef.current.close();
    }

    try {
      const ws = new WebSocket(WS_URL);
      wsRef.current = ws;

      ws.onopen = () => {
        console.log('WebSocket connected');
        reconnectAttemptRef.current = 0;
        setConnectionState({
          isConnected: true,
          error: null,
          reconnectAttempt: 0,
        });
      };

      ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data);
          
          if (message.type === 'gainers' && Array.isArray(message.data)) {
            setGainers(message.data);
            setLastUpdate(new Date(message.timestamp));
          }
        } catch (err) {
          console.error('Failed to parse WebSocket message:', err);
        }
      };

      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        setConnectionState((prev) => ({
          ...prev,
          error: 'Connection error occurred',
        }));
      };

      ws.onclose = (event) => {
        console.log('WebSocket closed:', event.code, event.reason);
        setConnectionState((prev) => ({
          ...prev,
          isConnected: false,
        }));

        // Attempt to reconnect with exponential backoff
        if (reconnectAttemptRef.current < MAX_RECONNECT_ATTEMPTS) {
          const delay = INITIAL_RECONNECT_DELAY * Math.pow(2, reconnectAttemptRef.current);
          reconnectAttemptRef.current += 1;
          
          setConnectionState((prev) => ({
            ...prev,
            reconnectAttempt: reconnectAttemptRef.current,
            error: `Reconnecting in ${delay / 1000}s... (attempt ${reconnectAttemptRef.current}/${MAX_RECONNECT_ATTEMPTS})`,
          }));

          reconnectTimeoutRef.current = window.setTimeout(() => {
            connect();
          }, delay);
        } else {
          setConnectionState((prev) => ({
            ...prev,
            error: 'Max reconnection attempts reached. Please refresh the page.',
          }));
        }
      };
    } catch (err) {
      console.error('Failed to create WebSocket:', err);
      setConnectionState({
        isConnected: false,
        error: 'Failed to connect to server',
        reconnectAttempt: 0,
      });
    }
  }, []);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
  }, []);

  const retry = useCallback(() => {
    reconnectAttemptRef.current = 0;
    setConnectionState({
      isConnected: false,
      error: null,
      reconnectAttempt: 0,
    });
    connect();
  }, [connect]);

  useEffect(() => {
    connect();
    return () => {
      disconnect();
    };
  }, [connect, disconnect]);

  return {
    gainers,
    isConnected: connectionState.isConnected,
    error: connectionState.error,
    reconnectAttempt: connectionState.reconnectAttempt,
    lastUpdate,
    retry,
  };
}

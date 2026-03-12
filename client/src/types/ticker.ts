export interface Ticker {
  symbol: string;
  name: string;
  price: number;
  change: number;
  changePercent: number;
}

export interface WatchlistResponse {
  tickers: Ticker[];
}

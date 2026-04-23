export interface Stock {
  stockId: number;
  symbol: string;
  name: string;
  quantity: number;
  avgPrice: number;
  currentPrice: number;
  changePct: number;
  market: "KR" | "US";
  isWatchlist: boolean;
}

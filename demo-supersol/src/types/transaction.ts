export interface Transaction {
  txId: number;
  accountId: number;
  txType: "DEPOSIT" | "WITHDRAW" | "TRANSFER";
  amount: number;
  balanceAfter: number;
  description: string;
  counterparty: string;
  category: string;
  createdAt: string;
}

// 거래내역 조회 필터
export interface TransactionFilter {
  accountId: number;
  days: number;        // 최근 N일
  category?: string;   // 카테고리 필터 (optional)
  limit?: number;      // 조회 건수 제한
}

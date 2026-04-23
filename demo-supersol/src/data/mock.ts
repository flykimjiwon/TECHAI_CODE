import type { Account, Card } from "@/types/account";
import type { Transaction } from "@/types/transaction";
import type { Stock } from "@/types/stock";

// ── 고객 정보 ──
export const CUSTOMER = {
  name: "김지원",
  membership: "PREMIER",
  points: 2751,
  lastLogin: "2026.04.21 09:32:15",
};

// ── 이번 달 포인트 적립 목표 ──
export const POINT_GOAL = {
  target: 5000,       // 이번 달 목표 포인트
  label: "4월 적립 목표",
};

// ── 계좌 ──
export const accounts: Account[] = [
  {
    accountId: 1,
    accountNumber: "110-456-789012",
    accountName: "신한 주거래통장",
    bankName: "신한은행",
    balance: 11234510,
    accountType: "CHECKING",
    isPrimary: true,
  },
  {
    accountId: 2,
    accountNumber: "110-789-012345",
    accountName: "신한 급여통장",
    bankName: "신한은행",
    balance: 3200000,
    accountType: "CHECKING",
    isPrimary: false,
  },
  {
    accountId: 3,
    accountNumber: "270-80-900312",
    accountName: "ISA(중개형)",
    bankName: "신한투자증권",
    balance: 536082,
    accountType: "ISA",
    isPrimary: false,
  },
];

// ── 카드 ──
export const cards: Card[] = [
  {
    cardId: 1,
    cardName: "신한은행 Welpro",
    cardNumber: "9411-****-****-7823",
    monthlyUsage: 230340,
  },
];

// ── 거래내역 (최근 30일) ──
export const transactions: Transaction[] = [
  { txId: 1,  accountId: 1, txType: "WITHDRAW", amount: 4500,    balanceAfter: 11234510, description: "스타벅스 강남점",      counterparty: "스타벅스",     category: "FOOD",      createdAt: "2026-04-21T10:30:00" },
  { txId: 2,  accountId: 1, txType: "DEPOSIT",  amount: 3200000, balanceAfter: 11239010, description: "4월 급여",            counterparty: "신한은행",     category: "SALARY",    createdAt: "2026-04-20T09:00:00" },
  { txId: 3,  accountId: 1, txType: "WITHDRAW", amount: 12300,   balanceAfter: 8039010,  description: "네이버페이 결제",      counterparty: "네이버",       category: "SHOPPING",  createdAt: "2026-04-19T14:22:00" },
  { txId: 4,  accountId: 1, txType: "WITHDRAW", amount: 1250,    balanceAfter: 8051310,  description: "서울 지하철",          counterparty: "서울교통공사",  category: "TRANSPORT", createdAt: "2026-04-18T08:15:00" },
  { txId: 5,  accountId: 1, txType: "WITHDRAW", amount: 8900,    balanceAfter: 8052560,  description: "GS25 편의점",         counterparty: "GS리테일",     category: "FOOD",      createdAt: "2026-04-18T19:45:00" },
  { txId: 6,  accountId: 1, txType: "WITHDRAW", amount: 45000,   balanceAfter: 8061460,  description: "CGV 영등포",          counterparty: "CJ올리브",     category: "SHOPPING",  createdAt: "2026-04-16T20:00:00" },
  { txId: 7,  accountId: 1, txType: "TRANSFER", amount: 500000,  balanceAfter: 8106460,  description: "김지원 → 신한급여통장", counterparty: "본인이체",     category: "TRANSFER",  createdAt: "2026-04-16T11:00:00" },
  { txId: 8,  accountId: 1, txType: "WITHDRAW", amount: 15800,   balanceAfter: 8606460,  description: "배달의민족",           counterparty: "우아한형제",   category: "FOOD",      createdAt: "2026-04-15T12:30:00" },
  // ── 7일 이후 데이터 (현재 API에서 안 보임) ──
  { txId: 9,  accountId: 1, txType: "WITHDRAW", amount: 89000,   balanceAfter: 8622260,  description: "쿠팡 주문",           counterparty: "쿠팡",         category: "SHOPPING",  createdAt: "2026-04-13T16:00:00" },
  { txId: 10, accountId: 1, txType: "WITHDRAW", amount: 32000,   balanceAfter: 8711260,  description: "주유소 SK에너지",      counterparty: "SK에너지",     category: "TRANSPORT", createdAt: "2026-04-11T10:20:00" },
  { txId: 11, accountId: 1, txType: "DEPOSIT",  amount: 3200000, balanceAfter: 8743260,  description: "3월 급여",            counterparty: "신한은행",     category: "SALARY",    createdAt: "2026-04-09T09:00:00" },
  { txId: 12, accountId: 1, txType: "WITHDRAW", amount: 150000,  balanceAfter: 5543260,  description: "신한카드 자동이체",     counterparty: "신한카드",     category: "TRANSFER",  createdAt: "2026-04-06T00:00:00" },
  { txId: 13, accountId: 1, txType: "WITHDRAW", amount: 67000,   balanceAfter: 5693260,  description: "이마트 장보기",        counterparty: "이마트",       category: "FOOD",      createdAt: "2026-04-03T15:30:00" },
  { txId: 14, accountId: 1, txType: "WITHDRAW", amount: 25000,   balanceAfter: 5760260,  description: "택시비",              counterparty: "카카오모빌",   category: "TRANSPORT", createdAt: "2026-03-30T23:10:00" },
  { txId: 15, accountId: 1, txType: "WITHDRAW", amount: 120000,  balanceAfter: 5785260,  description: "병원비 (건강검진)",     counterparty: "서울아산",     category: "MEDICAL",   createdAt: "2026-03-24T11:00:00" },
];

// ── 관심 종목 ──
export const stocks: Stock[] = [
  { stockId: 1, symbol: "TSLA",   name: "테슬라",        quantity: 5,  avgPrice: 350.00, currentPrice: 392.50, changePct: -2.03, market: "US", isWatchlist: true },
  { stockId: 2, symbol: "NVDA",   name: "엔비디아",      quantity: 3,  avgPrice: 180.00, currentPrice: 202.06, changePct: 0.19,  market: "US", isWatchlist: true },
  { stockId: 3, symbol: "PLTR",   name: "팔란티어 테크",  quantity: 10, avgPrice: 120.00, currentPrice: 145.89, changePct: -0.34, market: "US", isWatchlist: true },
  { stockId: 4, symbol: "005930", name: "삼성전자",       quantity: 50, avgPrice: 72000,  currentPrice: 68500,  changePct: -1.42, market: "KR", isWatchlist: true },
];

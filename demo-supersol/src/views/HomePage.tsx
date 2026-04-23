"use client";

import { ChevronRight, Search } from "lucide-react";
import Header from "@/components/Header";
import Card from "@/components/Card";
import { CUSTOMER, accounts, cards, transactions } from "@/data/mock";
import TransactionItem from "@/components/TransactionItem";

function formatWon(amount: number): string {
  return amount.toLocaleString("ko-KR") + "원";
}

export default function HomePage() {
  const primaryAccount = accounts.find((a) => a.isPrimary);
  const card = cards[0];
  const totalAssets = accounts.reduce((sum, a) => sum + a.balance, 0);

  // 최근 거래 3건 (latest first)
  const recentTx = transactions.slice(0, 3);

  return (
    <div className="p-4 space-y-4 pb-20">
      <Header />

      {/* 마이신한포인트 */}
      <div className="bg-blue-50 rounded-2xl p-4 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="bg-blue-600 text-white p-1 rounded-full text-xs font-bold w-5 h-5 flex items-center justify-center">
            P
          </div>
          <span className="text-sm font-medium">마이신한포인트</span>
          <span className="text-blue-600 font-bold ml-1">
            {CUSTOMER.points.toLocaleString()} P
          </span>
        </div>
        <div className="w-10 h-5 bg-blue-500 rounded-full relative">
          <div className="absolute right-1 top-1 w-3 h-3 bg-white rounded-full" />
        </div>
      </div>

      {/* 은행 섹션 */}
      <div>
        <div className="flex justify-between items-center mb-2 px-1">
          <h2 className="text-lg font-bold">은행</h2>
          <ChevronRight className="w-5 h-5 text-gray-400" />
        </div>
        {primaryAccount && (
          <div className="bg-[#0046FF] text-white rounded-3xl p-5 shadow-sm mb-4">
            <div className="flex items-center gap-2 mb-2">
              <div className="bg-white rounded-full p-1 w-6 h-6 flex items-center justify-center text-[10px] font-bold text-blue-600">
                S
              </div>
              <span className="text-sm opacity-90">{primaryAccount.accountName}</span>
              <span className="text-xs opacity-60">{primaryAccount.accountNumber}</span>
            </div>
            <div className="flex justify-between items-end">
              <span className="text-2xl font-bold">{formatWon(primaryAccount.balance)}</span>
              <button className="bg-white/20 hover:bg-white/30 px-4 py-1.5 rounded-lg text-sm font-medium">
                이체
              </button>
            </div>
          </div>
        )}
      </div>

      {/* 카드 섹션 */}
      <div>
        <div className="flex justify-between items-center mb-2 px-1">
          <h2 className="text-lg font-bold">카드</h2>
          <ChevronRight className="w-5 h-5 text-gray-400" />
        </div>
        {card && (
          <Card>
            <div className="flex items-start gap-4">
              <div className="w-10 h-16 bg-gradient-to-b from-blue-600 to-blue-800 rounded-sm" />
              <div className="flex-1">
                <p className="text-sm text-gray-500">{card.cardName}</p>
                <p className="text-xs text-gray-400 mb-2">4월 이용금액</p>
                <p className="text-xl font-bold">{formatWon(card.monthlyUsage)}</p>
              </div>
              <button className="text-blue-500 bg-blue-50 px-3 py-1.5 rounded-lg text-xs font-bold">
                결제카드 등록
              </button>
            </div>
          </Card>
        )}
      </div>

      {/* 최근 거래 내역 미리보기 */}
      <Card className="mb-0">
        <h3 className="text-sm font-bold mb-2">최근 거래내역</h3>
        {recentTx.map((tx) => (
          <TransactionItem key={tx.txId} tx={tx} />
        ))}
        <div className="text-right mt-2">
          <a href="/transactions" className="text-blue-600 text-xs underline">
            전체 보기 &gt;
          </a>
        </div>
      </Card>

      {/* 자산 + 보험 */}
      <div className="grid grid-cols-2 gap-4">
        <Card className="mb-0">
          <p className="text-xs text-gray-500 mb-1">총 자산</p>
          <p className="font-bold text-lg leading-tight">{formatWon(totalAssets)}</p>
          <p className="text-red-500 text-xs">+ 0(0.00%)</p>
        </Card>
        <Card className="mb-0">
          <p className="font-bold text-sm">내 보험 분석</p>
          <p className="text-xs text-gray-400">꼭 필요한 보험만</p>
          <div className="flex justify-end mt-2">
            <div className="bg-blue-50 p-2 rounded-lg">
              <Search className="w-4 h-4 text-blue-500" />
            </div>
          </div>
        </Card>
      </div>

      {/* 투자 배너 */}
      <Card className="bg-gradient-to-r from-blue-50 to-white">
        <p className="text-lg font-bold leading-snug mb-3">
          {CUSTOMER.name} 님,<br />
          얼마 투자해볼까요?
        </p>
        <div className="flex items-center gap-2">
          <span className="text-xl font-bold text-blue-600">₩</span>
          <input
            type="text"
            placeholder="금액입력"
            className="flex-1 bg-gray-50 border border-gray-200 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-blue-200 focus:outline-none"
          />
          <span className="text-sm text-gray-500">만원</span>
          <button className="bg-blue-600 text-white px-4 py-2 rounded-lg text-sm font-bold shrink-0">
            확인
          </button>
        </div>
      </Card>
    </div>
  );
}

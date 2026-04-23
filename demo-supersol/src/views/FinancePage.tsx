"use client";

import { ChevronRight } from "lucide-react";
import Card from "@/components/Card";
import AccountCard from "@/components/AccountCard";
import { accounts, stocks } from "@/data/mock";

function formatWon(amount: number): string {
  return amount.toLocaleString("ko-KR") + "원";
}

export default function FinancePage() {
  const totalAssets = accounts.reduce((sum, a) => sum + a.balance, 0);

  return (
    <div className="bg-gray-50 min-h-screen pb-20">
      <div className="bg-white p-4">
        <div className="flex gap-6 mb-4">
          <span className="font-bold text-lg">은행</span>
          <span className="font-bold text-lg">카드</span>
          <span className="font-bold text-lg text-blue-600 border-b-2 border-blue-600 pb-1">
            증권
          </span>
          <span className="font-bold text-lg text-gray-400">보험</span>
        </div>
        <div className="mt-4">
          <p className="text-sm text-gray-500">총 자산</p>
          <div className="flex items-center gap-1">
            <h1 className="text-3xl font-bold">{formatWon(totalAssets)}</h1>
            <ChevronRight className="w-6 h-6 text-gray-400" />
          </div>
          <p className="text-gray-400 text-sm">0(0.00%)</p>

          {/* 내 신한은행 계좌 목록 */}
          <div className="mt-4 space-y-2">
            {accounts.map((account) => (
              <AccountCard key={account.accountId} account={account} />
            ))}
          </div>
        </div>
      </div>

      <div className="p-4">
        <Card className="!p-0 overflow-hidden">
          <div className="flex p-4 gap-2 border-b border-gray-100">
            <button className="flex-1 py-2 rounded-full text-sm font-medium border border-gray-200">
              보유
            </button>
            <button className="flex-1 py-2 rounded-full text-sm font-medium bg-blue-600 text-white">
              관심
            </button>
            <button className="flex-1 py-2 rounded-full text-sm font-medium border border-gray-200">
              인기
            </button>
          </div>
          <div className="p-2 space-y-1">
            {stocks.map((stock) => (
              <div key={stock.stockId} className="flex items-center justify-between p-3">
                <div className="flex items-center gap-3">
                  <div
                    className={`w-10 h-10 rounded-full flex items-center justify-center font-bold text-white ${
                      stock.changePct >= 0 ? "bg-red-500" : "bg-blue-600"
                    }`}
                  >
                    {stock.name[0]}
                  </div>
                  <span className="font-bold text-gray-800">{stock.name}</span>
                </div>
                <div className="text-right">
                  <p
                    className={`font-bold ${
                      stock.changePct >= 0 ? "text-red-500" : "text-blue-500"
                    }`}
                  >
                    {stock.market === "US" ? `$${stock.currentPrice}` : formatWon(stock.currentPrice)}
                  </p>
                  <p
                    className={`text-xs ${
                      stock.changePct >= 0 ? "text-red-500" : "text-blue-500"
                    }`}
                  >
                    {stock.changePct >= 0 ? "+" : ""}
                    {stock.changePct}%
                  </p>
                </div>
              </div>
            ))}
          </div>
          <button className="w-full py-4 text-sm text-gray-500 bg-gray-50 border-t border-gray-100 flex items-center justify-center gap-1">
            전체보기 <ChevronRight className="w-4 h-4" />
          </button>
        </Card>
      </div>
    </div>
  );
}

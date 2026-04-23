"use client";

import { ChevronRight, Settings, TrendingUp, Search } from "lucide-react";
import Card from "@/components/Card";
import { accounts, stocks } from "@/data/mock";

function formatWon(amount: number): string {
  return amount.toLocaleString("ko-KR") + "원";
}

export default function StocksPage() {
  const isaAccount = accounts.find((a) => a.accountType === "ISA");

  return (
    <div className="bg-gray-50 min-h-screen pb-20">
      <header className="bg-white p-4 flex justify-between items-center sticky top-0 z-10">
        <h1 className="text-xl font-bold">주식</h1>
        <div className="flex items-center gap-4">
          <div className="relative">
            <TrendingUp className="w-6 h-6" />
            <Search className="w-3 h-3 absolute -bottom-1 -right-1" />
          </div>
          <Settings className="w-6 h-6" />
        </div>
      </header>

      <div className="p-4 space-y-4">
        {/* 시장 지수 */}
        <div className="flex gap-3 overflow-x-auto pb-2">
          <Card className="min-w-[160px] mb-0 p-4">
            <p className="text-sm font-medium mb-1 flex items-center gap-1">
              코스피 <span className="w-2 h-2 bg-green-400 rounded-full" />
            </p>
            <p className="text-xl font-bold">2,684.32</p>
            <p className="text-xs text-red-500">+34.50 (1.30%)</p>
          </Card>
          <Card className="min-w-[160px] mb-0 p-4">
            <p className="text-sm font-medium mb-1 flex items-center gap-1">
              코스닥 <span className="w-2 h-2 bg-green-400 rounded-full" />
            </p>
            <p className="text-xl font-bold">876.52</p>
            <p className="text-xs text-red-500">+1.67 (0.19%)</p>
          </Card>
        </div>

        {/* 국내/해외 탭 */}
        <div className="flex bg-white rounded-2xl p-1">
          <button className="flex-1 py-3 text-center font-bold border-r border-gray-100 flex items-center justify-center gap-2">
            <span className="w-6 h-6 bg-red-500 rounded-full text-[10px] text-white flex items-center justify-center">
              KR
            </span>
            국내주식 <ChevronRight className="w-4 h-4 text-gray-300" />
          </button>
          <button className="flex-1 py-3 text-center font-bold flex items-center justify-center gap-2">
            <span className="w-6 h-6 bg-blue-800 rounded-full text-[10px] text-white flex items-center justify-center">
              US
            </span>
            해외주식 <ChevronRight className="w-4 h-4 text-gray-300" />
          </button>
        </div>

        {/* 내 주식 / 투자정보 탭 */}
        <div className="flex gap-8 border-b border-gray-200 px-2">
          <button className="py-3 font-bold border-b-2 border-black">내 주식</button>
          <button className="py-3 font-bold text-gray-400">투자정보</button>
        </div>

        <p className="text-center text-sm font-medium py-2">오늘의 투자정보를 가져왔어요</p>

        {/* ISA 계좌 */}
        <Card className="p-0 overflow-hidden">
          <div className="p-5">
            <div className="flex justify-between items-center mb-4">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 bg-blue-600 rounded-full flex items-center justify-center text-white">
                  S
                </div>
                <div>
                  <p className="font-bold">ISA(중개형)</p>
                  <p className="text-xs text-gray-400">
                    {isaAccount?.accountNumber}
                  </p>
                </div>
              </div>
              <ChevronRight className="w-5 h-5 text-gray-300" />
            </div>

            <div className="space-y-4">
              <div className="flex justify-between items-end">
                <div>
                  <p className="text-xs text-gray-400">평가금액</p>
                  <p className="text-2xl font-bold">
                    {isaAccount ? formatWon(isaAccount.balance) : "-"}
                  </p>
                  <p className="text-xs text-gray-300">0(0.00%)</p>
                </div>
              </div>
              <div className="flex justify-between items-center pt-2 border-t border-gray-50">
                <span className="text-xs text-gray-400">투자가능금액</span>
                <span className="font-bold">
                  {isaAccount ? formatWon(isaAccount.balance) : "-"}
                </span>
              </div>
            </div>
          </div>
          <button className="w-full py-4 bg-gray-50 text-blue-600 text-sm font-bold">
            총자산 현황
          </button>
        </Card>

        {/* 관심종목 리스트 */}
        <div>
          <h3 className="font-bold text-lg mb-3 px-1">관심종목</h3>
          <Card className="!p-0 overflow-hidden">
            <div className="divide-y divide-gray-50">
              {stocks
                .filter((s) => s.isWatchlist)
                .map((stock) => (
                  <div
                    key={stock.stockId}
                    className="flex items-center justify-between p-4"
                  >
                    <div className="flex items-center gap-3">
                      <div
                        className={`w-10 h-10 rounded-full flex items-center justify-center font-bold text-white ${
                          stock.changePct >= 0 ? "bg-red-500" : "bg-blue-600"
                        }`}
                      >
                        {stock.name[0]}
                      </div>
                      <div>
                        <p className="font-bold text-gray-800">{stock.name}</p>
                        <p className="text-xs text-gray-400">{stock.symbol}</p>
                      </div>
                    </div>
                    <div className="text-right">
                      <p
                        className={`font-bold ${
                          stock.changePct >= 0 ? "text-red-500" : "text-blue-500"
                        }`}
                      >
                        {stock.market === "US"
                          ? `$${stock.currentPrice}`
                          : formatWon(stock.currentPrice)}
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
          </Card>
        </div>
      </div>
    </div>
  );
}

"use client";

import { Settings, X, Search, ChevronRight } from "lucide-react";
import { CUSTOMER } from "@/data/mock";

const sideMenuItems = [
  "혜택",
  "포인트/미션",
  "통합금융존",
  "신한은행",
  "신한카드",
  "신한투자증권",
  "신한라이프",
  "고객센터",
];

export default function FullMenuPage() {
  return (
    <div className="bg-white min-h-screen pb-20">
      <div className="p-4 flex justify-end gap-4">
        <Settings className="w-6 h-6 text-gray-700" />
        <X className="w-6 h-6 text-gray-700" />
      </div>

      <div className="p-5">
        <h1 className="text-2xl font-bold mb-1">
          {CUSTOMER.name} 님, 즐거운 하루되세요.
        </h1>
        <p className="text-xs text-gray-400 mb-4">
          최근접속 {CUSTOMER.lastLogin}{" "}
          <span className="ml-2 underline">로그아웃</span>
        </p>

        <div className="flex gap-2 overflow-x-auto mb-6">
          <button className="bg-blue-50 text-blue-600 px-4 py-2 rounded-full text-sm font-medium flex items-center gap-1 whitespace-nowrap">
            최근 메뉴 <ChevronRight className="w-4 h-4" />
          </button>
          <button className="bg-gray-50 text-gray-600 px-4 py-2 rounded-full text-sm font-medium flex items-center gap-1 whitespace-nowrap">
            관심종목 <X className="w-3 h-3" />
          </button>
          <button className="bg-gray-50 text-gray-600 px-4 py-2 rounded-full text-sm font-medium flex items-center gap-1 whitespace-nowrap">
            보유주식 <X className="w-3 h-3" />
          </button>
        </div>

        <div className="relative mb-8">
          <input
            type="text"
            placeholder="궁금한 키워드를 검색해보세요!"
            className="w-full bg-gray-50 border-none rounded-xl p-4 text-sm focus:ring-2 focus:ring-blue-200"
          />
          <Search className="absolute right-4 top-4 w-5 h-5 text-gray-300" />
        </div>

        <div className="flex border-t border-gray-100">
          {/* 왼쪽 메뉴 */}
          <div className="w-1/3 bg-blue-600 text-white flex flex-col">
            {sideMenuItems.map((menu, i) => (
              <button
                key={i}
                className={`p-4 text-left font-bold text-sm ${
                  menu === "혜택" ? "bg-white text-blue-600" : ""
                }`}
              >
                {menu}
              </button>
            ))}
          </div>

          {/* 오른쪽 상세 */}
          <div className="flex-1 p-5 space-y-6 overflow-y-auto">
            <div>
              <div className="flex justify-between items-center border-b border-gray-100 pb-2 mb-4">
                <h3 className="font-bold text-lg">혜택</h3>
              </div>
              <div className="space-y-4">
                <div className="flex justify-between items-center">
                  <span className="font-bold">멤버십</span>
                  <ChevronRight className="w-4 h-4 rotate-90" />
                </div>
                <p className="text-sm text-gray-400 pl-2">내 SOL멤버십 등급</p>

                <div className="pt-2 border-t border-gray-50">
                  <span className="font-bold">이벤트</span>
                </div>

                <div className="pt-4 border-t border-gray-50">
                  <div className="flex justify-between items-center mb-2">
                    <span className="font-bold">쿠폰</span>
                    <ChevronRight className="w-4 h-4 rotate-90" />
                  </div>
                  <div className="space-y-2 pl-2">
                    <p className="text-sm text-gray-400">쿠폰받기</p>
                    <p className="text-sm text-gray-400">쿠폰사용하기</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

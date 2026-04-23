"use client";

import { ChevronRight, Target } from "lucide-react";
import Header from "@/components/Header";
import Card from "@/components/Card";
import ProgressBar from "@/components/ProgressBar";
import { CUSTOMER, POINT_GOAL } from "@/data/mock";

export default function BenefitsPage() {
  return (
    <div className="bg-[#f8f9ff] min-h-screen pb-20">
      <Header leftContent={<h1 className="text-xl font-bold">혜택</h1>} />

      <div className="p-5 flex justify-between items-start">
        <div>
          <h2 className="text-2xl font-bold leading-tight">
            {CUSTOMER.name} 님은
            <br />
            <span className="text-blue-600">프리미어 혜택</span>을 받고 있어요
          </h2>
          <button className="mt-2 text-sm text-gray-400 flex items-center gap-1">
            내 혜택 보기 <ChevronRight className="w-4 h-4" />
          </button>
        </div>
        <div className="relative">
          <div className="w-16 h-16 bg-blue-600 rounded-2xl rotate-12 flex items-center justify-center text-white text-3xl font-black">
            P
          </div>
          <div className="absolute -top-2 -right-1 w-6 h-6 bg-purple-500 rounded-full border-2 border-white" />
        </div>
      </div>

      <div className="px-5">
        <Card>
          <p className="text-sm text-gray-500 mb-1">마이신한포인트</p>
          <div className="flex justify-between items-center mb-4">
            <h3 className="text-2xl font-bold text-blue-600">
              {CUSTOMER.points.toLocaleString()}P
            </h3>
            <button className="bg-blue-50 text-blue-600 px-4 py-1.5 rounded-lg text-sm font-bold">
              사용하기
            </button>
          </div>
          {/* 포인트 목표 진행률 */}
          <ProgressBar
            current={CUSTOMER.points}
            target={POINT_GOAL.target}
            label={POINT_GOAL.label}
            color="bg-blue-600"
          />
          <div className="border-t border-gray-100 pt-4">
            <p className="text-sm mb-3">포인트 모으는 중</p>
            <div className="flex gap-2">
              <div className="flex-1 bg-gray-50 rounded-xl p-3 flex justify-between items-center">
                <div className="flex items-center gap-2">
                  <Target className="w-4 h-4 text-blue-500" />
                  <span className="text-sm">미션</span>
                </div>
                <span className="text-sm font-bold text-blue-600">2개</span>
              </div>
              <div className="flex-1 bg-gray-50 rounded-xl p-3 flex justify-between items-center">
                <span className="text-sm">이벤트</span>
                <span className="text-sm font-bold text-blue-600">1개</span>
              </div>
            </div>
          </div>
        </Card>

        <div className="mt-2 flex justify-center">
          <button className="text-blue-600 text-sm font-medium flex items-center gap-1">
            내 쿠폰 보러가기 <ChevronRight className="w-4 h-4" />
          </button>
        </div>

        <div className="mt-6 bg-blue-600 rounded-3xl p-6 text-white relative overflow-hidden h-40">
          <div className="relative z-10">
            <p className="text-lg font-bold leading-snug">
              상품 가입 시, 보험료 10%
              <br />
              마이신한포인트 적립
            </p>
            <p className="text-xs opacity-80 mt-2 text-blue-100">
              신한SOL 대중교통보험 mini(무배당)
            </p>
          </div>
          <div className="absolute right-[-20px] bottom-[-10px] w-40 h-24 bg-blue-400/30 rounded-full blur-2xl" />
          <div className="absolute right-4 bottom-4 w-24 h-16 bg-white/10 rounded-xl border border-white/20 backdrop-blur-sm" />
        </div>
      </div>
    </div>
  );
}

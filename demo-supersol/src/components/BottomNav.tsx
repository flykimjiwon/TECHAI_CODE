"use client";

import { Home, Wallet, Percent, TrendingUp, Menu } from "lucide-react";

type TabId = "home" | "finance" | "benefits" | "stocks" | "menu";

interface BottomNavProps {
  activeTab: TabId;
  onTabChange: (tab: TabId) => void;
}

const tabs: { id: TabId; label: string; Icon: typeof Home }[] = [
  { id: "home",     label: "홈",      Icon: Home },
  { id: "finance",  label: "금융",    Icon: Wallet },
  { id: "benefits", label: "혜택",    Icon: Percent },
  { id: "stocks",   label: "주식",    Icon: TrendingUp },
  { id: "menu",     label: "전체메뉴", Icon: Menu },
];

export default function BottomNav({ activeTab, onTabChange }: BottomNavProps) {
  return (
    <nav className="fixed bottom-0 left-0 right-0 max-w-md mx-auto bg-white border-t border-gray-100 flex justify-around items-center py-2 px-1 z-50">
      {tabs.map(({ id, label, Icon }) => (
        <button
          key={id}
          onClick={() => onTabChange(id)}
          className={`flex flex-col items-center gap-1 flex-1 py-1 ${
            activeTab === id ? "text-blue-600" : "text-gray-400"
          }`}
        >
          <Icon className="w-6 h-6" />
          <span className="text-[10px] font-bold">{label}</span>
        </button>
      ))}
    </nav>
  );
}

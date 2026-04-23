"use client";

import { useState } from "react";
import BottomNav from "@/components/BottomNav";
import HomePage from "@/views/HomePage";
import FinancePage from "@/views/FinancePage";
import BenefitsPage from "@/views/BenefitsPage";
import StocksPage from "@/views/StocksPage";
import FullMenuPage from "@/views/FullMenuPage";

type TabId = "home" | "finance" | "benefits" | "stocks" | "menu";

export default function App() {
  const [activeTab, setActiveTab] = useState<TabId>("home");

  const renderContent = () => {
    switch (activeTab) {
      case "home":
        return <HomePage />;
      case "finance":
        return <FinancePage />;
      case "benefits":
        return <BenefitsPage />;
      case "stocks":
        return <StocksPage />;
      case "menu":
        return <FullMenuPage />;
    }
  };

  return (
    <div className="max-w-md mx-auto bg-white min-h-screen relative shadow-2xl overflow-hidden font-sans text-gray-900">
      <div className="h-full overflow-y-auto no-scrollbar">
        {renderContent()}
      </div>
      <BottomNav activeTab={activeTab} onTabChange={setActiveTab} />
    </div>
  );
}

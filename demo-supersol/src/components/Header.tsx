"use client";

import { Search, Bell } from "lucide-react";
import type { ReactNode } from "react";

interface HeaderProps {
  title?: string;
  showIcons?: boolean;
  leftContent?: ReactNode;
}

export default function Header({ showIcons = true, leftContent }: HeaderProps) {
  return (
    <header className="flex items-center justify-between p-4 bg-white sticky top-0 z-10">
      <div className="flex items-center gap-2">
        {leftContent ? (
          leftContent
        ) : (
          <h1 className="text-xl font-bold italic">
            <span className="text-green-500">Super</span><span className="text-blue-600">SOL</span>
          </h1>
        )}
      </div>
      {showIcons && (
        <div className="flex items-center gap-4">
          <Search className="w-6 h-6 text-gray-700" />
          <div className="relative">
            <Bell className="w-6 h-6 text-gray-700" />
            <span className="absolute top-0 right-0 w-2 h-2 bg-red-500 rounded-full border-2 border-white" />
          </div>
        </div>
      )}
    </header>
  );
}

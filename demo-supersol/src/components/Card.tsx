"use client";

import type { ReactNode } from "react";

interface CardProps {
  children: ReactNode;
  className?: string;
}

export default function Card({ children, className = "" }: CardProps) {
  return (
    <div className={`bg-white rounded-3xl p-5 shadow-sm border border-gray-50 mb-4 ${className}`}>
      {children}
    </div>
  );
}

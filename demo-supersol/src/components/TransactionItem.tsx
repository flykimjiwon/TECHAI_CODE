"use client";

import type { Transaction } from "@/types/transaction";

interface TransactionItemProps {
  tx: Transaction;
}

function formatWon(amount: number): string {
  return amount.toLocaleString("ko-KR");
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr);
  return `${String(d.getMonth() + 1).padStart(2, "0")}.${String(d.getDate()).padStart(2, "0")}`;
}

export default function TransactionItem({ tx }: TransactionItemProps) {
  const isDeposit = tx.txType === "DEPOSIT";

  return (
    <div className="flex items-center justify-between py-2.5">
      <div className="flex items-center gap-3">
        <span className="text-xs text-gray-400 w-12">{formatDate(tx.createdAt)}</span>
        <span className="text-sm text-gray-800">{tx.description}</span>
      </div>
      <span className={`text-sm font-bold ${isDeposit ? "text-blue-600" : "text-gray-900"}`}>
        {isDeposit ? "+" : "-"}{formatWon(tx.amount)}원
      </span>
    </div>
  );
}

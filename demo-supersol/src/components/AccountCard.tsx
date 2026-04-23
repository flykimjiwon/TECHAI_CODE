"use client";

import type { Account } from "@/types/account";

interface AccountCardProps {
  account: Account;
}

function formatWon(amount: number): string {
  return amount.toLocaleString("ko-KR") + "원";
}

export default function AccountCard({ account }: AccountCardProps) {
  return (
    <div className="flex items-center justify-between p-4 border-b border-gray-100 last:border-0">
      <div className="flex items-center gap-3">
        <div className="w-10 h-10 bg-blue-600 rounded-full flex items-center justify-center text-white font-bold text-sm shrink-0">
          S
        </div>
        <div>
          <p className="text-sm font-bold text-gray-800">{account.accountName}</p>
          <p className="text-xs text-gray-400">{account.accountNumber}</p>
        </div>
      </div>
      <div className="text-right">
        <p className="font-bold text-gray-900">{formatWon(account.balance)}</p>
        <p className="text-xs text-gray-400">{account.bankName}</p>
      </div>
    </div>
  );
}

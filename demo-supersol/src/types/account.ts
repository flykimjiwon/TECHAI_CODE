export interface Account {
  accountId: number;
  accountNumber: string;
  accountName: string;
  bankName: string;
  balance: number;
  accountType: "CHECKING" | "SAVINGS" | "ISA";
  isPrimary: boolean;
}

export interface Card {
  cardId: number;
  cardName: string;
  cardNumber: string;
  monthlyUsage: number;
}

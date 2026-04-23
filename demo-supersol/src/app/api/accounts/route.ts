import { NextResponse } from "next/server";
import { query } from "@/lib/db";
import type { Account } from "@/types/account";

/**
 * GET /api/accounts?customerId=1
 *
 * 고객의 전체 계좌 목록을 조회합니다.
 */
export async function GET(request: Request) {
  const { searchParams } = new URL(request.url);
  const customerId = searchParams.get("customerId") ?? "1";

  try {
    const sql = `
      SELECT
        account_id     AS "accountId",
        account_number AS "accountNumber",
        account_name   AS "accountName",
        bank_name      AS "bankName",
        balance,
        account_type   AS "accountType",
        is_primary     AS "isPrimary"
      FROM accounts
      WHERE customer_id = $1
      ORDER BY is_primary DESC, account_name
    `;

    const rows = await query<Account>(sql, [customerId]);

    return NextResponse.json({
      success: true,
      data: rows,
    });
  } catch (error) {
    console.error("[계좌 조회 실패]", error);
    return NextResponse.json(
      { success: false, error: "계좌 조회 중 오류가 발생했습니다." },
      { status: 500 }
    );
  }
}

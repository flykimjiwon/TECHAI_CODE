import { NextResponse } from "next/server";
import { query } from "@/lib/db";
import type { Transaction } from "@/types/transaction";

// 거래내역 조회 기간 (일)
const RECENT_DAYS = 7;

/**
 * GET /api/transactions?accountId=1&category=FOOD
 *
 * 최근 거래내역을 조회합니다.
 * - 기본: 최근 7일간 거래내역
 * - category 파라미터로 카테고리 필터 가능
 */
export async function GET(request: Request) {
  const { searchParams } = new URL(request.url);
  const accountId = searchParams.get("accountId") ?? "1";
  const category = searchParams.get("category");

  try {
    let sql = `
      SELECT
        tx_id        AS "txId",
        account_id   AS "accountId",
        tx_type      AS "txType",
        amount,
        balance_after AS "balanceAfter",
        description,
        counterparty,
        category,
        created_at   AS "createdAt"
      FROM transactions
      WHERE account_id = $1
        AND created_at >= NOW() - INTERVAL '${RECENT_DAYS} days'
    `;
    const params: unknown[] = [accountId];

    if (category) {
      sql += ` AND category = $2`;
      params.push(category);
    }

    sql += ` ORDER BY created_at DESC`;

    const rows = await query<Transaction>(sql, params);

    return NextResponse.json({
      success: true,
      data: rows,
      meta: {
        accountId: Number(accountId),
        periodDays: RECENT_DAYS,
        totalCount: rows.length,
      },
    });
  } catch (error) {
    console.error("[거래내역 조회 실패]", error);
    return NextResponse.json(
      { success: false, error: "거래내역 조회 중 오류가 발생했습니다." },
      { status: 500 }
    );
  }
}

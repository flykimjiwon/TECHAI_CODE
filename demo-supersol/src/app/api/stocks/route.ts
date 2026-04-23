import { NextResponse } from "next/server";
import { query } from "@/lib/db";
import type { Stock } from "@/types/stock";

/**
 * GET /api/stocks?customerId=1&watchlist=true
 *
 * 고객의 주식 보유/관심 종목을 조회합니다.
 */
export async function GET(request: Request) {
  const { searchParams } = new URL(request.url);
  const customerId = searchParams.get("customerId") ?? "1";
  const watchlistOnly = searchParams.get("watchlist") === "true";

  try {
    let sql = `
      SELECT
        stock_id      AS "stockId",
        symbol,
        name,
        quantity,
        avg_price     AS "avgPrice",
        current_price AS "currentPrice",
        change_pct    AS "changePct",
        market,
        is_watchlist  AS "isWatchlist"
      FROM stocks
      WHERE customer_id = $1
    `;

    if (watchlistOnly) {
      sql += ` AND is_watchlist = TRUE`;
    }

    sql += ` ORDER BY market, name`;

    const rows = await query<Stock>(sql, [customerId]);

    return NextResponse.json({
      success: true,
      data: rows,
    });
  } catch (error) {
    console.error("[주식 조회 실패]", error);
    return NextResponse.json(
      { success: false, error: "주식 조회 중 오류가 발생했습니다." },
      { status: 500 }
    );
  }
}

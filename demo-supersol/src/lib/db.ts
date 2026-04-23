import { Pool } from "pg";

// 신한 SuperSOL 데모 DB 연결
// 환경변수: DATABASE_URL=postgresql://user:pass@localhost:5432/supersol_demo
const pool = new Pool({
  connectionString: process.env.DATABASE_URL,
  max: 10,
  idleTimeoutMillis: 30000,
});

export async function query<T>(text: string, params?: unknown[]): Promise<T[]> {
  const result = await pool.query(text, params);
  return result.rows as T[];
}

export default pool;

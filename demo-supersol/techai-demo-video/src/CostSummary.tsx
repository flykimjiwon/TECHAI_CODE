import { AbsoluteFill, useCurrentFrame, spring, useVideoConfig, interpolate } from "remotion";

export const CostSummary: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const fadeIn = spring({ frame, fps, config: { damping: 20 } });
  const row1 = spring({ frame: frame - 10, fps, config: { damping: 18 } });
  const row2 = spring({ frame: frame - 20, fps, config: { damping: 18 } });
  const row3 = spring({ frame: frame - 30, fps, config: { damping: 18 } });
  const totalRow = spring({ frame: frame - 45, fps, config: { damping: 15 } });

  return (
    <AbsoluteFill
      style={{
        background: "#0a0a0a",
        justifyContent: "center",
        alignItems: "center",
        fontFamily: "'Pretendard Variable', system-ui, sans-serif",
      }}
    >
      {/* 헤더 */}
      <div style={{ position: "absolute", top: 50, textAlign: "center", opacity: fadeIn }}>
        <p style={{ fontSize: 24, color: "#666", letterSpacing: "0.1em" }}>
          본 시연은 2025년 8월 5일 공개된
        </p>
        <h1 style={{ fontSize: 52, fontWeight: 800, color: "#60A5FA", marginTop: 8 }}>
          GPT-OSS-120B
        </h1>
        <p style={{ fontSize: 20, color: "#555", marginTop: 4 }}>
          오픈소스 120B MoE (5.1B active)
        </p>
      </div>

      {/* 비용 테이블 — 2배 확대 */}
      <div style={{ marginTop: 60, width: 1200 }}>
        {/* 헤더 */}
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            padding: "16px 24px",
            borderBottom: "2px solid #333",
            opacity: fadeIn,
          }}
        >
          <span style={{ fontSize: 24, color: "#666", fontWeight: 600, width: 460 }}>시나리오</span>
          <span style={{ fontSize: 24, color: "#666", fontWeight: 600, width: 200, textAlign: "right" }}>토큰</span>
          <span style={{ fontSize: 24, color: "#666", fontWeight: 600, width: 160, textAlign: "right" }}>소요시간</span>
          <span style={{ fontSize: 24, color: "#666", fontWeight: 600, width: 160, textAlign: "right" }}>비용</span>
        </div>

        {/* 시나리오 1 */}
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            padding: "24px 24px",
            borderBottom: "1px solid #1a1a1a",
            opacity: Math.max(0, row1),
            transform: `translateX(${interpolate(Math.max(0, row1), [0, 1], [30, 0])}px)`,
          }}
        >
          <span style={{ fontSize: 30, color: "#f0f0f0", fontWeight: 600, width: 460 }}>
            1. 홈 — 거래 미리보기 추가
          </span>
          <span style={{ fontSize: 30, color: "#94A3B8", width: 200, textAlign: "right", fontFamily: "'JetBrains Mono', monospace" }}>
            22,302
          </span>
          <span style={{ fontSize: 30, color: "#94A3B8", width: 160, textAlign: "right" }}>~45초</span>
          <span style={{ fontSize: 30, color: "#34D399", fontWeight: 700, width: 160, textAlign: "right" }}>$0.007</span>
        </div>

        {/* 시나리오 2 */}
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            padding: "24px 24px",
            borderBottom: "1px solid #1a1a1a",
            opacity: Math.max(0, row2),
            transform: `translateX(${interpolate(Math.max(0, row2), [0, 1], [30, 0])}px)`,
          }}
        >
          <span style={{ fontSize: 30, color: "#f0f0f0", fontWeight: 600, width: 460 }}>
            2. 금융 — 계좌 목록 추가
          </span>
          <span style={{ fontSize: 30, color: "#94A3B8", width: 200, textAlign: "right", fontFamily: "'JetBrains Mono', monospace" }}>
            13,128
          </span>
          <span style={{ fontSize: 30, color: "#94A3B8", width: 160, textAlign: "right" }}>~30초</span>
          <span style={{ fontSize: 30, color: "#34D399", fontWeight: 700, width: 160, textAlign: "right" }}>$0.004</span>
        </div>

        {/* 시나리오 3 */}
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            padding: "24px 24px",
            borderBottom: "2px solid #333",
            opacity: Math.max(0, row3),
            transform: `translateX(${interpolate(Math.max(0, row3), [0, 1], [30, 0])}px)`,
          }}
        >
          <span style={{ fontSize: 30, color: "#f0f0f0", fontWeight: 600, width: 460 }}>
            3. 혜택 — 프로그레스바 추가
          </span>
          <span style={{ fontSize: 30, color: "#94A3B8", width: 200, textAlign: "right", fontFamily: "'JetBrains Mono', monospace" }}>
            15,999
          </span>
          <span style={{ fontSize: 30, color: "#94A3B8", width: 160, textAlign: "right" }}>~25초</span>
          <span style={{ fontSize: 30, color: "#34D399", fontWeight: 700, width: 160, textAlign: "right" }}>$0.005</span>
        </div>

        {/* 합계 */}
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            padding: "28px 24px",
            opacity: Math.max(0, totalRow),
            transform: `scale(${interpolate(Math.max(0, totalRow), [0, 1], [0.95, 1])})`,
          }}
        >
          <span style={{ fontSize: 36, color: "#f0f0f0", fontWeight: 800, width: 460 }}>합계</span>
          <span style={{ fontSize: 36, color: "#f0f0f0", fontWeight: 700, width: 200, textAlign: "right", fontFamily: "'JetBrains Mono', monospace" }}>
            51,429
          </span>
          <span style={{ fontSize: 36, color: "#f0f0f0", fontWeight: 700, width: 160, textAlign: "right" }}>~100초</span>
          <span style={{ fontSize: 44, color: "#34D399", fontWeight: 900, width: 160, textAlign: "right" }}>$0.016</span>
        </div>
      </div>

      {/* 하단 요약 */}
      <div style={{ position: "absolute", bottom: 50, textAlign: "center", opacity: Math.max(0, totalRow) }}>
        <p style={{ fontSize: 28, color: "#666" }}>
          UI 수정 3건, 총 <span style={{ color: "#34D399", fontWeight: 800 }}>약 23원</span> — 커피 한 잔이면 100건 이상 가능
        </p>
      </div>
    </AbsoluteFill>
  );
};

import { AbsoluteFill, useCurrentFrame, spring, useVideoConfig, interpolate } from "remotion";

interface RowProps {
  task: string;
  oss: string;
  opus: string;
  gap: string;
  maxGap: string;
  delay: number;
}

const GapRow: React.FC<RowProps> = ({ task, oss, opus, gap, maxGap, delay }) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const anim = spring({ frame: frame - delay, fps, config: { damping: 18 } });
  const x = interpolate(Math.max(0, anim), [0, 1], [40, 0]);

  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        padding: "24px 0",
        borderBottom: "1px solid #1a1a1a",
        opacity: Math.max(0, anim),
        transform: `translateX(${x}px)`,
      }}
    >
      <span style={{ width: 360, fontSize: 28, color: "#f0f0f0", fontWeight: 600 }}>{task}</span>
      <span style={{ width: 340, fontSize: 24, color: "#EF4444" }}>{oss}</span>
      <span style={{ width: 300, fontSize: 24, color: "#34D399" }}>{opus}</span>
      <span style={{ width: 160, fontSize: 28, color: "#94A3B8", fontWeight: 700, textAlign: "center" }}>{gap}</span>
      <span
        style={{
          width: 140,
          fontSize: 36,
          fontWeight: 900,
          color: "#F59E0B",
          textAlign: "center",
          fontFamily: "'JetBrains Mono', monospace",
        }}
      >
        {maxGap}
      </span>
    </div>
  );
};

export const RealWorldGap: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const fadeIn = spring({ frame, fps, config: { damping: 20 } });
  const conclusionAnim = spring({ frame: frame - 120, fps, config: { damping: 12 } });
  const conclusionScale = interpolate(Math.max(0, conclusionAnim), [0, 1], [0.85, 1]);

  return (
    <AbsoluteFill
      style={{
        background: "#0a0a0a",
        fontFamily: "'Pretendard Variable', system-ui, sans-serif",
        padding: "36px 60px",
      }}
    >
      {/* 헤더 */}
      <div style={{ textAlign: "center", opacity: fadeIn, marginBottom: 20 }}>
        <h1 style={{ fontSize: 40, fontWeight: 800, color: "#f0f0f0" }}>
          오늘 시연 기준 — <span style={{ color: "#60A5FA" }}>실사용 체감 차이</span>
        </h1>
      </div>

      {/* 테이블 헤더 */}
      <div
        style={{
          display: "flex",
          alignItems: "center",
          padding: "16px 0",
          borderBottom: "2px solid #333",
          opacity: fadeIn,
        }}
      >
        <span style={{ width: 360, fontSize: 22, color: "#666", fontWeight: 600 }}>작업 유형</span>
        <span style={{ width: 340, fontSize: 22, color: "#EF4444", fontWeight: 600 }}>OSS-120B</span>
        <span style={{ width: 300, fontSize: 22, color: "#34D399", fontWeight: 600 }}>Opus 4.7급</span>
        <span style={{ width: 160, fontSize: 22, color: "#666", fontWeight: 600, textAlign: "center" }}>체감 차이</span>
        <span style={{ width: 140, fontSize: 22, color: "#F59E0B", fontWeight: 700, textAlign: "center" }}>최대</span>
      </div>

      {/* 행들 */}
      <GapRow
        task="단순 UI 수정 (1파일)"
        oss="됨 (file_write 강제)"
        opus="됨 (더 빠르고 정확)"
        gap="2~3배"
        maxGap="×3"
        delay={15}
      />
      <GapRow
        task="다중 파일 수정"
        oss="불안정 (patch 실패 반복)"
        opus="한 번에 성공"
        gap="5~10배"
        maxGap="×10"
        delay={30}
      />
      <GapRow
        task="복잡한 리팩토링"
        oss="거의 불가 (빈 응답, 루프)"
        opus="안정적으로 수행"
        gap="10~20배"
        maxGap="×20"
        delay={45}
      />
      <GapRow
        task="대규모 코드베이스 이해"
        oss="131K 한계, 파일 다 못 읽음"
        opus="1M에 프로젝트 전체 적재"
        gap="측정 불가"
        maxGap="∞"
        delay={60}
      />
      <GapRow
        task="Reasoning 품질"
        oss="같은 말 반복, 영어로 생각"
        opus="구조적, 한국어 대응"
        gap="~5배"
        maxGap="×5"
        delay={75}
      />

      {/* 결론 */}
      <div
        style={{
          display: "flex",
          justifyContent: "center",
          marginTop: 36,
          opacity: Math.max(0, conclusionAnim),
          transform: `scale(${conclusionScale})`,
        }}
      >
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 28,
            padding: "24px 56px",
            background: "rgba(245,158,11,0.08)",
            border: "1px solid rgba(245,158,11,0.3)",
            borderRadius: 16,
          }}
        >
          <span style={{ fontSize: 28, color: "#f0f0f0", fontWeight: 600 }}>
            작업 난이도에 따라
          </span>
          <span
            style={{
              fontSize: 52,
              fontWeight: 900,
              color: "#F59E0B",
              fontFamily: "'JetBrains Mono', monospace",
            }}
          >
            ×3 ~ ∞
          </span>
          <span style={{ fontSize: 28, color: "#f0f0f0", fontWeight: 600 }}>
            생산성 차이
          </span>
        </div>
      </div>
    </AbsoluteFill>
  );
};

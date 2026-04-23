import { AbsoluteFill, useCurrentFrame, spring, useVideoConfig, interpolate } from "remotion";

const BAR_HEIGHT = 36;

interface BarProps {
  label: string;
  ossValue: number;
  opusValue: number;
  unit: string;
  multiplier: string;
  delay: number;
}

const ComparisonBar: React.FC<BarProps> = ({ label, ossValue, opusValue, unit, multiplier, delay }) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const anim = spring({ frame: frame - delay, fps, config: { damping: 15 } });
  const ossWidth = interpolate(Math.max(0, anim), [0, 1], [0, (ossValue / opusValue) * 400]);
  const opusWidth = interpolate(Math.max(0, anim), [0, 1], [0, 400]);

  return (
    <div style={{ marginBottom: 28, opacity: Math.max(0, anim) }}>
      <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 6 }}>
        <span style={{ fontSize: 16, color: "#94A3B8", fontWeight: 600 }}>{label}</span>
        <span style={{ fontSize: 14, color: "#F59E0B", fontWeight: 700 }}>{multiplier}</span>
      </div>
      {/* OSS bar */}
      <div style={{ display: "flex", alignItems: "center", gap: 12, marginBottom: 4 }}>
        <span style={{ fontSize: 13, color: "#666", width: 80, textAlign: "right" }}>OSS-120B</span>
        <div style={{ flex: 1, position: "relative", height: BAR_HEIGHT }}>
          <div style={{ position: "absolute", top: 0, left: 0, width: "100%", height: "100%", background: "#1a1a1a", borderRadius: 6 }} />
          <div style={{ position: "absolute", top: 0, left: 0, width: ossWidth, height: "100%", background: "#3B82F6", borderRadius: 6 }} />
          <span style={{ position: "absolute", right: 8, top: "50%", transform: "translateY(-50%)", fontSize: 14, color: "#f0f0f0", fontWeight: 700 }}>
            {ossValue.toLocaleString()}{unit}
          </span>
        </div>
      </div>
      {/* Opus bar */}
      <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
        <span style={{ fontSize: 13, color: "#666", width: 80, textAlign: "right" }}>Opus 4.7</span>
        <div style={{ flex: 1, position: "relative", height: BAR_HEIGHT }}>
          <div style={{ position: "absolute", top: 0, left: 0, width: "100%", height: "100%", background: "#1a1a1a", borderRadius: 6 }} />
          <div style={{ position: "absolute", top: 0, left: 0, width: opusWidth, height: "100%", background: "#F59E0B", borderRadius: 6 }} />
          <span style={{ position: "absolute", right: 8, top: "50%", transform: "translateY(-50%)", fontSize: 14, color: "#f0f0f0", fontWeight: 700 }}>
            {opusValue.toLocaleString()}{unit}
          </span>
        </div>
      </div>
    </div>
  );
};

export const ModelComparison: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const fadeIn = spring({ frame, fps, config: { damping: 20 } });
  const priceAnim = spring({ frame: frame - 90, fps, config: { damping: 15 } });

  return (
    <AbsoluteFill
      style={{
        background: "#0a0a0a",
        fontFamily: "'Pretendard Variable', system-ui, sans-serif",
        padding: "50px 100px",
      }}
    >
      {/* 헤더 */}
      <div style={{ textAlign: "center", opacity: fadeIn, marginBottom: 30 }}>
        <p style={{ fontSize: 14, color: "#555", letterSpacing: "0.15em" }}>MODEL COMPARISON · 2026.04.22</p>
        <div style={{ display: "flex", justifyContent: "center", alignItems: "center", gap: 20, marginTop: 12 }}>
          <span style={{ fontSize: 28, fontWeight: 800, color: "#3B82F6" }}>gpt-oss-120b</span>
          <span style={{ fontSize: 20, color: "#444" }}>vs</span>
          <span style={{ fontSize: 28, fontWeight: 800, color: "#F59E0B" }}>Claude Opus 4.7</span>
        </div>
      </div>

      {/* 좌우 2컬럼 */}
      <div style={{ display: "flex", gap: 60 }}>
        {/* 왼쪽: 성능 비교 바 */}
        <div style={{ flex: 1 }}>
          <p style={{ fontSize: 16, color: "#666", fontWeight: 600, marginBottom: 20, letterSpacing: "0.05em" }}>
            코딩 · 개발 성능
          </p>

          <ComparisonBar
            label="SWE-bench Verified (GitHub 이슈 해결)"
            ossValue={62} opusValue={87.6} unit="%" multiplier="+25.6p" delay={15}
          />
          <ComparisonBar
            label="SWE-bench Pro (실무 이슈)"
            ossValue={30} opusValue={64.3} unit="%" multiplier="2.1배" delay={30}
          />
          <ComparisonBar
            label="컨텍스트 윈도우"
            ossValue={131} opusValue={1000} unit="K" multiplier="7.6배" delay={45}
          />
          <ComparisonBar
            label="종합 지능 지수 (AA Intelligence)"
            ossValue={33} opusValue={70} unit="" multiplier="2.1배" delay={60}
          />
        </div>

        {/* 오른쪽: 컨텍스트 + 가격 */}
        <div style={{ width: 420 }}>
          {/* 컨텍스트 면적 비교 */}
          <p style={{ fontSize: 16, color: "#666", fontWeight: 600, marginBottom: 20, letterSpacing: "0.05em" }}>
            컨텍스트 윈도우 스케일
          </p>
          <div style={{ position: "relative", height: 200, marginBottom: 30 }}>
            {/* Opus 큰 원 */}
            <div
              style={{
                position: "absolute",
                top: 10,
                left: 80,
                width: 180,
                height: 180,
                borderRadius: "50%",
                border: "2px solid #F59E0B",
                background: "rgba(245,158,11,0.08)",
                display: "flex",
                flexDirection: "column",
                justifyContent: "center",
                alignItems: "center",
                opacity: Math.max(0, spring({ frame: frame - 50, fps, config: { damping: 15 } })),
              }}
            >
              <span style={{ fontSize: 28, fontWeight: 900, color: "#F59E0B" }}>1,000K</span>
              <span style={{ fontSize: 12, color: "#888" }}>Opus 4.7</span>
            </div>
            {/* OSS 작은 원 */}
            <div
              style={{
                position: "absolute",
                top: 75,
                left: 290,
                width: 70,
                height: 70,
                borderRadius: "50%",
                border: "2px solid #3B82F6",
                background: "rgba(59,130,246,0.15)",
                display: "flex",
                flexDirection: "column",
                justifyContent: "center",
                alignItems: "center",
                opacity: Math.max(0, spring({ frame: frame - 40, fps, config: { damping: 15 } })),
              }}
            >
              <span style={{ fontSize: 14, fontWeight: 900, color: "#3B82F6" }}>131K</span>
              <span style={{ fontSize: 9, color: "#888" }}>OSS</span>
            </div>
          </div>

          {/* 가격 비교 */}
          <p style={{ fontSize: 16, color: "#666", fontWeight: 600, marginBottom: 16, letterSpacing: "0.05em" }}>
            API 가격 (per 1M 토큰)
          </p>
          <div style={{ opacity: Math.max(0, priceAnim) }}>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", padding: "14px 0", borderBottom: "1px solid #1a1a1a" }}>
              <span style={{ fontSize: 15, color: "#94A3B8" }}>gpt-oss-120b</span>
              <span style={{ fontSize: 22, fontWeight: 800, color: "#34D399", fontFamily: "'JetBrains Mono', monospace" }}>
                $0.04 / $0.19
              </span>
            </div>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", padding: "14px 0", borderBottom: "1px solid #1a1a1a" }}>
              <span style={{ fontSize: 15, color: "#94A3B8" }}>Claude Opus 4.7</span>
              <span style={{ fontSize: 22, fontWeight: 800, color: "#F87171", fontFamily: "'JetBrains Mono', monospace" }}>
                $5.00 / $25.00
              </span>
            </div>
            <div style={{ display: "flex", justifyContent: "center", marginTop: 16 }}>
              <div
                style={{
                  padding: "10px 28px",
                  background: "rgba(52,211,153,0.1)",
                  border: "1px solid #34D399",
                  borderRadius: 10,
                }}
              >
                <span style={{ fontSize: 20, fontWeight: 800, color: "#34D399" }}>
                  128배 저렴
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* 하단 출처 */}
      <div style={{ position: "absolute", bottom: 30, left: 0, right: 0, textAlign: "center" }}>
        <p style={{ fontSize: 12, color: "#444" }}>
          Sources: Anthropic · OpenAI · Artificial Analysis · Vellum · For 지원 / AI Champion
        </p>
      </div>
    </AbsoluteFill>
  );
};

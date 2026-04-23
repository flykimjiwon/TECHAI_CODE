import { AbsoluteFill, Img, useCurrentFrame, interpolate, spring, useVideoConfig, staticFile } from "remotion";

interface TitleSlideProps {
  isEnding?: boolean;
}

export const TitleSlide: React.FC<TitleSlideProps> = ({ isEnding = false }) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const logoOpacity = spring({ frame, fps, config: { damping: 20 } });
  const subtitleOpacity = spring({ frame: frame - 20, fps, config: { damping: 20 } });
  const disclaimerOpacity = spring({ frame: frame - 40, fps, config: { damping: 20 } });

  const logoY = interpolate(logoOpacity, [0, 1], [30, 0]);
  const subtitleY = interpolate(Math.max(0, subtitleOpacity), [0, 1], [20, 0]);

  return (
    <AbsoluteFill
      style={{
        background: "#0a0a0a",
        justifyContent: "center",
        alignItems: "center",
        fontFamily: "Pretendard, 'Pretendard Variable', system-ui, sans-serif",
      }}
    >
      <div
        style={{
          position: "absolute",
          top: "15%",
          left: "20%",
          width: 500,
          height: 500,
          borderRadius: "50%",
          background: "radial-gradient(circle, rgba(59,130,246,0.08) 0%, transparent 70%)",
        }}
      />

      {/* ASCII 로고 — PNG 이미지로 교체 */}
      <div
        style={{
          opacity: logoOpacity,
          transform: `translateY(${logoY}px)`,
          textAlign: "center",
        }}
      >
        <Img
          src={staticFile("assets/techai-logo.png")}
          style={{ width: 720, height: "auto" }}
        />
      </div>

      {/* "시연" / "감사합니다" */}
      <div
        style={{
          opacity: Math.max(0, subtitleOpacity),
          transform: `translateY(${subtitleY}px)`,
          textAlign: "center",
          marginTop: 32,
        }}
      >
        <h1
          style={{
            fontSize: 63,
            fontWeight: 800,
            color: "#f0f0f0",
            letterSpacing: "0.04em",
          }}
        >
          {isEnding ? "감사합니다" : "시연"}
        </h1>

        <div
          style={{
            display: "inline-flex",
            alignItems: "center",
            gap: 16,
            marginTop: 28,
            padding: "16px 44px",
            border: "1px solid #3b82f6",
            borderRadius: 10,
            background: "rgba(59,130,246,0.08)",
          }}
        >
          <span style={{ fontSize: 30, fontWeight: 600, color: "#888", letterSpacing: "0.06em" }}>
            Tech혁신Unit
          </span>
          <span
            style={{
              fontSize: 33,
              fontWeight: 700,
              color: "#3b82f6",
              letterSpacing: "0.08em",
              fontFamily: "'JetBrains Mono', monospace",
            }}
          >
            개발 AX CELL
          </span>
        </div>
      </div>

      {/* 면책 자막 (오프닝만) */}
      {!isEnding && (
        <div
          style={{
            position: "absolute",
            bottom: 30,
            opacity: Math.max(0, disclaimerOpacity),
            textAlign: "center",
          }}
        >
          <p
            style={{
              fontSize: 20,
              color: "#555",
              lineHeight: 1.4,
            }}
          >
            시연을 위해 데모로 만들어진 슈퍼솔 클론 페이지입니다, 일부 가상 데이터를 넣어두었습니다
          </p>
        </div>
      )}
    </AbsoluteFill>
  );
};

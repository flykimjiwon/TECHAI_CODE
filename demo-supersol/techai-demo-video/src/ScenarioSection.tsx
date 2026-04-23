import {
  AbsoluteFill,
  OffthreadVideo,
  Img,
  useCurrentFrame,
  interpolate,
  spring,
  useVideoConfig,
  staticFile,
  Sequence,
} from "remotion";

interface ScenarioSectionProps {
  videoSrc: string;
  beforeImg: string;
  afterImg: string;
  scenarioTitle: string;
  scenarioSubtitle: string;
  envLabel: string;
  completionLabel: string;
  videoStartSec: number;
  videoDuration: number;
  resultDuration: number;
  comparisonDuration: number;
}

export const ScenarioSection: React.FC<ScenarioSectionProps> = ({
  videoSrc,
  beforeImg,
  afterImg,
  scenarioTitle,
  scenarioSubtitle,
  envLabel,
  completionLabel,
  videoStartSec,
  videoDuration,
  resultDuration,
  comparisonDuration,
}) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  // 마지막 3초 = 90프레임은 정배속, 나머지는 3배속
  const normalSpeedFrames = 3 * fps; // 90프레임
  const fastPhaseFrames = videoDuration - normalSpeedFrames;

  const isVideoPhase = frame < videoDuration;
  const isResultPhase = frame >= videoDuration && frame < videoDuration + resultDuration;
  const compFrame = frame - videoDuration - resultDuration;

  const titleOpacity = spring({ frame, fps, config: { damping: 20 } });

  // ── 영상 재생 구간 ──
  if (isVideoPhase) {
    const isFastPhase = frame < fastPhaseFrames;

    return (
      <AbsoluteFill style={{ background: "#0a0a0a", fontFamily: "Pretendard, 'Pretendard Variable', system-ui, sans-serif" }}>
        {/* 시나리오 제목 — 1.5배 크기 */}
        <div
          style={{
            position: "absolute",
            top: 30,
            left: 0,
            right: 0,
            zIndex: 10,
            textAlign: "center",
            opacity: titleOpacity,
          }}
        >
          <div
            style={{
              display: "inline-flex",
              flexDirection: "column",
              alignItems: "center",
              padding: "16px 40px",
              background: "rgba(0,0,0,0.7)",
              borderRadius: 16,
              backdropFilter: "blur(8px)",
            }}
          >
            <span style={{ fontSize: 42, fontWeight: 800, color: "#60A5FA" }}>
              {scenarioTitle}
            </span>
            <span style={{ fontSize: 27, color: "#ffffff", marginTop: 6 }}>
              {scenarioSubtitle}
            </span>
          </div>
        </div>

        {/* 영상: 3배속 구간 + 정배속 구간을 Sequence로 분리 */}
        {isFastPhase ? (
          <OffthreadVideo
            src={staticFile(videoSrc)}
            startFrom={videoStartSec * fps}
            playbackRate={3}
            style={{ width: "100%", height: "100%", objectFit: "contain" }}
          />
        ) : (
          <OffthreadVideo
            src={staticFile(videoSrc)}
            startFrom={videoStartSec * fps + fastPhaseFrames * 3}
            playbackRate={1}
            style={{ width: "100%", height: "100%", objectFit: "contain" }}
          />
        )}

        {/* 환경 라벨 — 1.5배 (33px), 흰색 */}
        <div
          style={{
            position: "absolute",
            bottom: 30,
            left: 0,
            right: 0,
            textAlign: "center",
          }}
        >
          <span
            style={{
              fontSize: 33,
              color: "#ffffff",
              background: "rgba(0,0,0,0.6)",
              padding: "10px 24px",
              borderRadius: 10,
            }}
          >
            {envLabel}
          </span>
        </div>
      </AbsoluteFill>
    );
  }

  // ── 완성 결과 유지 구간 (5초) ──
  if (isResultPhase) {
    const resultFrame = frame - videoDuration;
    const badgeOpacity = spring({ frame: resultFrame, fps, config: { damping: 15 } });
    const badgeScale = interpolate(badgeOpacity, [0, 1], [0.8, 1]);

    return (
      <AbsoluteFill style={{ background: "#0a0a0a", fontFamily: "Pretendard, 'Pretendard Variable', system-ui, sans-serif" }}>
        <OffthreadVideo
          src={staticFile(videoSrc)}
          startFrom={videoStartSec * fps + fastPhaseFrames * 3 + normalSpeedFrames}
          playbackRate={1}
          style={{ width: "100%", height: "100%", objectFit: "contain" }}
        />

        <div
          style={{
            position: "absolute",
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            display: "flex",
            justifyContent: "center",
            alignItems: "center",
            background: "rgba(0,0,0,0.4)",
          }}
        >
          <div
            style={{
              opacity: badgeOpacity,
              transform: `scale(${badgeScale})`,
              display: "flex",
              flexDirection: "column",
              alignItems: "center",
              padding: "28px 60px",
              background: "rgba(0,0,0,0.85)",
              borderRadius: 20,
              border: "2px solid #3B82F6",
              boxShadow: "0 0 60px rgba(59,130,246,0.3)",
            }}
          >
            <span style={{ fontSize: 24, color: "#34D399", fontWeight: 700 }}>✓</span>
            <span
              style={{
                fontSize: 32,
                fontWeight: 800,
                color: "#ffffff",
                marginTop: 8,
                fontFamily: "Pretendard, 'Pretendard Variable', system-ui, sans-serif",
              }}
            >
              {completionLabel}
            </span>
            <span style={{ fontSize: 18, color: "#ffffff", marginTop: 6, opacity: 0.7 }}>
              {scenarioSubtitle}
            </span>
          </div>
        </div>
      </AbsoluteFill>
    );
  }

  // ── 비교 사진 구간 (8초) ──
  const compOpacity = spring({ frame: compFrame, fps, config: { damping: 15 } });
  const slideIn = interpolate(compOpacity, [0, 1], [60, 0]);

  return (
    <AbsoluteFill
      style={{
        background: "linear-gradient(135deg, #0a0a0a 0%, #1a1a2e 100%)",
        justifyContent: "center",
        alignItems: "center",
        padding: 60,
        fontFamily: "Pretendard, 'Pretendard Variable', system-ui, sans-serif",
      }}
    >
      {/* 상단: 시나리오 제목만 (파란색) */}
      <div
        style={{
          position: "absolute",
          top: 40,
          textAlign: "center",
          opacity: compOpacity,
        }}
      >
        <h2 style={{ fontSize: 48, fontWeight: 800, color: "#60A5FA" }}>
          {scenarioTitle}
        </h2>
      </div>

      {/* 좌우 비교 */}
      <div
        style={{
          display: "flex",
          gap: 60,
          alignItems: "center",
          opacity: compOpacity,
          transform: `translateY(${slideIn}px)`,
          marginTop: 40,
        }}
      >
        <div style={{ textAlign: "center" }}>
          <div
            style={{
              border: "3px solid #374151",
              borderRadius: 24,
              overflow: "hidden",
              boxShadow: "0 20px 60px rgba(0,0,0,0.5)",
            }}
          >
            <Img
              src={staticFile(beforeImg)}
              style={{ width: 420, height: "auto", display: "block" }}
            />
          </div>
          <p style={{ fontSize: 28, color: "#ffffff", marginTop: 20, fontWeight: 600, opacity: 0.6 }}>
            변경 전
          </p>
        </div>

        <div style={{ fontSize: 60, color: "#60A5FA", fontWeight: 900 }}>→</div>

        <div style={{ textAlign: "center" }}>
          <div
            style={{
              border: "3px solid #3B82F6",
              borderRadius: 24,
              overflow: "hidden",
              boxShadow: "0 20px 60px rgba(59,130,246,0.3)",
            }}
          >
            <Img
              src={staticFile(afterImg)}
              style={{ width: 420, height: "auto", display: "block" }}
            />
          </div>
          <p style={{ fontSize: 28, color: "#60A5FA", marginTop: 20, fontWeight: 600 }}>
            변경 후
          </p>
        </div>
      </div>
    </AbsoluteFill>
  );
};

import { AbsoluteFill, Sequence } from "remotion";
import { TitleSlide } from "./TitleSlide";
import { ScenarioSection } from "./ScenarioSection";
import { CostSummary } from "./CostSummary";
import { ModelComparison } from "./ModelComparison";
import { RealWorldGap } from "./RealWorldGap";

const FPS = 30;

// 오프닝: 5초
const OPENING_DURATION = 5 * FPS;

// 시나리오 1 — 홈: 영상 10초 + 완성 5초 + 비교 8초 = 23초
const S1_VIDEO = 10 * FPS;
const S1_RESULT = 5 * FPS;
const S1_COMPARE = 8 * FPS;
const S1_TOTAL = S1_VIDEO + S1_RESULT + S1_COMPARE;

// 시나리오 2 — 금융
const S2_VIDEO = 10 * FPS;
const S2_RESULT = 5 * FPS;
const S2_COMPARE = 8 * FPS;
const S2_TOTAL = S2_VIDEO + S2_RESULT + S2_COMPARE;

// 시나리오 3 — 혜택
const S3_VIDEO = 10 * FPS;
const S3_RESULT = 5 * FPS;
const S3_COMPARE = 8 * FPS;
const S3_TOTAL = S3_VIDEO + S3_RESULT + S3_COMPARE;

// 비용 요약: 8초
const COST_DURATION = 8 * FPS;

// 모델 비교: 10초
const COMPARISON_DURATION = 10 * FPS;

// 실사용 체감 차이: 10초
const REALWORLD_DURATION = 10 * FPS;

// 엔딩: 5초
const ENDING_DURATION = 5 * FPS;

export const DemoVideo: React.FC = () => {
  let offset = 0;

  return (
    <AbsoluteFill style={{ background: "#0a0a0a" }}>
      {/* 오프닝 */}
      <Sequence from={offset} durationInFrames={OPENING_DURATION}>
        <TitleSlide />
      </Sequence>
      {(offset += OPENING_DURATION)}

      {/* 시나리오 1: 홈 */}
      <Sequence from={offset} durationInFrames={S1_TOTAL}>
        <ScenarioSection
          videoSrc="assets/홈화면변경영상.mov"
          beforeImg="assets/홈변경전.png"
          afterImg="assets/홈변경후.png"
          scenarioTitle="시나리오 1: 홈"
          scenarioSubtitle="계좌 카드에 최근 거래 미리보기 추가"
          envLabel="터미널 CLI에서 실행한 모습"
          completionLabel="약 45초 만에 완성"
          videoStartSec={2}
          videoDuration={S1_VIDEO}
          resultDuration={S1_RESULT}
          comparisonDuration={S1_COMPARE}
        />
      </Sequence>
      {(offset += S1_TOTAL)}

      {/* 시나리오 2: 금융 */}
      <Sequence from={offset} durationInFrames={S2_TOTAL}>
        <ScenarioSection
          videoSrc="assets/금융변경영상.mov"
          beforeImg="assets/금융변경전.png"
          afterImg="assets/금융변경후.png"
          scenarioTitle="시나리오 2: 금융"
          scenarioSubtitle="총 자산 아래에 내 계좌 목록 추가"
          envLabel="터미널 CLI에서 실행한 모습"
          completionLabel="약 30초 만에 완성"
          videoStartSec={3}
          videoDuration={S2_VIDEO}
          resultDuration={S2_RESULT}
          comparisonDuration={S2_COMPARE}
        />
      </Sequence>
      {(offset += S2_TOTAL)}

      {/* 시나리오 3: 혜택 */}
      <Sequence from={offset} durationInFrames={S3_TOTAL}>
        <ScenarioSection
          videoSrc="assets/혜택변경영상.mov"
          beforeImg="assets/혜택변경전.png"
          afterImg="assets/혜택변경후.png"
          scenarioTitle="시나리오 3: 혜택"
          scenarioSubtitle="적립 목표 포인트 프로그레스바 추가"
          envLabel="VSCode IDE 터미널에서 실행한 모습"
          completionLabel="약 25초 만에 완성"
          videoStartSec={2}
          videoDuration={S3_VIDEO}
          resultDuration={S3_RESULT}
          comparisonDuration={S3_COMPARE}
        />
      </Sequence>
      {(offset += S3_TOTAL)}

      {/* 비용 요약 */}
      <Sequence from={offset} durationInFrames={COST_DURATION}>
        <CostSummary />
      </Sequence>
      {(offset += COST_DURATION)}

      {/* 모델 비교 */}
      <Sequence from={offset} durationInFrames={COMPARISON_DURATION}>
        <ModelComparison />
      </Sequence>
      {(offset += COMPARISON_DURATION)}

      {/* 실사용 체감 차이 */}
      <Sequence from={offset} durationInFrames={REALWORLD_DURATION}>
        <RealWorldGap />
      </Sequence>
      {(offset += REALWORLD_DURATION)}

      {/* 엔딩 */}
      <Sequence from={offset} durationInFrames={ENDING_DURATION}>
        <TitleSlide isEnding />
      </Sequence>
    </AbsoluteFill>
  );
};

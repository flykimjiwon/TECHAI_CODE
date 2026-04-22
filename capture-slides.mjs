import puppeteer from 'puppeteer';
import { resolve } from 'path';

const htmlPath = resolve('index.html');
const outputPath = resolve('techaicode.pdf');

const BASE_W = 1920;
const BASE_H = 1080;
const DPR = 2;

// 슬라이드별 CSS zoom 배율 (0-indexed)
const ZOOM_MAP = {
  0: 1.5, 1: 1.5, 2: 1.5,       // 1,2,3페이지
  3: 1.2,                         // 4페이지
  6: 1.2,                         // 7페이지
  8: 1.3, 9: 1.3, 10: 1.3,       // 9,10,11페이지
  11: 1.3, 12: 1.3, 13: 1.3, 14: 1.3, 15: 1.3, 16: 1.3, // 12~17페이지
};

(async () => {
  const browser = await puppeteer.launch({
    headless: 'new',
    executablePath: '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
    args: ['--no-sandbox', '--disable-setuid-sandbox']
  });

  const page = await browser.newPage();
  await page.setViewport({ width: BASE_W, height: BASE_H, deviceScaleFactor: DPR });
  await page.goto(`file://${htmlPath}`, { waitUntil: 'networkidle0', timeout: 30000 });

  await page.waitForFunction(() => document.fonts.ready);
  await new Promise(r => setTimeout(r, 1500));

  const total = await page.evaluate(() => document.querySelectorAll('.slide').length);
  console.log(`Found ${total} slides (DPR=${DPR})`);

  const screenshots = [];

  for (let i = 0; i < total; i++) {
    const zoom = ZOOM_MAP[i] || 1.0;

    // 슬라이드 활성화
    await page.evaluate((idx) => {
      const slides = document.querySelectorAll('.slide');
      slides.forEach(s => s.classList.remove('active'));
      slides[idx].classList.add('active');
      const bar = document.querySelector('.progress-bar');
      if (bar) bar.style.width = ((idx + 1) / slides.length * 100) + '%';
      const counter = document.getElementById('counter');
      if (counter) counter.textContent = `${idx + 1} / ${slides.length}`;
    }, i);

    // CSS zoom 적용 (콘텐츠만 확대, 뷰포트는 그대로)
    await page.evaluate((z) => {
      const active = document.querySelector('.slide.active');
      if (active) active.style.transform = `scale(${z})`;
      if (active) active.style.transformOrigin = 'center center';
    }, zoom);

    await new Promise(r => setTimeout(r, 800));

    const screenshot = await page.screenshot({ type: 'png', fullPage: false });
    screenshots.push(screenshot);
    console.log(`Captured slide ${i + 1}/${total} (zoom: ${zoom}x)`);

    // zoom 리셋
    await page.evaluate(() => {
      const active = document.querySelector('.slide.active');
      if (active) active.style.transform = '';
    });
  }

  const pdfPage = await browser.newPage();
  await pdfPage.setViewport({ width: BASE_W, height: BASE_H });

  const imagesHtml = screenshots.map((buf, i) => {
    const base64 = buf.toString('base64');
    const isLast = i === screenshots.length - 1;
    return `<div style="${isLast ? '' : 'page-break-after: always;'} margin: 0; padding: 0; width: ${BASE_W}px; height: ${BASE_H}px;">
      <img src="data:image/png;base64,${base64}" style="width: 100%; height: 100%; object-fit: contain;" />
    </div>`;
  }).join('\n');

  await pdfPage.setContent(`
    <!DOCTYPE html>
    <html>
    <head>
      <style>
        * { margin: 0; padding: 0; }
        body { background: #0a0a0a; }
        @page { size: ${BASE_W}px ${BASE_H}px; margin: 0; }
      </style>
    </head>
    <body>${imagesHtml}</body>
    </html>
  `, { waitUntil: 'networkidle0' });

  await pdfPage.pdf({
    path: outputPath,
    width: `${BASE_W}px`,
    height: `${BASE_H}px`,
    printBackground: true,
    margin: { top: 0, right: 0, bottom: 0, left: 0 }
  });

  console.log(`PDF saved: ${outputPath} (${screenshots.length} pages)`);
  await browser.close();
})();

import puppeteer from 'puppeteer';
import { resolve } from 'path';

const htmlPath = resolve('index.html');
const outputPath = resolve('presentation.pdf');

// 슬라이드별 확대 배율 (1-indexed → 0-indexed)
// 1,2,3: 1.5x | 4: 1.3x | 5,6,7: 1x | 8,9,10: 1.5x | 11: 1x | 12,13,14,15,16: 1.5x | 17: 1x
const SCALE_MAP = {
  0: 1.5, 1: 1.5, 2: 1.5,    // 1,2,3페이지
  3: 1.3,                      // 4페이지
  7: 1.5, 8: 1.5, 9: 1.5,     // 8,9,10페이지
  11: 1.5, 12: 1.5, 13: 1.5, 14: 1.5, 15: 1.5, // 12,13,14,15,16페이지
};

const BASE_W = 1920;
const BASE_H = 1080;

(async () => {
  const browser = await puppeteer.launch({
    headless: 'new',
    executablePath: '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
    args: ['--no-sandbox', '--disable-setuid-sandbox']
  });

  const page = await browser.newPage();

  // 기본 뷰포트 (고해상도)
  await page.setViewport({ width: BASE_W * 2, height: BASE_H * 2, deviceScaleFactor: 2 });
  await page.goto(`file://${htmlPath}`, { waitUntil: 'networkidle0', timeout: 30000 });

  await page.waitForFunction(() => document.fonts.ready);
  await new Promise(r => setTimeout(r, 1500));

  const total = await page.evaluate(() => document.querySelectorAll('.slide').length);
  console.log(`Found ${total} slides`);

  const screenshots = [];

  for (let i = 0; i < total; i++) {
    const scale = SCALE_MAP[i] || 1.0;
    const w = Math.round(BASE_W * scale);
    const h = Math.round(BASE_H * scale);

    // 슬라이드별 뷰포트 + deviceScaleFactor=2 for retina
    await page.setViewport({ width: w, height: h, deviceScaleFactor: 2 });

    await page.evaluate((idx) => {
      const slides = document.querySelectorAll('.slide');
      slides.forEach(s => s.classList.remove('active'));
      slides[idx].classList.add('active');
      const bar = document.querySelector('.progress-bar');
      if (bar) bar.style.width = ((idx + 1) / slides.length * 100) + '%';
      const counter = document.getElementById('counter');
      if (counter) counter.textContent = `${idx + 1} / ${slides.length}`;
    }, i);

    await new Promise(r => setTimeout(r, 1000));

    const screenshot = await page.screenshot({
      type: 'png',
      fullPage: false,
      clip: { x: 0, y: 0, width: w, height: h }
    });
    screenshots.push(screenshot);
    console.log(`Captured slide ${i + 1}/${total} (scale: ${scale}x, ${w}x${h})`);
  }

  // PDF 생성 — 마지막 빈 페이지 방지
  const pdfPage = await browser.newPage();
  await pdfPage.setViewport({ width: BASE_W, height: BASE_H });

  const imagesHtml = screenshots.map((buf, i) => {
    const base64 = buf.toString('base64');
    const isLast = i === screenshots.length - 1;
    return `<div style="${isLast ? '' : 'page-break-after: always;'} margin: 0; padding: 0; width: 1920px; height: 1080px;">
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
        @page { size: 1920px 1080px; margin: 0; }
      </style>
    </head>
    <body>${imagesHtml}</body>
    </html>
  `, { waitUntil: 'networkidle0' });

  await pdfPage.pdf({
    path: outputPath,
    width: '1920px',
    height: '1080px',
    printBackground: true,
    margin: { top: 0, right: 0, bottom: 0, left: 0 }
  });

  console.log(`PDF saved: ${outputPath} (${screenshots.length} pages)`);
  await browser.close();
})();

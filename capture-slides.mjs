import puppeteer from 'puppeteer';
import { resolve } from 'path';
import { writeFileSync } from 'fs';

const htmlPath = resolve('index.html');
const outputPath = resolve('presentation.pdf');

(async () => {
  const browser = await puppeteer.launch({
    headless: 'new',
    executablePath: '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
    args: ['--no-sandbox', '--disable-setuid-sandbox']
  });

  const page = await browser.newPage();
  await page.setViewport({ width: 1920, height: 1080 });
  await page.goto(`file://${htmlPath}`, { waitUntil: 'networkidle0', timeout: 30000 });

  // Wait for fonts to load
  await page.waitForFunction(() => document.fonts.ready);
  await new Promise(r => setTimeout(r, 1000));

  // Get total slides
  const total = await page.evaluate(() => document.querySelectorAll('.slide').length);
  console.log(`Found ${total} slides`);

  const screenshots = [];

  for (let i = 0; i < total; i++) {
    // Navigate to slide
    await page.evaluate((idx) => {
      const slides = document.querySelectorAll('.slide');
      slides.forEach(s => s.classList.remove('active'));
      slides[idx].classList.add('active');
      // Update progress bar
      const bar = document.querySelector('.progress-bar');
      if (bar) bar.style.width = ((idx + 1) / slides.length * 100) + '%';
      // Update counter
      const counter = document.getElementById('counter');
      if (counter) counter.textContent = `${idx + 1} / ${slides.length}`;
    }, i);

    // Wait for animations
    await new Promise(r => setTimeout(r, 800));

    // Take screenshot
    const screenshot = await page.screenshot({ type: 'png', fullPage: false });
    screenshots.push(screenshot);
    console.log(`Captured slide ${i + 1}/${total}`);
  }

  // Generate PDF with all screenshots as pages
  // Use a new page to create PDF from images
  const pdfPage = await browser.newPage();
  await pdfPage.setViewport({ width: 1920, height: 1080 });

  const imagesHtml = screenshots.map((buf, i) => {
    const base64 = buf.toString('base64');
    return `<div style="page-break-after: always; margin: 0; padding: 0; width: 1920px; height: 1080px;">
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

  console.log(`PDF saved: ${outputPath}`);
  await browser.close();
})();

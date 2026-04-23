import puppeteer from 'puppeteer';
import { resolve } from 'path';

(async () => {
  const browser = await puppeteer.launch({
    headless: 'new',
    executablePath: '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
    args: ['--no-sandbox']
  });

  const page = await browser.newPage();
  await page.setViewport({ width: 1200, height: 600, deviceScaleFactor: 3 });

  await page.setContent(`
    <!DOCTYPE html>
    <html>
    <head>
      <link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;600;700&display=swap">
      <style>
        * { margin: 0; padding: 0; }
        body {
          background: transparent;
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          height: 100vh;
          font-family: 'JetBrains Mono', monospace;
        }
        .techai {
          font-size: 28px;
          line-height: 1.1;
          color: #60A5FA;
          font-weight: 700;
          white-space: pre;
          text-align: center;
        }
        .separator {
          width: 100%;
          height: 2px;
          background: linear-gradient(90deg, transparent, #374151, transparent);
          margin: 14px 0;
        }
        .code {
          font-size: 28px;
          line-height: 1.1;
          color: #93C5FD;
          font-weight: 700;
          white-space: pre;
          text-align: center;
        }
      </style>
    </head>
    <body>
      <pre class="techai"> ████████╗███████╗ ██████╗██╗  ██╗ █████╗ ██╗
╚══██╔══╝██╔════╝██╔════╝██║  ██║██╔══██╗██║
   ██║   █████╗  ██║     ███████║███████║██║
   ██║   ██╔══╝  ██║     ██╔══██║██╔══██║██║
   ██║   ███████╗╚██████╗██║  ██║██║  ██║██║
   ╚═╝   ╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝</pre>
      <div class="separator"></div>
      <pre class="code">   ██████╗  ██████╗  ██████╗  ███████╗
  ██╔════╝ ██╔═══██╗ ██╔══██╗ ██╔════╝
  ██║      ██║   ██║ ██║  ██║ █████╗
  ██║      ██║   ██║ ██║  ██║ ██╔══╝
  ╚██████╗ ╚██████╔╝ ██████╔╝ ███████╗
   ╚═════╝  ╚═════╝  ╚═════╝  ╚══════╝</pre>
    </body>
    </html>
  `, { waitUntil: 'networkidle0' });

  await page.waitForFunction(() => document.fonts.ready);
  await new Promise(r => setTimeout(r, 1500));

  await page.screenshot({
    path: resolve('public/assets/techai-logo.png'),
    type: 'png',
    omitBackground: true,
  });

  console.log('Logo captured: public/assets/techai-logo.png');
  await browser.close();
})();

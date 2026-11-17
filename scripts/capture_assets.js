const path = require('path');
const fs = require('fs');

const scratchNodeModules = 'C:/Users/alexander/.gemini/antigravity/brain/045b00da-cd6d-47ef-a759-58ba5b70583f/scratch/node_modules';
const puppeteer = require(path.join(scratchNodeModules, 'puppeteer'));

const outputDir = path.resolve('B:/workgit/agentdiff/docs/images');
const framesDir = path.resolve('B:/workgit/agentdiff/docs/images/frames');

if (!fs.existsSync(outputDir)) fs.mkdirSync(outputDir, { recursive: true });
if (!fs.existsSync(framesDir)) fs.mkdirSync(framesDir, { recursive: true });

async function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

async function capture() {
  console.log('Launching Puppeteer browser...');
  const browser = await puppeteer.launch({
    headless: true,
    defaultViewport: null,
    args: ['--no-sandbox', '--disable-setuid-sandbox', '--font-render-hinting=max']
  });

  const page = await browser.newPage();

  // Helper to click tab by text
  async function clickTab(name) {
    const buttons = await page.$$('button');
    for (const b of buttons) {
      const text = await page.evaluate(el => el.textContent, b);
      if (text && text.includes(name)) {
        await b.click();
        await sleep(400);
        return true;
      }
    }
    return false;
  }

  // 1. DESKTOP DASHBOARD (1440x960, 2x dpr for ultra-crisp display)
  console.log('1. Capturing Desktop Dashboard...');
  await page.setViewport({ width: 1440, height: 960, deviceScaleFactor: 2 });
  await page.goto('http://localhost:7676', { waitUntil: 'domcontentloaded' });
  await page.waitForSelector('header', { timeout: 10000 });
  await sleep(1500);
  await page.screenshot({
    path: path.join(outputDir, 'agentdiff_studio_dashboard.png'),
    fullPage: false
  });
  console.log('✓ Captured agentdiff_studio_dashboard.png');

  // 2. AST SEMANTIC DIFF VIEW
  console.log('2. Capturing AST Semantic Diff View...');
  await clickTab('AST Semantic Diff');
  await sleep(800);
  await page.screenshot({
    path: path.join(outputDir, 'agentdiff_ast_diff.png'),
    fullPage: false
  });
  console.log('✓ Captured agentdiff_ast_diff.png');

  // 3. LEADERBOARD VIEW
  console.log('3. Capturing Leaderboard Matrix View...');
  await clickTab('Leaderboard');
  await sleep(800);
  await page.screenshot({
    path: path.join(outputDir, 'agentdiff_leaderboard.png'),
    fullPage: false
  });
  console.log('✓ Captured agentdiff_leaderboard.png');

  // 4. TABLET VIEW (820x1180, iPad Air)
  console.log('4. Capturing Tablet Responsive View...');
  await page.setViewport({ width: 820, height: 1080, deviceScaleFactor: 2 });
  await clickTab('Synchronized Player');
  await sleep(800);
  await page.screenshot({
    path: path.join(outputDir, 'agentdiff_tablet_view.png'),
    fullPage: false
  });
  console.log('✓ Captured agentdiff_tablet_view.png');

  // 5. MOBILE VIEW (390x844, iPhone 14/15)
  console.log('5. Capturing Mobile Responsive View...');
  await page.setViewport({ width: 390, height: 844, deviceScaleFactor: 2 });
  await sleep(800);
  await page.screenshot({
    path: path.join(outputDir, 'agentdiff_mobile_responsive.png'),
    fullPage: false
  });
  console.log('✓ Captured agentdiff_mobile_responsive.png');

  // 6. RECORD FRAMES FOR DEMO GIF
  console.log('6. Recording frames for agentdiff_demo.gif...');
  await page.setViewport({ width: 1280, height: 800, deviceScaleFactor: 1.5 });
  await page.goto('http://localhost:7676', { waitUntil: 'domcontentloaded' });
  await page.waitForSelector('header', { timeout: 10000 });
  await sleep(1500);

  let frameIdx = 0;
  async function saveFrame(count = 1) {
    for (let c = 0; c < count; c++) {
      const filename = path.join(framesDir, `frame_${String(frameIdx).padStart(4, '0')}.png`);
      await page.screenshot({ path: filename });
      frameIdx++;
    }
  }

  // Initial dashboard view (Turn 1)
  await saveFrame(6);

  // Click Step Forward (chevron right)
  const buttons = await page.$$('button');
  let stepFwdBtn = null;
  for (const b of buttons) {
    const title = await page.evaluate(el => el.getAttribute('title'), b);
    if (title === 'Next Turn') {
      stepFwdBtn = b;
      break;
    }
  }

  if (stepFwdBtn) {
    // Step to Turn 2
    await stepFwdBtn.click();
    await sleep(400);
    await saveFrame(5);

    // Step to Turn 3 (divergence turn!)
    await stepFwdBtn.click();
    await sleep(400);
    await saveFrame(6);

    // Step to Turn 4
    await stepFwdBtn.click();
    await sleep(400);
    await saveFrame(5);
  }

  // Switch to AST Semantic Diff
  await clickTab('AST Semantic Diff');
  await sleep(500);
  await saveFrame(7);

  // Click an AST node or file tab if available
  const nodeChips = await page.$$('span');
  for (const chip of nodeChips) {
    const text = await page.evaluate(el => el.textContent, chip);
    if (text && text.includes('FuncDecl')) {
      await chip.click();
      await sleep(300);
      break;
    }
  }
  await saveFrame(5);

  // Switch to Anti-Patterns tab
  await clickTab('Anti-Patterns');
  await sleep(500);
  await saveFrame(6);

  // Switch to Leaderboard tab
  await clickTab('Leaderboard');
  await sleep(500);
  await saveFrame(7);

  // Switch to Task 2 (task_02_python_data_pipeline)
  const taskCards = await page.$$('button');
  for (const tc of taskCards) {
    const text = await page.evaluate(el => el.textContent, tc);
    if (text && text.includes('task_02_python_data_pipeline')) {
      await tc.click();
      await sleep(600);
      break;
    }
  }
  await clickTab('Synchronized Player');
  await sleep(500);
  await saveFrame(6);

  console.log(`✓ Total recorded ${frameIdx} frames.`);
  await browser.close();
}

capture().catch(err => {
  console.error('Capture failed:', err);
  process.exit(1);
});

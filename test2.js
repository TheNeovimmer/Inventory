const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ 
    executablePath: '/usr/bin/chromium',
    headless: true,
    args: ['--no-sandbox']
  });
  const page = await browser.newPage();
  
  // Login first
  console.log('Logging in...');
  await page.goto('http://localhost:8080/login');
  await page.fill('input[type="email"]', 'admin@inventory.com');
  await page.fill('input[type="password"]', 'admin123');
  await page.click('button[type="submit"]');
  await page.waitForTimeout(2000);
  
  // Test products page
  console.log('\nTesting products page...');
  await page.goto('http://localhost:8080/products');
  await page.waitForTimeout(2000);
  
  // Get full HTML
  const html = await page.content();
  console.log('Products HTML length:', html.length);
  console.log('HTML snippet:', html.substring(0, 2000));
  
  // Check for errors in console
  page.on('console', msg => console.log('Console:', msg.text()));
  
  await browser.close();
})();

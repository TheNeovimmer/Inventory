const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ 
    executablePath: '/usr/bin/chromium',
    headless: true,
    args: ['--no-sandbox']
  });
  const page = await browser.newPage();
  
  // Test login page
  console.log('Testing login page...');
  await page.goto('http://localhost:8080/login');
  const loginTitle = await page.title();
  console.log('Login title:', loginTitle);
  
  // Check for form
  const hasLoginForm = await page.$('form');
  console.log('Has login form:', !!hasLoginForm);
  
  // Try logging in
  console.log('\nLogging in...');
  await page.fill('input[type="email"]', 'admin@inventory.com');
  await page.fill('input[type="password"]', 'admin123');
  await page.click('button[type="submit"]');
  await page.waitForTimeout(2000);
  
  // Check dashboard
  const dashboardUrl = page.url();
  console.log('After login URL:', dashboardUrl);
  
  // Test products page
  console.log('\nTesting products page...');
  await page.goto('http://localhost:8080/products');
  await page.waitForTimeout(1000);
  const productsTitle = await page.title();
  console.log('Products title:', productsTitle);
  
  // Check sidebar
  const sidebar = await page.$('.sidebar');
  console.log('Has sidebar:', !!sidebar);
  
  // Check table
  const table = await page.$('table');
  console.log('Has table:', !!table);
  
  // Get page content for debugging
  const bodyContent = await page.evaluate(() => document.body.innerText.substring(0, 500));
  console.log('Page content preview:', bodyContent);
  
  await browser.close();
  console.log('\nTest completed!');
})();

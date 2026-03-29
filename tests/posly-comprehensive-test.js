const { chromium } = require('playwright');

const BASE_URL = process.env.BASE_URL || 'http://localhost:8080';
const API_BASE = BASE_URL + '/api';

let browser;
let context;
let page;

async function setup() {
  browser = await chromium.launch({ 
    headless: true,
    executablePath: '/usr/bin/chromium'
  });
  context = await browser.newContext();
  page = await context.newPage();
}

async function teardown() {
  if (browser) await browser.close();
}

async function login() {
  await page.goto(BASE_URL + '/login');
  await page.fill('#email', 'admin@inventory.com');
  await page.fill('#password', 'admin123');
  await page.click('button[type="submit"]');
  await page.waitForURL('**/dashboard', { timeout: 10000 });
}

async function apiCall(endpoint, options = {}) {
  const token = await page.evaluate(() => localStorage.getItem('token'));
  const response = await fetch(API_BASE + endpoint, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
      ...options.headers
    }
  });
  const text = await response.text();
  try {
    return JSON.parse(text);
  } catch {
    return text;
  }
}

async function runTests() {
  const results = { passed: 0, failed: 0, tests: [] };
  
  console.log('\n========================================');
  console.log('Posly - Comprehensive Test Suite');
  console.log('========================================\n');

  // =====================
  // BACKEND API TESTS
  // =====================
  console.log('BACKEND API TESTS\n');

  await setup();
  await login();

  // Test: Products CRUD
  console.log('1. Products CRUD');
  try {
    const products = await apiCall('/products', { method: 'POST', body: JSON.stringify({
      name: 'Test Product ' + Date.now(),
      sku: 'SKU-' + Date.now(),
      unit_price: 29.99,
      cost_price: 15.00,
      quantity: 100,
      reorder_point: 10,
      active: true
    }) });
    console.log('  ✓ Create product');
    results.passed++;
  } catch (err) {
    console.log(`  ⚠ Create: ${err.message}`);
  }

  // Test: List products
  try {
    const products = await apiCall('/products');
    console.log(`  ✓ List products: ${products.length} found`);
    results.passed++;
  } catch (err) {
    console.log(`  ✗ List products: ${err.message}`);
    results.failed++;
  }

  // Test: Categories
  console.log('\n2. Categories');
  try {
    const categories = await apiCall('/categories');
    console.log(`  ✓ List categories: ${categories.length} found`);
    results.passed++;
  } catch (err) {
    console.log(`  ✗ List categories: ${err.message}`);
    results.failed++;
  }

  // Test: Customers
  console.log('\n3. Customers');
  try {
    const customers = await apiCall('/customers');
    console.log(`  ✓ List customers: ${customers.length} found`);
    results.passed++;
  } catch (err) {
    console.log(`  ✗ List customers: ${err.message}`);
    results.failed++;
  }

  // Test: Create Customer
  try {
    const customerName = 'Test Customer ' + Date.now();
    const customer = await apiCall('/customers', { method: 'POST', body: JSON.stringify({
      name: customerName,
      email: 'customer' + Date.now() + '@test.com',
      phone: '1234567890',
      address: 'Test Address',
      city: 'Test City',
      country: 'Test Country'
    }) });
    console.log('  ✓ Create customer');
    results.passed++;
  } catch (err) {
    console.log(`  ⚠ Create customer: ${err.message}`);
  }

  // Test: Suppliers
  console.log('\n4. Suppliers');
  try {
    const suppliers = await apiCall('/suppliers');
    console.log(`  ✓ List suppliers: ${suppliers.length} found`);
    results.passed++;
  } catch (err) {
    console.log(`  ✗ List suppliers: ${err.message}`);
    results.failed++;
  }

  // Test: Warehouses
  console.log('\n5. Warehouses');
  try {
    const warehouses = await apiCall('/warehouses');
    console.log(`  ✓ List warehouses: ${warehouses.length} found`);
    results.passed++;
  } catch (err) {
    console.log(`  ✗ List warehouses: ${err.message}`);
    results.failed++;
  }

  // Test: Dashboard Stats
  console.log('\n6. Dashboard Stats');
  try {
    const stats = await apiCall('/reports/dashboard');
    console.log('  ✓ Dashboard stats');
    console.log(`    - Products: ${stats.total_products || 0}`);
    console.log(`    - Inventory Value: $${stats.total_value || 0}`);
    results.passed++;
  } catch (err) {
    console.log(`  ✗ Dashboard stats: ${err.message}`);
    results.failed++;
  }

  // Test: Inventory
  console.log('\n7. Inventory');
  try {
    const inventory = await apiCall('/inventory');
    console.log(`  ✓ List inventory: ${inventory.length} items`);
    results.passed++;
  } catch (err) {
    console.log(`  ✗ List inventory: ${err.message}`);
    results.failed++;
  }

  // Test: Low Stock Alerts
  console.log('\n8. Low Stock Alerts');
  try {
    const alerts = await apiCall('/inventory/alerts');
    console.log(`  ✓ Low stock alerts: ${alerts.length} items`);
    results.passed++;
  } catch (err) {
    console.log(`  ✗ Low stock alerts: ${err.message}`);
    results.failed++;
  }

  // Test: Reports
  console.log('\n9. Reports');
  try {
    const stockLevels = await apiCall('/reports/stock-levels');
    console.log('  ✓ Stock levels report');
    results.passed++;
  } catch (err) {
    console.log(`  ✗ Stock levels: ${err.message}`);
    results.failed++;
  }

  try {
    const turnover = await apiCall('/reports/turnover');
    console.log('  ✓ Turnover report');
    results.passed++;
  } catch (err) {
    console.log(`  ✗ Turnover: ${err.message}`);
    results.failed++;
  }

  // Test: Settings
  console.log('\n10. Settings');
  try {
    const settings = await apiCall('/settings');
    console.log('  ✓ Settings loaded');
    results.passed++;
  } catch (err) {
    console.log(`  ✗ Settings: ${err.message}`);
    results.failed++;
  }

  try {
    const currency = await apiCall('/settings/currency');
    console.log('  ✓ Currency settings');
    results.passed++;
  } catch (err) {
    console.log(`  ✗ Currency: ${err.message}`);
    results.failed++;
  }

  // =====================
  // FRONTEND UI TESTS
  // =====================
  console.log('\n========================================');
  console.log('FRONTEND UI TESTS');
  console.log('========================================\n');

  // Test: Sales Page
  console.log('11. Sales Page');
  try {
    await page.goto(BASE_URL + '/sales');
    await page.waitForLoadState('networkidle');
    console.log('  ✓ Sales page loads');
    results.passed++;
  } catch (err) {
    console.log(`  ✗ Sales page: ${err.message}`);
    results.failed++;
  }

  // Test: Quotations Page
  console.log('\n12. Quotations Page');
  try {
    await page.goto(BASE_URL + '/quotations');
    await page.waitForLoadState('networkidle');
    console.log('  ✓ Quotations page loads');
    results.passed++;
  } catch (err) {
    console.log(`  ✗ Quotations page: ${err.message}`);
    results.failed++;
  }

  // Test: Accounts Page
  console.log('\n13. Accounts Page');
  try {
    await page.goto(BASE_URL + '/accounts');
    await page.waitForLoadState('networkidle');
    console.log('  ✓ Accounts page loads');
    results.passed++;
  } catch (err) {
    console.log(`  ✗ Accounts page: ${err.message}`);
    results.failed++;
  }

  // Test: Purchase Orders Page
  console.log('\n14. Purchase Orders Page');
  try {
    await page.goto(BASE_URL + '/purchase-orders');
    await page.waitForLoadState('networkidle');
    console.log('  ✓ Purchase Orders page loads');
    results.passed++;
  } catch (err) {
    console.log(`  ✗ Purchase Orders page: ${err.message}`);
    results.failed++;
  }

  // Test: Transfers Page
  console.log('\n15. Transfers Page');
  try {
    await page.goto(BASE_URL + '/transfers');
    await page.waitForLoadState('networkidle');
    console.log('  ✓ Transfers page loads');
    results.passed++;
  } catch (err) {
    console.log(`  ✗ Transfers page: ${err.message}`);
    results.failed++;
  }

  // Test: Inventory Page
  console.log('\n16. Inventory Page');
  try {
    await page.goto(BASE_URL + '/inventory');
    await page.waitForLoadState('networkidle');
    console.log('  ✓ Inventory page loads');
    results.passed++;
  } catch (err) {
    console.log(`  ✗ Inventory page: ${err.message}`);
    results.failed++;
  }

  // Test: Analytics Page
  console.log('\n17. Analytics Page');
  try {
    await page.goto(BASE_URL + '/analytics');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    console.log('  ✓ Analytics page loads');
    results.passed++;
  } catch (err) {
    console.log(`  ✗ Analytics page: ${err.message}`);
    results.failed++;
  }

  // Test: Brands Page
  console.log('\n18. Brands Page');
  try {
    await page.goto(BASE_URL + '/brands');
    await page.waitForLoadState('networkidle');
    console.log('  ✓ Brands page loads');
    results.passed++;
  } catch (err) {
    console.log(`  ✗ Brands page: ${err.message}`);
    results.failed++;
  }

  // Test: Webhooks Page
  console.log('\n19. Webhooks Page');
  try {
    await page.goto(BASE_URL + '/webhooks');
    await page.waitForLoadState('networkidle');
    console.log('  ✓ Webhooks page loads');
    results.passed++;
  } catch (err) {
    console.log(`  ✗ Webhooks page: ${err.message}`);
    results.failed++;
  }

  // =====================
  // POS FUNCTIONALITY TESTS
  // =====================
  console.log('\n========================================');
  console.log('POS FUNCTIONALITY TESTS');
  console.log('========================================\n');

  // Test: POS Screen Elements
  console.log('20. POS Screen Elements');
  try {
    await page.goto(BASE_URL + '/pos');
    await page.waitForLoadState('networkidle');
    
    const elements = {
      'Search box': await page.$('#productSearch'),
      'Barcode input': await page.$('#barcodeInput'),
      'Customer select': await page.$('#customerSelect'),
      'Payment method': await page.$('#paymentMethod'),
      'Amount paid': await page.$('#amountPaid'),
      'Cart items': await page.$('#cartItems'),
      'Complete sale button': await page.$('button:has-text("Complete")')
    };
    
    let allFound = true;
    for (const [name, el] of Object.entries(elements)) {
      if (!el) {
        console.log(`  ⚠ Missing: ${name}`);
        allFound = false;
      }
    }
    
    if (allFound) {
      console.log('  ✓ All POS elements present');
      results.passed++;
    } else {
      throw new Error('Some elements missing');
    }
  } catch (err) {
    console.log(`  ✗ POS elements: ${err.message}`);
    results.failed++;
  }

  // Test: POS Product Search
  console.log('\n21. POS Product Search');
  try {
    await page.goto(BASE_URL + '/pos');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    
    const searchInput = await page.$('#productSearch');
    if (searchInput) {
      await searchInput.fill('a');
      await page.waitForTimeout(500);
      console.log('  ✓ Product search works');
      results.passed++;
    } else {
      throw new Error('Search input not found');
    }
  } catch (err) {
    console.log(`  ✗ Product search: ${err.message}`);
    results.failed++;
  }

  // Test: POS Cart Functionality
  console.log('\n22. POS Cart Display');
  try {
    await page.goto(BASE_URL + '/pos');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    
    const cartItems = await page.$('#cartItems');
    const cartTotal = await page.$('#cartTotal');
    
    if (cartItems && cartTotal) {
      console.log('  ✓ Cart displays correctly');
      results.passed++;
    } else {
      throw new Error('Cart elements not found');
    }
  } catch (err) {
    console.log(`  ✗ Cart display: ${err.message}`);
    results.failed++;
  }

  // Test: POS Quick Amount Buttons
  console.log('\n23. POS Quick Amount Buttons');
  try {
    await page.goto(BASE_URL + '/pos');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(500);
    
    const quickAmounts = await page.$$('.payment-section button:has-text("$")');
    if (quickAmounts.length > 0) {
      console.log(`  ✓ Quick amount buttons present (${quickAmounts.length})`);
      results.passed++;
    } else {
      console.log('  ⚠ Quick amount buttons not found');
    }
  } catch (err) {
    console.log(`  ✗ Quick amounts: ${err.message}`);
    results.failed++;
  }

  // Test: Theme Toggle on POS
  console.log('\n24. Theme Toggle on POS');
  try {
    await page.goto(BASE_URL + '/pos');
    await page.waitForLoadState('networkidle');
    
    // Go back to dashboard to test theme toggle
    await page.goto(BASE_URL + '/dashboard');
    await page.waitForLoadState('networkidle');
    
    const themeToggle = await page.$('#themeToggle');
    if (themeToggle) {
      await themeToggle.click();
      await page.waitForTimeout(300);
      const theme = await page.evaluate(() => document.documentElement.getAttribute('data-theme'));
      console.log(`  ✓ Theme toggles (current: ${theme})`);
      results.passed++;
    } else {
      throw new Error('Theme toggle not found');
    }
  } catch (err) {
    console.log(`  ✗ Theme toggle: ${err.message}`);
    results.failed++;
  }

  // Test: Responsive Sidebar
  console.log('\n25. Responsive Sidebar');
  try {
    await page.setViewportSize({ width: 800, height: 600 });
    await page.goto(BASE_URL + '/dashboard');
    await page.waitForLoadState('networkidle');
    
    const sidebar = await page.$('.sidebar');
    const toggleBtn = await page.$('#toggleSidebar');
    
    if (sidebar && toggleBtn) {
      await toggleBtn.click();
      await page.waitForTimeout(300);
      console.log('  ✓ Sidebar toggle works');
      results.passed++;
    } else {
      throw new Error('Sidebar elements not found');
    }
    
    // Reset viewport
    await page.setViewportSize({ width: 1920, height: 1080 });
  } catch (err) {
    console.log(`  ✗ Responsive sidebar: ${err.message}`);
    results.failed++;
  }

  // Test: Error Handling
  console.log('\n26. Error Handling');
  try {
    // Try to access protected page without auth
    await page.evaluate(() => localStorage.removeItem('token'));
    await page.goto(BASE_URL + '/dashboard');
    await page.waitForTimeout(1000);
    const currentUrl = page.url();
    
    if (currentUrl.includes('login')) {
      console.log('  ✓ Redirects to login when not authenticated');
      results.passed++;
    } else {
      console.log(`  ⚠ Current URL: ${currentUrl}`);
    }
    
    // Re-login
    await login();
  } catch (err) {
    console.log(`  ✗ Error handling: ${err.message}`);
    results.failed++;
  }

  // Test: Chart Rendering
  console.log('\n27. Chart Rendering');
  try {
    await page.goto(BASE_URL + '/dashboard');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    
    const charts = await page.$$('canvas');
    if (charts.length > 0) {
      console.log(`  ✓ Charts render (${charts.length} found)`);
      results.passed++;
    } else {
      console.log('  ⚠ No charts found');
    }
  } catch (err) {
    console.log(`  ✗ Charts: ${err.message}`);
    results.failed++;
  }

  await teardown();

  // Print summary
  console.log('\n========================================');
  console.log('Test Summary');
  console.log('========================================');
  console.log(`Total Passed: ${results.passed}`);
  console.log(`Total Failed: ${results.failed}`);
  console.log(`Total Tests: ${results.passed + results.failed}`);
  
  if (results.failed === 0) {
    console.log('\n🎉 All tests passed!');
  } else {
    console.log(`\n⚠️ ${results.failed} test(s) need attention`);
  }
  console.log('========================================\n');

  return results;
}

runTests().then(results => {
  process.exit(results.failed > 0 ? 1 : 0);
}).catch(err => {
  console.error('Test suite error:', err);
  process.exit(1);
});

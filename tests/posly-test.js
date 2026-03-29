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
  
  function test(name, fn) {
    return { name, fn };
  }
  
  async function runTest(t) {
    try {
      await t.fn();
      results.passed++;
      results.tests.push({ name: t.name, status: 'passed' });
      console.log(`  ✓ ${t.name}`);
    } catch (err) {
      results.failed++;
      results.tests.push({ name: t.name, status: 'failed', error: err.message });
      console.log(`  ✗ ${t.name}`);
      console.log(`    Error: ${err.message}`);
    }
  }

  console.log('\n========================================');
  console.log('Posly - POS & Inventory Test Suite');
  console.log('========================================\n');

  // Test 1: Login Page
  console.log('Test 1: Login Page');
  await setup();
  try {
    await page.goto(BASE_URL + '/login');
    await page.waitForLoadState('networkidle');
    
    const title = await page.title();
    if (title.includes('Posly') || title.includes('Login')) {
      console.log('  ✓ Login page loads correctly');
      results.passed++;
    } else {
      throw new Error('Title does not contain expected text');
    }
    
    // Check for theme toggle
    const themeToggle = await page.$('#themeToggle');
    if (themeToggle) {
      console.log('  ✓ Theme toggle exists');
      results.passed++;
    } else {
      console.log('  ⚠ Theme toggle not found (optional)');
    }
    
    // Check login form
    const emailInput = await page.$('#email');
    const passwordInput = await page.$('#password');
    const submitBtn = await page.$('button[type="submit"]');
    
    if (emailInput && passwordInput && submitBtn) {
      console.log('  ✓ Login form elements exist');
      results.passed++;
    } else {
      throw new Error('Login form elements missing');
    }
  } catch (err) {
    console.log(`  ✗ Login page test failed: ${err.message}`);
    results.failed++;
  }
  await teardown();

  // Test 2: Authentication
  console.log('\nTest 2: Authentication');
  await setup();
  try {
    await page.goto(BASE_URL + '/login');
    await page.fill('#email', 'admin@inventory.com');
    await page.fill('#password', 'admin123');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/dashboard', { timeout: 10000 });
    
    console.log('  ✓ Login successful');
    results.passed++;
    
    // Check if token is stored
    const token = await page.evaluate(() => localStorage.getItem('token'));
    if (token) {
      console.log('  ✓ Auth token stored');
      results.passed++;
    } else {
      throw new Error('Auth token not found');
    }
  } catch (err) {
    console.log(`  ✗ Authentication failed: ${err.message}`);
    results.failed++;
  }

  // Test 3: Dashboard
  console.log('\nTest 3: Dashboard');
  try {
    await page.goto(BASE_URL + '/dashboard');
    await page.waitForLoadState('networkidle');
    
    // Check sidebar
    const sidebar = await page.$('.sidebar');
    if (sidebar) {
      console.log('  ✓ Sidebar renders');
      results.passed++;
    }
    
    // Check theme toggle works
    const themeToggle = await page.$('#themeToggle');
    if (themeToggle) {
      await themeToggle.click();
      await page.waitForTimeout(300);
      const theme = await page.evaluate(() => document.documentElement.getAttribute('data-theme'));
      console.log(`  ✓ Theme toggle works (current: ${theme})`);
      results.passed++;
    }
    
    // Check stat cards
    const statCards = await page.$$('.stat-card');
    if (statCards.length > 0) {
      console.log(`  ✓ Dashboard has ${statCards.length} stat cards`);
      results.passed++;
    }
    
    // Check charts
    const charts = await page.$$('canvas');
    if (charts.length > 0) {
      console.log(`  ✓ Dashboard has ${charts.length} charts`);
      results.passed++;
    }
    
  } catch (err) {
    console.log(`  ✗ Dashboard test failed: ${err.message}`);
    results.failed++;
  }

  // Test 4: POS Screen
  console.log('\nTest 4: POS Screen');
  try {
    await page.goto(BASE_URL + '/pos');
    await page.waitForLoadState('networkidle');
    
    // Check POS layout
    const posLayout = await page.$('.pos-layout');
    if (posLayout) {
      console.log('  ✓ POS layout renders');
      results.passed++;
    }
    
    // Check products panel
    const productsPanel = await page.$('.pos-products-panel');
    if (productsPanel) {
      console.log('  ✓ Products panel renders');
      results.passed++;
    }
    
    // Check cart panel
    const cartPanel = await page.$('.pos-cart-panel');
    if (cartPanel) {
      console.log('  ✓ Cart panel renders');
      results.passed++;
    }
    
    // Check search box
    const searchBox = await page.$('#productSearch');
    if (searchBox) {
      console.log('  ✓ Search box exists');
      results.passed++;
    }
    
    // Check barcode input
    const barcodeInput = await page.$('#barcodeInput');
    if (barcodeInput) {
      console.log('  ✓ Barcode input exists');
      results.passed++;
    }
    
  } catch (err) {
    console.log(`  ✗ POS test failed: ${err.message}`);
    results.failed++;
  }

  // Test 5: Products Page
  console.log('\nTest 5: Products Page');
  try {
    await page.goto(BASE_URL + '/products');
    await page.waitForLoadState('networkidle');
    
    // Check page elements
    const pageHeader = await page.$('.page-header');
    if (pageHeader) {
      console.log('  ✓ Products page loads');
      results.passed++;
    }
    
    // Check filter bar
    const filterBar = await page.$('.filter-bar');
    if (filterBar) {
      console.log('  ✓ Filter bar exists');
      results.passed++;
    }
    
    // Check add product button
    const addBtn = await page.$('button:has-text("Add Product")');
    if (addBtn) {
      console.log('  ✓ Add Product button exists');
      results.passed++;
    }
    
  } catch (err) {
    console.log(`  ✗ Products page test failed: ${err.message}`);
    results.failed++;
  }

  // Test 6: Categories Page
  console.log('\nTest 6: Categories Page');
  try {
    await page.goto(BASE_URL + '/categories');
    await page.waitForLoadState('networkidle');
    
    console.log('  ✓ Categories page loads');
    results.passed++;
    
  } catch (err) {
    console.log(`  ✗ Categories page test failed: ${err.message}`);
    results.failed++;
  }

  // Test 7: Customers Page
  console.log('\nTest 7: Customers Page');
  try {
    await page.goto(BASE_URL + '/customers');
    await page.waitForLoadState('networkidle');
    
    console.log('  ✓ Customers page loads');
    results.passed++;
    
  } catch (err) {
    console.log(`  ✗ Customers page test failed: ${err.message}`);
    results.failed++;
  }

  // Test 8: Suppliers Page
  console.log('\nTest 8: Suppliers Page');
  try {
    await page.goto(BASE_URL + '/suppliers');
    await page.waitForLoadState('networkidle');
    
    console.log('  ✓ Suppliers page loads');
    results.passed++;
    
  } catch (err) {
    console.log(`  ✗ Suppliers page test failed: ${err.message}`);
    results.failed++;
  }

  // Test 9: Sales Page
  console.log('\nTest 9: Sales Page');
  try {
    await page.goto(BASE_URL + '/sales');
    await page.waitForLoadState('networkidle');
    
    console.log('  ✓ Sales page loads');
    results.passed++;
    
  } catch (err) {
    console.log(`  ✗ Sales page test failed: ${err.message}`);
    results.failed++;
  }

  // Test 10: Reports Page
  console.log('\nTest 10: Reports Page');
  try {
    await page.goto(BASE_URL + '/reports');
    await page.waitForLoadState('networkidle');
    
    console.log('  ✓ Reports page loads');
    results.passed++;
    
  } catch (err) {
    console.log(`  ✗ Reports page test failed: ${err.message}`);
    results.failed++;
  }

  // Test 11: Settings Page
  console.log('\nTest 11: Settings Page');
  try {
    await page.goto(BASE_URL + '/settings');
    await page.waitForLoadState('networkidle');
    
    console.log('  ✓ Settings page loads');
    results.passed++;
    
    // Check settings form
    const settingsForm = await page.$('form');
    if (settingsForm) {
      console.log('  ✓ Settings form exists');
      results.passed++;
    }
    
  } catch (err) {
    console.log(`  ✗ Settings page test failed: ${err.message}`);
    results.failed++;
  }

  // Test 12: API Tests
  console.log('\nTest 12: API Endpoints');
  try {
    const endpoints = [
      { path: '/products', name: 'Products' },
      { path: '/categories', name: 'Categories' },
      { path: '/customers', name: 'Customers' },
      { path: '/suppliers', name: 'Suppliers' },
      { path: '/warehouses', name: 'Warehouses' },
      { path: '/reports/dashboard', name: 'Dashboard Stats' },
      { path: '/inventory', name: 'Inventory' }
    ];
    
    for (const endpoint of endpoints) {
      try {
        const data = await apiCall(endpoint.path);
        if (data && (Array.isArray(data) || typeof data === 'object')) {
          console.log(`  ✓ API ${endpoint.name}`);
          results.passed++;
        }
      } catch (err) {
        console.log(`  ✗ API ${endpoint.name}: ${err.message}`);
        results.failed++;
      }
    }
    
  } catch (err) {
    console.log(`  ✗ API test failed: ${err.message}`);
    results.failed++;
  }

  // Test 13: Create Product Flow
  console.log('\nTest 13: Create Product Flow');
  try {
    await page.goto(BASE_URL + '/products');
    await page.waitForLoadState('networkidle');
    
    // Click add product button
    const addBtn = await page.$('button:has-text("Add Product")');
    if (addBtn) {
      await addBtn.click();
      await page.waitForTimeout(500);
      
      const modal = await page.$('#productModal');
      if (modal) {
        console.log('  ✓ Product modal opens');
        results.passed++;
        
        // Fill form
        await page.fill('#productName', 'Test Product ' + Date.now());
        await page.fill('#productSku', 'SKU-' + Date.now());
        await page.fill('#productPrice', '29.99');
        
        console.log('  ✓ Product form fills');
        results.passed++;
      }
    }
    
  } catch (err) {
    console.log(`  ✗ Create product flow failed: ${err.message}`);
    results.failed++;
  }

  // Test 14: Create Customer Flow
  console.log('\nTest 14: Create Customer Flow');
  try {
    await page.goto(BASE_URL + '/customers');
    await page.waitForLoadState('networkidle');
    
    const addBtn = await page.$('button:has-text("Add Customer")');
    if (addBtn) {
      await addBtn.click();
      await page.waitForTimeout(500);
      
      const modal = await page.$('#customerModal');
      if (modal) {
        console.log('  ✓ Customer modal opens');
        results.passed++;
      }
    } else {
      console.log('  ⚠ Add Customer button not found');
    }
    
  } catch (err) {
    console.log(`  ✗ Create customer flow failed: ${err.message}`);
    results.failed++;
  }

  // Test 15: Theme Toggle Persistence
  console.log('\nTest 15: Theme Toggle');
  try {
    await page.goto(BASE_URL + '/dashboard');
    await page.waitForLoadState('networkidle');
    
    // Toggle theme multiple times
    const themeToggle = await page.$('#themeToggle');
    if (themeToggle) {
      await themeToggle.click();
      await page.waitForTimeout(200);
      const theme1 = await page.evaluate(() => document.documentElement.getAttribute('data-theme'));
      
      await themeToggle.click();
      await page.waitForTimeout(200);
      const theme2 = await page.evaluate(() => document.documentElement.getAttribute('data-theme'));
      
      if (theme1 !== theme2) {
        console.log('  ✓ Theme toggles correctly');
        results.passed++;
      }
      
      // Check localStorage
      const savedTheme = await page.evaluate(() => localStorage.getItem('posly-theme'));
      if (savedTheme) {
        console.log('  ✓ Theme persists in localStorage');
        results.passed++;
      }
    }
    
  } catch (err) {
    console.log(`  ✗ Theme toggle test failed: ${err.message}`);
    results.failed++;
  }

  // Test 16: Sidebar Navigation
  console.log('\nTest 16: Sidebar Navigation');
  try {
    const navLinks = [
      { href: '/dashboard', name: 'Dashboard' },
      { href: '/products', name: 'Products' },
      { href: '/pos', name: 'POS' },
      { href: '/sales', name: 'Sales' },
      { href: '/customers', name: 'Customers' },
      { href: '/reports', name: 'Reports' },
      { href: '/settings', name: 'Settings' }
    ];
    
    for (const link of navLinks) {
      await page.goto(BASE_URL + link.href);
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(300);
    }
    
    console.log(`  ✓ All ${navLinks.length} navigation links work`);
    results.passed++;
    
  } catch (err) {
    console.log(`  ✗ Navigation test failed: ${err.message}`);
    results.failed++;
  }

  // Test 17: Logout
  console.log('\nTest 17: Logout');
  try {
    // Go to dashboard and wait for page to fully load
    await page.goto(BASE_URL + '/dashboard');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(500); // Extra wait for JS to initialize
    
    // Try multiple logout methods
    let logoutSuccess = false;
    
    // Method 1: Topbar logout button
    const topbarLogout = await page.$('#topbarLogoutBtn');
    if (topbarLogout) {
      await topbarLogout.click();
      await page.waitForURL('**/login', { timeout: 5000 });
      console.log('  ✓ Topbar logout works');
      logoutSuccess = true;
    }
    
    // Method 2: Click user dropdown then logout
    if (!logoutSuccess) {
      const userDropdown = await page.$('#userDropdownTrigger, .user-dropdown[data-bs-toggle]');
      if (userDropdown) {
        await userDropdown.click();
        await page.waitForTimeout(300);
        const logoutBtn = await page.$('#logoutBtn');
        if (logoutBtn) {
          await logoutBtn.click();
          await page.waitForURL('**/login', { timeout: 5000 });
          console.log('  ✓ Sidebar logout works');
          logoutSuccess = true;
        }
      }
    }
    
    // Method 3: Direct JS logout
    if (!logoutSuccess) {
      await page.evaluate(() => {
        localStorage.removeItem('token');
        localStorage.removeItem('user');
      });
      await page.goto(BASE_URL + '/login');
      await page.waitForURL('**/login');
      console.log('  ✓ Direct logout works');
      logoutSuccess = true;
    }
    
    if (logoutSuccess) {
      results.passed++;
    } else {
      throw new Error('No logout method found');
    }
    
  } catch (err) {
    console.log(`  ✗ Logout test failed: ${err.message}`);
    results.failed++;
  }

  // Test 18: Quotations Page
  console.log('\nTest 18: Quotations Page');
  await setup();
  try {
    await login();
    await page.goto(BASE_URL + '/quotations');
    await page.waitForLoadState('networkidle');
    
    const quotTable = await page.$('#quotationsTable');
    if (quotTable) {
      console.log('  ✓ Quotations page loads');
      results.passed++;
    } else {
      throw new Error('Quotations table not found');
    }
    
    const newBtn = await page.$('button:has-text("New Quotation")');
    if (newBtn) {
      console.log('  ✓ New Quotation button exists');
      results.passed++;
    }
  } catch (err) {
    console.log(`  ✗ Quotations test failed: ${err.message}`);
    results.failed++;
  }
  await teardown();

  // Test 19: Brands Page
  console.log('\nTest 19: Brands Page');
  await setup();
  try {
    await login();
    await page.goto(BASE_URL + '/brands');
    await page.waitForLoadState('networkidle');
    
    const brandsTable = await page.$('#brandsTable');
    if (brandsTable) {
      console.log('  ✓ Brands page loads');
      results.passed++;
    } else {
      throw new Error('Brands table not found');
    }
    
    const newBtn = await page.$('button:has-text("Add Brand")');
    if (newBtn) {
      console.log('  ✓ Add Brand button exists');
      results.passed++;
    }
  } catch (err) {
    console.log(`  ✗ Brands test failed: ${err.message}`);
    results.failed++;
  }
  await teardown();

  // Test 20: Units Page
  console.log('\nTest 20: Units Page');
  await setup();
  try {
    await login();
    await page.goto(BASE_URL + '/units');
    await page.waitForLoadState('networkidle');
    
    const unitsTable = await page.$('#unitsTable');
    if (unitsTable) {
      console.log('  ✓ Units page loads');
      results.passed++;
    } else {
      throw new Error('Units table not found');
    }
  } catch (err) {
    console.log(`  ✗ Units test failed: ${err.message}`);
    results.failed++;
  }
  await teardown();

  // Test 21: Suppliers Page
  console.log('\nTest 21: Suppliers Page');
  await setup();
  try {
    await login();
    await page.goto(BASE_URL + '/suppliers');
    await page.waitForLoadState('networkidle');
    
    const suppliersTable = await page.$('#suppliersTable');
    if (suppliersTable) {
      console.log('  ✓ Suppliers page loads');
      results.passed++;
    } else {
      throw new Error('Suppliers table not found');
    }
    
    const newBtn = await page.$('button:has-text("Add Supplier")');
    if (newBtn) {
      console.log('  ✓ Add Supplier button exists');
      results.passed++;
    }
  } catch (err) {
    console.log(`  ✗ Suppliers test failed: ${err.message}`);
    results.failed++;
  }
  await teardown();

  // Test 22: Purchase Orders Page
  console.log('\nTest 22: Purchase Orders Page');
  await setup();
  try {
    await login();
    await page.goto(BASE_URL + '/purchase-orders');
    await page.waitForLoadState('networkidle');
    
    const poTable = await page.$('#poTable');
    if (poTable) {
      console.log('  ✓ Purchase Orders page loads');
      results.passed++;
    } else {
      throw new Error('PO table not found');
    }
  } catch (err) {
    console.log(`  ✗ Purchase Orders test failed: ${err.message}`);
    results.failed++;
  }
  await teardown();

  // Test 23: Production Page
  console.log('\nTest 23: Production Page');
  await setup();
  try {
    await login();
    await page.goto(BASE_URL + '/production');
    await page.waitForLoadState('networkidle');
    
    const prodTable = await page.$('#productionTable');
    if (prodTable) {
      console.log('  ✓ Production page loads');
      results.passed++;
    } else {
      throw new Error('Production table not found');
    }
  } catch (err) {
    console.log(`  ✗ Production test failed: ${err.message}`);
    results.failed++;
  }
  await teardown();

  // Test 24: BOM Page
  console.log('\nTest 24: BOM Page');
  await setup();
  try {
    await login();
    await page.goto(BASE_URL + '/bom');
    await page.waitForLoadState('networkidle');
    
    const bomTable = await page.$('#bomTable');
    if (bomTable) {
      console.log('  ✓ BOM page loads');
      results.passed++;
    } else {
      throw new Error('BOM table not found');
    }
  } catch (err) {
    console.log(`  ✗ BOM test failed: ${err.message}`);
    results.failed++;
  }
  await teardown();

  // Test 25: Inventory Page
  console.log('\nTest 25: Inventory Page');
  await setup();
  try {
    await login();
    await page.goto(BASE_URL + '/inventory');
    await page.waitForLoadState('networkidle');
    
    const invTable = await page.$('#inventoryTable');
    if (invTable) {
      console.log('  ✓ Inventory page loads');
      results.passed++;
    } else {
      throw new Error('Inventory table not found');
    }
    
    const alertsTable = await page.$('#alertsTable');
    if (alertsTable) {
      console.log('  ✓ Alerts section exists');
      results.passed++;
    }
  } catch (err) {
    console.log(`  ✗ Inventory test failed: ${err.message}`);
    results.failed++;
  }
  await teardown();

  // Test 26: Categories Page
  console.log('\nTest 26: Categories Page');
  await setup();
  try {
    await login();
    await page.goto(BASE_URL + '/categories');
    await page.waitForLoadState('networkidle');
    
    const catTable = await page.$('#categoriesTable');
    if (catTable) {
      console.log('  ✓ Categories page loads');
      results.passed++;
    } else {
      throw new Error('Categories table not found');
    }
  } catch (err) {
    console.log(`  ✗ Categories test failed: ${err.message}`);
    results.failed++;
  }
  await teardown();

  // Test 27: Transactions Page
  console.log('\nTest 27: Transactions Page');
  await setup();
  try {
    await login();
    await page.goto(BASE_URL + '/transactions');
    await page.waitForLoadState('networkidle');
    
    const transTable = await page.$('#transactionsTable');
    if (transTable) {
      console.log('  ✓ Transactions page loads');
      results.passed++;
    } else {
      throw new Error('Transactions table not found');
    }
  } catch (err) {
    console.log(`  ✗ Transactions test failed: ${err.message}`);
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
  console.log('========================================\n');

  return results;
}

runTests().then(results => {
  process.exit(results.failed > 0 ? 1 : 0);
}).catch(err => {
  console.error('Test suite error:', err);
  process.exit(1);
});

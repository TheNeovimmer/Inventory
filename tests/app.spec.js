const { chromium } = require('playwright');

const BASE_URL = process.env.BASE_URL || 'http://localhost:8080';
const API_BASE = BASE_URL + '/api';

let browser;
let context;
let page;

async function setup() {
  browser = await chromium.launch({ headless: true });
  context = await browser.newContext();
  page = await context.newPage();
}

async function teardown() {
  if (browser) await browser.close();
}

async function login() {
  await page.goto(BASE_URL + '/login');
  await page.fill('input[name="email"]', 'admin@inventory.com');
  await page.fill('input[name="password"]', 'admin123');
  await page.click('button[type="submit"]');
  await page.waitForURL('**/dashboard');
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
  return response.json();
}

describe('Authentication', () => {
  test('should show login page', async () => {
    await page.goto(BASE_URL + '/login');
    const title = await page.title();
    expect(title).toContain('Login');
  });

  test('should login with valid credentials', async () => {
    await page.goto(BASE_URL + '/login');
    await page.fill('input[name="email"]', 'admin@inventory.com');
    await page.fill('input[name="password"]', 'admin123');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/dashboard', { timeout: 5000 });
  });

  test('should show error with invalid credentials', async () => {
    await page.goto(BASE_URL + '/login');
    await page.fill('input[name="email"]', 'invalid@test.com');
    await page.fill('input[name="password"]', 'wrongpassword');
    await page.click('button[type="submit"]');
    await page.waitForSelector('.alert-danger', { timeout: 3000 });
  });
});

describe('Dashboard', () => {
  test.beforeEach(async () => {
    await setup();
    await login();
  });

  test('should load dashboard', async () => {
    await page.goto(BASE_URL + '/dashboard');
    await page.waitForSelector('#totalProducts', { timeout: 5000 });
  });

  test('should display KPI cards', async () => {
    await page.goto(BASE_URL + '/dashboard');
    const totalProducts = await page.textContent('#totalProducts');
    expect(totalProducts).toBeDefined();
  });

  test('should display charts', async () => {
    await page.goto(BASE_URL + '/dashboard');
    await page.waitForSelector('#categoryChart', { timeout: 5000 });
  });

  test('should toggle theme', async () => {
    await page.goto(BASE_URL + '/dashboard');
    const themeToggle = await page.$('#themeToggle');
    if (themeToggle) {
      await themeToggle.click();
      const theme = await page.evaluate(() => document.documentElement.getAttribute('data-theme'));
      expect(['light', 'dark']).toContain(theme);
    }
  });
});

describe('Products', () => {
  test.beforeEach(async () => {
    await setup();
    await login();
  });

  test('should list products', async () => {
    await page.goto(BASE_URL + '/products');
    await page.waitForSelector('table', { timeout: 5000 });
  });

  test('should open create product modal', async () => {
    await page.goto(BASE_URL + '/products');
    await page.click('button:has-text("Add Product")');
    await page.waitForSelector('#productModal', { timeout: 3000 });
  });

  test('should search products', async () => {
    await page.goto(BASE_URL + '/products');
    await page.fill('#searchProducts', 'test');
    await page.waitForTimeout(500);
  });
});

describe('Inventory', () => {
  test.beforeEach(async () => {
    await setup();
    await login();
  });

  test('should list inventory', async () => {
    await page.goto(BASE_URL + '/inventory');
    await page.waitForSelector('table', { timeout: 5000 });
  });

  test('should show low stock alerts', async () => {
    await page.goto(BASE_URL + '/inventory');
    await page.click('#lowStockFilter');
    await page.waitForTimeout(500);
  });
});

describe('Categories', () => {
  test.beforeEach(async () => {
    await setup();
    await login();
  });

  test('should list categories', async () => {
    await page.goto(BASE_URL + '/categories');
    await page.waitForSelector('table', { timeout: 5000 });
  });

  test('should create category', async () => {
    await page.goto(BASE_URL + '/categories');
    await page.click('button:has-text("Add Category")');
    await page.waitForSelector('#categoryModal', { timeout: 3000 });
  });
});

describe('Suppliers', () => {
  test.beforeEach(async () => {
    await setup();
    await login();
  });

  test('should list suppliers', async () => {
    await page.goto(BASE_URL + '/suppliers');
    await page.waitForSelector('table', { timeout: 5000 });
  });

  test('should create supplier', async () => {
    await page.goto(BASE_URL + '/suppliers');
    await page.click('button:has-text("Add Supplier")');
    await page.waitForSelector('#supplierModal', { timeout: 3000 });
  });
});

describe('Purchase Orders', () => {
  test.beforeEach(async () => {
    await setup();
    await login();
  });

  test('should list purchase orders', async () => {
    await page.goto(BASE_URL + '/purchase-orders');
    await page.waitForSelector('table', { timeout: 5000 });
  });
});

describe('Production', () => {
  test.beforeEach(async () => {
    await setup();
    await login();
  });

  test('should list production orders', async () => {
    await page.goto(BASE_URL + '/production');
    await page.waitForSelector('table', { timeout: 5000 });
  });
});

describe('BOM', () => {
  test.beforeEach(async () => {
    await setup();
    await login();
  });

  test('should list BOMs', async () => {
    await page.goto(BASE_URL + '/bom');
    await page.waitForSelector('table', { timeout: 5000 });
  });
});

describe('Transactions', () => {
  test.beforeEach(async () => {
    await setup();
    await login();
  });

  test('should list transactions', async () => {
    await page.goto(BASE_URL + '/transactions');
    await page.waitForSelector('table', { timeout: 5000 });
  });
});

describe('Reports', () => {
  test.beforeEach(async () => {
    await setup();
    await login();
  });

  test('should load reports page', async () => {
    await page.goto(BASE_URL + '/reports');
    await page.waitForSelector('.card', { timeout: 5000 });
  });
});

describe('Settings', () => {
  test.beforeEach(async () => {
    await setup();
    await login();
  });

  test('should load settings page', async () => {
    await page.goto(BASE_URL + '/settings');
    await page.waitForSelector('form', { timeout: 5000 });
  });

  test('should list users in settings', async () => {
    await page.goto(BASE_URL + '/settings');
    await page.waitForSelector('#usersTable', { timeout: 5000 });
  });
});

describe('Stock Transfers', () => {
  test.beforeEach(async () => {
    await setup();
    await login();
  });

  test('should load transfers page', async () => {
    await page.goto(BASE_URL + '/transfers');
    await page.waitForSelector('table', { timeout: 5000 });
  });

  test('should open create transfer modal', async () => {
    await page.goto(BASE_URL + '/transfers');
    await page.click('button:has-text("New Transfer")');
    await page.waitForSelector('#transferModal', { timeout: 3000 });
  });
});

describe('Stock Audits', () => {
  test.beforeEach(async () => {
    await setup();
    await login();
  });

  test('should load audits page', async () => {
    await page.goto(BASE_URL + '/audits');
    await page.waitForSelector('table', { timeout: 5000 });
  });
});

describe('Import/Export', () => {
  test.beforeEach(async () => {
    await setup();
    await login();
  });

  test('should load import page', async () => {
    await page.goto(BASE_URL + '/import');
    await page.waitForSelector('form', { timeout: 5000 });
  });

  test('should load export page', async () => {
    await page.goto(BASE_URL + '/export');
    await page.waitForSelector('.card', { timeout: 5000 });
  });
});

describe('Webhooks', () => {
  test.beforeEach(async () => {
    await setup();
    await login();
  });

  test('should load webhooks page', async () => {
    await page.goto(BASE_URL + '/webhooks');
    await page.waitForSelector('table', { timeout: 5000 });
  });
});

describe('Analytics', () => {
  test.beforeEach(async () => {
    await setup();
    await login();
  });

  test('should load analytics page', async () => {
    await page.goto(BASE_URL + '/analytics');
    await page.waitForSelector('#kpiProducts', { timeout: 10000 });
  });

  test('should display KPI values', async () => {
    await page.goto(BASE_URL + '/analytics');
    const products = await page.textContent('#kpiProducts');
    expect(products).toBeDefined();
  });
});

describe('Audit Logs', () => {
  test.beforeEach(async () => {
    await setup();
    await login();
  });

  test('should load audit logs page', async () => {
    await page.goto(BASE_URL + '/audit-logs');
    await page.waitForSelector('table', { timeout: 5000 });
  });
});

describe('API Endpoints', () => {
  test.beforeEach(async () => {
    await setup();
    await login();
  });

  test('should fetch dashboard stats', async () => {
    const data = await apiCall('/reports/dashboard');
    expect(data).toBeDefined();
  });

  test('should fetch products', async () => {
    const data = await apiCall('/products');
    expect(Array.isArray(data)).toBe(true);
  });

  test('should fetch categories', async () => {
    const data = await apiCall('/categories');
    expect(Array.isArray(data)).toBe(true);
  });

  test('should fetch inventory', async () => {
    const data = await apiCall('/inventory');
    expect(Array.isArray(data)).toBe(true);
  });

  test('should fetch suppliers', async () => {
    const data = await apiCall('/suppliers');
    expect(Array.isArray(data)).toBe(true);
  });

  test('should fetch warehouses', async () => {
    const data = await apiCall('/warehouses');
    expect(Array.isArray(data)).toBe(true);
  });

  test('should fetch analytics dashboard', async () => {
    const data = await apiCall('/analytics/dashboard');
    expect(data).toBeDefined();
    expect(data.kpi).toBeDefined();
  });

  test('should fetch ABC analysis', async () => {
    const data = await apiCall('/analytics/abc');
    expect(Array.isArray(data)).toBe(true);
  });

  test('should fetch trends', async () => {
    const data = await apiCall('/analytics/trends?days=30');
    expect(Array.isArray(data)).toBe(true);
  });

  test('should fetch top movers', async () => {
    const data = await apiCall('/analytics/top-movers?days=30');
    expect(Array.isArray(data)).toBe(true);
  });

  test('should fetch audit logs', async () => {
    const data = await apiCall('/audit-logs');
    expect(Array.isArray(data)).toBe(true);
  });
});

describe('Error Handling', () => {
  test.beforeEach(async () => {
    await setup();
  });

  test('should redirect unauthenticated user to login', async () => {
    await page.goto(BASE_URL + '/dashboard');
    await page.waitForURL('**/login');
  });

  test('should show 401 for unauthorized API call', async () => {
    const response = await fetch(API_BASE + '/products');
    expect(response.status).toBe(401);
  });
});

afterAll(async () => {
  await teardown();
});

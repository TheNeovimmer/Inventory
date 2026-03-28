/**
 * Enterprise Inventory Management System
 * Main Application JavaScript
 */

const API_BASE = '/api';

// ==========================================
// Theme Management
// ==========================================
const ThemeManager = {
    key: 'ims-theme',
    
    init() {
        const saved = localStorage.getItem(this.key);
        if (saved) {
            document.documentElement.setAttribute('data-theme', saved);
        } else if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
            document.documentElement.setAttribute('data-theme', 'dark');
        }
    },
    
    toggle() {
        const current = document.documentElement.getAttribute('data-theme');
        const next = current === 'dark' ? 'light' : 'dark';
        document.documentElement.setAttribute('data-theme', next);
        localStorage.setItem(this.key, next);
        
        // Update charts if they exist
        if (typeof updateChartsTheme === 'function') {
            updateChartsTheme();
        }
    }
};

// ==========================================
// Chart.js Default Configurations
// ==========================================
const ChartDefaults = {
    colors: [
        '#0d9488', '#6366f1', '#10b981', '#f59e0b', 
        '#ef4444', '#06b6d4', '#8b5cf6', '#ec4899'
    ],
    
    getOptions(type = 'bar') {
        const isDark = document.documentElement.getAttribute('data-theme') === 'dark';
        const textColor = isDark ? '#94a3b8' : '#64748b';
        const gridColor = isDark ? '#334155' : '#e2e8f0';
        
        return {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    position: 'bottom',
                    labels: {
                        color: textColor,
                        padding: 20,
                        usePointStyle: true,
                        font: { family: 'Inter', size: 12 }
                    }
                },
                tooltip: {
                    backgroundColor: isDark ? '#1e293b' : '#fff',
                    titleColor: isDark ? '#f8fafc' : '#1e293b',
                    bodyColor: isDark ? '#94a3b8' : '#64748b',
                    borderColor: isDark ? '#334155' : '#e2e8f0',
                    borderWidth: 1,
                    padding: 12,
                    cornerRadius: 8,
                    titleFont: { family: 'Inter', weight: '600' },
                    bodyFont: { family: 'Inter' }
                }
            },
            scales: type !== 'doughnut' && type !== 'pie' ? {
                x: {
                    grid: { color: gridColor, drawBorder: false },
                    ticks: { color: textColor, font: { family: 'Inter' } }
                },
                y: {
                    grid: { color: gridColor, drawBorder: false },
                    ticks: { color: textColor, font: { family: 'Inter' } }
                }
            } : {}
        };
    }
};

// ==========================================
// Authentication
// ==========================================
function getToken() {
    return localStorage.getItem('token');
}

function isAuthenticated() {
    return !!getToken();
}

// ==========================================
// API Utilities
// ==========================================
async function apiCall(url, options = {}) {
    const token = getToken();
    const headers = {
        'Content-Type': 'application/json',
        ...options.headers
    };
    
    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }

    const response = await fetch(API_BASE + url, {
        ...options,
        headers
    });

    if (response.status === 401) {
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        window.location.href = '/login';
        throw new Error('Unauthorized');
    }

    if (!response.ok) {
        const error = await response.json().catch(() => ({ error: 'An error occurred' }));
        throw new Error(error.error || 'Request failed');
    }

    return response.json();
}

// ==========================================
// Formatting Utilities
// ==========================================
function formatCurrency(amount) {
    return new Intl.NumberFormat('en-US', {
        style: 'currency',
        currency: 'USD'
    }).format(amount || 0);
}

function formatDate(dateStr) {
    if (!dateStr) return '-';
    const date = new Date(dateStr);
    return date.toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
    });
}

function formatDateOnly(dateStr) {
    if (!dateStr) return '-';
    const date = new Date(dateStr);
    return date.toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
    });
}

function formatRelativeTime(dateStr) {
    if (!dateStr) return '-';
    const date = new Date(dateStr);
    const now = new Date();
    const diff = now - date;
    
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(diff / 3600000);
    const days = Math.floor(diff / 86400000);
    
    if (minutes < 1) return 'Just now';
    if (minutes < 60) return `${minutes}m ago`;
    if (hours < 24) return `${hours}h ago`;
    if (days < 7) return `${days}d ago`;
    
    return formatDateOnly(dateStr);
}

// ==========================================
// UI Utilities
// ==========================================
function showToast(message, type = 'info') {
    let container = document.getElementById('toastContainer');
    if (!container) {
        container = document.createElement('div');
        container.id = 'toastContainer';
        container.className = 'toast-container';
        document.body.appendChild(container);
    }
    
    const toast = document.createElement('div');
    toast.className = `toast align-items-center text-white bg-${type} border-0 show`;
    toast.setAttribute('role', 'alert');
    toast.setAttribute('aria-live', 'assertive');
    toast.setAttribute('aria-atomic', 'true');
    toast.innerHTML = `
        <div class="d-flex">
            <div class="toast-body">${message}</div>
            <button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast"></button>
        </div>
    `;
    container.appendChild(toast);
    
    setTimeout(() => {
        toast.classList.remove('show');
        setTimeout(() => toast.remove(), 300);
    }, 4000);
}

// ==========================================
// Sidebar
// ==========================================
function initSidebar() {
    const toggleBtn = document.getElementById('toggleSidebar');
    const sidebar = document.getElementById('sidebar');
    
    if (toggleBtn && sidebar) {
        toggleBtn.addEventListener('click', () => {
            sidebar.classList.toggle('collapsed');
        });
    }
}

// ==========================================
// Theme Toggle
// ==========================================
function initThemeToggle() {
    const themeToggle = document.getElementById('themeToggle');
    if (themeToggle) {
        themeToggle.addEventListener('click', () => {
            ThemeManager.toggle();
        });
    }
}

// ==========================================
// Logout
// ==========================================
function initLogout() {
    const logoutBtn = document.getElementById('logoutBtn');
    if (logoutBtn) {
        logoutBtn.addEventListener('click', (e) => {
            e.preventDefault();
            localStorage.removeItem('token');
            localStorage.removeItem('user');
            window.location.href = '/login';
        });
    }
}

// ==========================================
// Authentication Check
// ==========================================
function checkAuth() {
    if (!isAuthenticated() && !window.location.pathname.startsWith('/login') && !window.location.pathname.startsWith('/register')) {
        window.location.href = '/login';
    }
    
    if (isAuthenticated() && (window.location.pathname === '/login' || window.location.pathname === '/register')) {
        window.location.href = '/dashboard';
    }
}

// ==========================================
// HTMX Integration
// ==========================================
function initHTMX() {
    document.body.addEventListener('htmx:configRequest', (event) => {
        const token = getToken();
        if (token) {
            event.detail.headers['Authorization'] = `Bearer ${token}`;
        }
    });

    document.body.addEventListener('htmx:afterSwap', (event) => {
        initBootstrapElements();
    });

    document.body.addEventListener('htmx:responseError', (event) => {
        const response = event.detail.xhr.response;
        try {
            const data = JSON.parse(response);
            showToast(data.error || 'An error occurred', 'danger');
        } catch (e) {
            showToast('An error occurred', 'danger');
        }
    });
}

// ==========================================
// Bootstrap Elements
// ==========================================
function initBootstrapElements() {
    const dropdowns = document.querySelectorAll('[data-bs-toggle="dropdown"]:not(.dropdown-init)');
    dropdowns.forEach(dropdown => {
        dropdown.classList.add('dropdown-init');
        new bootstrap.Dropdown(dropdown);
    });
    
    const tooltips = document.querySelectorAll('[data-bs-toggle="tooltip"]');
    tooltips.forEach(tooltip => {
        new bootstrap.Tooltip(tooltip);
    });
}

// ==========================================
// User Info
// ==========================================
function initUserInfo() {
    const userStr = localStorage.getItem('user');
    if (userStr) {
        try {
            const user = JSON.parse(userStr);
            const userEl = document.getElementById('currentUser');
            if (userEl) userEl.textContent = user.username;
            
            // Update avatar
            const avatar = document.querySelector('.user-avatar');
            if (avatar) {
                avatar.textContent = user.username.charAt(0).toUpperCase();
            }
        } catch (e) {}
    }
}

// ==========================================
// Search Debounce
// ==========================================
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

// ==========================================
// Chart Functions
// ==========================================
let categoryChart = null;
let stockChart = null;
let turnoverChart = null;

function renderCategoryChart(data, canvasId = 'categoryChart') {
    const ctx = document.getElementById(canvasId);
    if (!ctx) return;
    
    if (categoryChart) {
        categoryChart.destroy();
    }
    
    if (!data || data.length === 0) {
        ctx.style.display = 'none';
        return;
    }
    
    ctx.style.display = 'block';
    
    categoryChart = new Chart(ctx.getContext('2d'), {
        type: 'doughnut',
        data: {
            labels: data.map(d => d.category_name),
            datasets: [{
                data: data.map(d => d.total_value),
                backgroundColor: ChartDefaults.colors,
                borderWidth: 0,
                hoverOffset: 8
            }]
        },
        options: {
            ...ChartDefaults.getOptions('doughnut'),
            cutout: '65%'
        }
    });
}

function renderStockChart(data, canvasId = 'stockChart') {
    const ctx = document.getElementById(canvasId);
    if (!ctx) return;
    
    if (stockChart) {
        stockChart.destroy();
    }
    
    if (!data || data.length === 0) {
        ctx.style.display = 'none';
        return;
    }
    
    ctx.style.display = 'block';
    
    const topProducts = data.slice(0, 10);
    
    stockChart = new Chart(ctx.getContext('2d'), {
        type: 'bar',
        data: {
            labels: topProducts.map(d => d.product_name.substring(0, 20)),
            datasets: [{
                label: 'Stock Quantity',
                data: topProducts.map(d => d.quantity),
                backgroundColor: topProducts.map(d => {
                    if (d.status === 'Out of Stock') return '#ef4444';
                    if (d.status === 'Low Stock') return '#f59e0b';
                    return '#10b981';
                }),
                borderRadius: 6,
                barThickness: 24
            }]
        },
        options: {
            ...ChartDefaults.getOptions('bar'),
            indexAxis: 'y',
            plugins: {
                ...ChartDefaults.getOptions('bar').plugins,
                legend: { display: false }
            }
        }
    });
}

function renderTurnoverChart(data, canvasId = 'turnoverChart') {
    const ctx = document.getElementById(canvasId);
    if (!ctx) return;
    
    if (turnoverChart) {
        turnoverChart.destroy();
    }
    
    if (!data || data.length === 0) {
        ctx.style.display = 'none';
        return;
    }
    
    ctx.style.display = 'block';
    
    turnoverChart = new Chart(ctx.getContext('2d'), {
        type: 'line',
        data: {
            labels: data.map(d => d.date),
            datasets: [
                {
                    label: 'Stock In',
                    data: data.map(d => d.total_in),
                    borderColor: '#10b981',
                    backgroundColor: 'rgba(16, 185, 129, 0.1)',
                    fill: true,
                    tension: 0.4,
                    pointRadius: 3,
                    pointHoverRadius: 6
                },
                {
                    label: 'Stock Out',
                    data: data.map(d => d.total_out),
                    borderColor: '#ef4444',
                    backgroundColor: 'rgba(239, 68, 68, 0.1)',
                    fill: true,
                    tension: 0.4,
                    pointRadius: 3,
                    pointHoverRadius: 6
                }
            ]
        },
        options: ChartDefaults.getOptions('line')
    });
}

function updateChartsTheme() {
    // Re-render charts with new theme
    if (typeof loadCategoryData === 'function') loadCategoryData();
    if (typeof loadStockChart === 'function') loadStockChart();
    if (typeof loadTurnoverChart === 'function') loadTurnoverChart();
}

// ==========================================
// Auto-refresh Polling
// ==========================================
let pollingInterval = null;

function startPolling(callback, interval = 30000) {
    callback();
    pollingInterval = setInterval(callback, interval);
}

function stopPolling() {
    if (pollingInterval) {
        clearInterval(pollingInterval);
        pollingInterval = null;
    }
}

// ==========================================
// Initialize
// ==========================================
document.addEventListener('DOMContentLoaded', () => {
    ThemeManager.init();
    checkAuth();
    initSidebar();
    initThemeToggle();
    initLogout();
    initHTMX();
    initBootstrapElements();
    initUserInfo();
});

window.addEventListener('storage', (e) => {
    if (e.key === 'token' && !e.newValue) {
        window.location.href = '/login';
    }
});

// Expose functions globally
window.showToast = showToast;
window.formatCurrency = formatCurrency;
window.formatDate = formatDate;
window.formatDateOnly = formatDateOnly;
window.formatRelativeTime = formatRelativeTime;
window.apiCall = apiCall;
window.renderCategoryChart = renderCategoryChart;
window.renderStockChart = renderStockChart;
window.renderTurnoverChart = renderTurnoverChart;
window.startPolling = startPolling;
window.stopPolling = stopPolling;
window.debounce = debounce;

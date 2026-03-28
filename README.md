# Inventory Management System (IMS)

A full-stack inventory management system for small to medium enterprises built with Go, SQLite, and HTMX.

## Features

### Core Modules
- **Products** - Full CRUD with SKU, barcode, pricing, and reorder points
- **Categories** - Product categorization with parent/child support
- **Inventory** - Real-time stock tracking with low-stock alerts
- **Production** - Bill of Materials (BOM) and production order workflow

### Procurement
- **Suppliers** - Supplier management with contact details
- **Purchase Orders** - Create, track, and receive purchase orders

### Analytics & Reporting
- **Transactions** - Complete audit trail of all inventory movements
- **Reports** - Stock levels, inventory valuation, low-stock alerts, turnover analysis

### System
- **Authentication** - JWT-based login with role-based access control
- **User Management** - Admin, Manager, and Staff roles

## Tech Stack

- **Backend**: Go (Golang) with Gin framework
- **Database**: SQLite with GORM ORM
- **Frontend**: HTML, CSS, Bootstrap 5, HTMX

## Getting Started

### Prerequisites
- Go 1.21+
- SQLite3

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd inventory
```

2. Install dependencies:
```bash
go mod download
```

3. Run the server:
```bash
go run cmd/server/main.go
```

4. Seed the admin user:
```bash
curl http://localhost:8080/api/auth/seed
```

### Default Credentials
- Email: `admin@inventory.com`
- Password: `admin123`

### Access the Application
- Web UI: http://localhost:8080
- Login: http://localhost:8080/login

## API Endpoints

### Authentication
- `POST /api/auth/login` - User login
- `POST /api/auth/register` - User registration
- `GET /api/auth/me` - Get current user
- `GET /api/auth/seed` - Create admin user

### Products
- `GET /api/products` - List all products
- `GET /api/products/:id` - Get product
- `POST /api/products` - Create product
- `PUT /api/products/:id` - Update product
- `DELETE /api/products/:id` - Delete product

### Categories
- `GET /api/categories` - List categories
- `POST /api/categories` - Create category
- `PUT /api/categories/:id` - Update category
- `DELETE /api/categories/:id` - Delete category

### Inventory
- `GET /api/inventory` - List inventory
- `POST /api/inventory/adjust` - Adjust stock
- `GET /api/inventory/alerts` - Get low stock alerts
- `GET /api/inventory/history` - Get stock history

### Bill of Materials
- `GET /api/bom` - List BOMs
- `GET /api/bom/:id` - Get BOM details
- `POST /api/bom` - Create BOM
- `DELETE /api/bom/:id` - Delete BOM

### Production Orders
- `GET /api/production-orders` - List orders
- `POST /api/production-orders` - Create order
- `POST /api/production-orders/:id/start` - Start production
- `POST /api/production-orders/:id/complete` - Complete production
- `POST /api/production-orders/:id/cancel` - Cancel order

### Suppliers
- `GET /api/suppliers` - List suppliers
- `POST /api/suppliers` - Create supplier
- `PUT /api/suppliers/:id` - Update supplier
- `DELETE /api/suppliers/:id` - Delete supplier

### Purchase Orders
- `GET /api/purchase-orders` - List POs
- `GET /api/purchase-orders/:id` - Get PO details
- `POST /api/purchase-orders` - Create PO
- `PUT /api/purchase-orders/:id/status` - Update PO status
- `POST /api/purchase-orders/:id/receive` - Receive PO

### Users
- `GET /api/users` - List users
- `POST /api/users` - Create user
- `PUT /api/users/:id` - Update user
- `DELETE /api/users/:id` - Delete user

### Reports
- `GET /api/reports/dashboard` - Dashboard stats
- `GET /api/reports/stock-levels` - Stock levels
- `GET /api/reports/valuation` - Inventory valuation
- `GET /api/reports/low-stock` - Low stock items
- `GET /api/reports/turnover` - Turnover report
- `GET /api/reports/transactions` - Transaction history
- `GET /api/reports/export/stock-levels` - Export CSV

## Project Structure

```
inventory/
├── cmd/
│   └── server/
│       └── main.go              # Entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration
│   ├── database/
│   │   └── database.go          # Database connection
│   ├── handlers/
│   │   ├── auth.go              # Auth & user handlers
│   │   ├── product.go           # Product handlers
│   │   ├── category.go          # Category handlers
│   │   ├── inventory.go         # Inventory handlers
│   │   ├── production.go        # Production handlers
│   │   ├── supplier.go          # Supplier & PO handlers
│   │   └── report.go            # Report handlers
│   ├── middleware/
│   │   └── auth.go              # JWT middleware
│   └── models/
│       ├── user.go
│       ├── product.go
│       ├── category.go
│       ├── inventory.go
│       ├── production.go
│       ├── supplier.go
│       └── transaction.go
├── web/
│   ├── static/
│   │   ├── css/
│   │   │   └── styles.css
│   │   └── js/
│   │       └── app.js
│   └── templates/
│       ├── layout.html
│       ├── login.html
│       ├── register.html
│       ├── dashboard.html
│       ├── categories.html
│       ├── products/
│       │   └── list.html
│       ├── inventory/
│       │   └── index.html
│       ├── production/
│       │   ├── bom.html
│       │   └── orders.html
│       ├── suppliers/
│       │   ├── index.html
│       │   └── purchase-orders.html
│       ├── reports/
│       │   ├── index.html
│       │   └── transactions.html
│       └── settings/
│           └── index.html
├── go.mod
├── go.sum
└── README.md
```

## User Roles

| Role   | Description                    |
|--------|--------------------------------|
| Admin  | Full system access             |
| Manager| Manage products, orders, reports|
| Staff  | View and basic operations     |

## License

MIT

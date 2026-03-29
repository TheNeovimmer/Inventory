package handlers

import (
	"net/http"
	"strconv"
	"time"

	"inventory-ims/internal/database"
	"inventory-ims/internal/models"

	"github.com/gin-gonic/gin"
)

type AccountHandler struct{}

func NewAccountHandler() *AccountHandler {
	return &AccountHandler{}
}

type CreateAccountRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	AccountType string `json:"account_type" binding:"required"`
	ParentID    *uint  `json:"parent_id"`
	Description string `json:"description"`
}

type CreateTransactionRequest struct {
	Date        time.Time `json:"date"`
	Description string    `json:"description" binding:"required"`
	Type        string    `json:"type" binding:"required"`
	Amount      float64   `json:"amount" binding:"required,min=1"`
	AccountID   uint      `json:"account_id" binding:"required"`
	Reference   string    `json:"reference"`
	Notes       string    `json:"notes"`
}

func (h *AccountHandler) ListAccounts(c *gin.Context) {
	var accounts []models.Account
	query := database.DB.Where("is_active = ?", true)

	accountType := c.Query("type")
	if accountType != "" {
		query = query.Where("account_type = ?", accountType)
	}

	if err := query.Order("account_type, code").Find(&accounts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch accounts"})
		return
	}

	c.JSON(http.StatusOK, accounts)
}

func (h *AccountHandler) GetAccount(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
		return
	}

	var account models.Account
	if err := database.DB.First(&account, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	c.JSON(http.StatusOK, account)
}

func (h *AccountHandler) CreateAccount(c *gin.Context) {
	var req CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.Account
	if err := database.DB.Where("code = ?", req.Code).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Account code already exists"})
		return
	}

	account := models.Account{
		Name:        req.Name,
		Code:        req.Code,
		AccountType: req.AccountType,
		ParentID:    req.ParentID,
		Description: req.Description,
		IsActive:    true,
	}

	if err := database.DB.Create(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account"})
		return
	}

	c.JSON(http.StatusCreated, account)
}

func (h *AccountHandler) UpdateAccount(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
		return
	}

	var account models.Account
	if err := database.DB.First(&account, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	var req CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{
		"name":         req.Name,
		"account_type": req.AccountType,
		"description":  req.Description,
	}

	if req.ParentID != nil {
		updates["parent_id"] = *req.ParentID
	}

	if err := database.DB.Model(&account).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account"})
		return
	}

	c.JSON(http.StatusOK, account)
}

func (h *AccountHandler) DeleteAccount(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
		return
	}

	var account models.Account
	if err := database.DB.First(&account, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	if err := database.DB.Model(&account).Update("is_active", false).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account deleted"})
}

func (h *AccountHandler) ListTransactions(c *gin.Context) {
	var transactions []models.AccountTransaction
	query := database.DB.Preload("Account").Preload("Creator")

	accountID := c.Query("account_id")
	if accountID != "" {
		id, _ := strconv.ParseUint(accountID, 10, 32)
		query = query.Where("account_id = ?", id)
	}

	txType := c.Query("type")
	if txType != "" {
		query = query.Where("type = ?", txType)
	}

	startDate := c.Query("start_date")
	if startDate != "" {
		query = query.Where("date >= ?", startDate)
	}

	endDate := c.Query("end_date")
	if endDate != "" {
		query = query.Where("date <= ?", endDate)
	}

	limit := c.DefaultQuery("limit", "100")
	l, _ := strconv.Atoi(limit)
	query = query.Order("date DESC").Limit(l)

	if err := query.Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	var response []models.AccountTransactionResponse
	for _, t := range transactions {
		creatorName := ""
		if t.Creator != nil {
			creatorName = t.Creator.Username
		}

		accountName := ""
		if t.Account != nil {
			accountName = t.Account.Name
		}

		response = append(response, models.AccountTransactionResponse{
			ID:          t.ID,
			Date:        t.Date.Format("2006-01-02"),
			Description: t.Description,
			Type:        t.Type,
			Amount:      t.Amount,
			AccountID:   t.AccountID,
			AccountName: accountName,
			Reference:   t.Reference,
			Notes:       t.Notes,
			CreatedBy:   t.CreatedBy,
			CreatorName: creatorName,
			CreatedAt:   t.CreatedAt.Format("2006-01-02 15:04"),
		})
	}

	c.JSON(http.StatusOK, response)
}

func (h *AccountHandler) CreateTransaction(c *gin.Context) {
	var req CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")

	date := req.Date
	if date.IsZero() {
		date = time.Now()
	}

	transaction := models.AccountTransaction{
		Date:        date,
		Description: req.Description,
		Type:        req.Type,
		Amount:      req.Amount,
		AccountID:   req.AccountID,
		Reference:   req.Reference,
		Notes:       req.Notes,
		CreatedBy:   userID.(uint),
	}

	if err := database.DB.Create(&transaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
		return
	}

	var account models.Account
	database.DB.First(&account, req.AccountID)

	switch req.Type {
	case models.AccountTransactionTypeDeposit, models.AccountTransactionTypeAdjustment:
		account.Balance += req.Amount
	case models.AccountTransactionTypeExpense:
		account.Balance -= req.Amount
	}
	database.DB.Save(&account)

	logAudit(c, "account_transaction", transaction.ID, models.AuditActionCreate, "", "Created transaction: "+req.Description, userID.(uint))

	c.JSON(http.StatusCreated, transaction)
}

func (h *AccountHandler) GetSummary(c *gin.Context) {
	type Summary struct {
		TotalIncome    float64 `json:"total_income"`
		TotalExpense   float64 `json:"total_expense"`
		NetProfit      float64 `json:"net_profit"`
		TotalAssets    float64 `json:"total_assets"`
		TotalLiability float64 `json:"total_liability"`
	}

	var summary Summary

	database.DB.Model(&models.AccountTransaction{}).Where("type = ?", models.AccountTransactionTypeDeposit).Select("COALESCE(SUM(amount), 0)").Scan(&summary.TotalIncome)

	database.DB.Model(&models.AccountTransaction{}).Where("type = ?", models.AccountTransactionTypeExpense).Select("COALESCE(SUM(amount), 0)").Scan(&summary.TotalExpense)

	summary.NetProfit = summary.TotalIncome - summary.TotalExpense

	database.DB.Model(&models.Account{}).Where("account_type = ? AND is_active = ?", models.AccountTypeAsset, true).Select("COALESCE(SUM(balance), 0)").Scan(&summary.TotalAssets)

	database.DB.Model(&models.Account{}).Where("account_type = ? AND is_active = ?", models.AccountTypeLiability, true).Select("COALESCE(SUM(balance), 0)").Scan(&summary.TotalLiability)

	c.JSON(http.StatusOK, summary)
}

func (h *AccountHandler) ListPaymentMethods(c *gin.Context) {
	var methods []models.PaymentMethod
	database.DB.Where("is_active = ?", true).Order("is_default DESC, name").Find(&methods)

	c.JSON(http.StatusOK, methods)
}

func (h *AccountHandler) CreatePaymentMethod(c *gin.Context) {
	var req struct {
		Name      string `json:"name" binding:"required"`
		Code      string `json:"code" binding:"required"`
		IsDefault bool   `json:"is_default"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.IsDefault {
		database.DB.Model(&models.PaymentMethod{}).Where("is_default = ?", true).Update("is_default", false)
	}

	method := models.PaymentMethod{
		Name:      req.Name,
		Code:      req.Code,
		IsDefault: req.IsDefault,
		IsActive:  true,
	}

	if err := database.DB.Create(&method).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment method"})
		return
	}

	c.JSON(http.StatusCreated, method)
}

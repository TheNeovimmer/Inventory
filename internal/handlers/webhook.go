package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"inventory-ims/internal/database"
	"inventory-ims/internal/models"

	"github.com/gin-gonic/gin"
)

type WebhookHandler struct{}

func NewWebhookHandler() *WebhookHandler {
	return &WebhookHandler{}
}

type CreateWebhookRequest struct {
	Name       string   `json:"name" binding:"required"`
	URL        string   `json:"url" binding:"required"`
	Events     []string `json:"events" binding:"required"`
	Secret     string   `json:"secret"`
	RetryCount int      `json:"retry_count"`
	Timeout    int      `json:"timeout"`
}

func (h *WebhookHandler) List(c *gin.Context) {
	var webhooks []models.Webhook
	database.DB.Order("created_at DESC").Find(&webhooks)

	var response []models.WebhookResponse
	for _, wh := range webhooks {
		events := []string{}
		if wh.Events != "" {
			events = strings.Split(wh.Events, ",")
		}
		response = append(response, models.WebhookResponse{
			ID:         wh.ID,
			Name:       wh.Name,
			URL:        wh.URL,
			Events:     events,
			IsActive:   wh.IsActive,
			RetryCount: wh.RetryCount,
			Timeout:    wh.Timeout,
			CreatedBy:  wh.CreatedBy,
			CreatedAt:  wh.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, response)
}

func (h *WebhookHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook ID"})
		return
	}

	var webhook models.Webhook
	if err := database.DB.First(&webhook, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		return
	}

	events := []string{}
	if webhook.Events != "" {
		events = strings.Split(webhook.Events, ",")
	}

	c.JSON(http.StatusOK, models.WebhookResponse{
		ID:         webhook.ID,
		Name:       webhook.Name,
		URL:        webhook.URL,
		Events:     events,
		IsActive:   webhook.IsActive,
		RetryCount: webhook.RetryCount,
		Timeout:    webhook.Timeout,
		CreatedBy:  webhook.CreatedBy,
		CreatedAt:  webhook.CreatedAt,
	})
}

func (h *WebhookHandler) Create(c *gin.Context) {
	var req CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")

	retryCount := req.RetryCount
	if retryCount == 0 {
		retryCount = 3
	}

	timeout := req.Timeout
	if timeout == 0 {
		timeout = 30
	}

	webhook := models.Webhook{
		Name:       req.Name,
		URL:        req.URL,
		Events:     strings.Join(req.Events, ","),
		Secret:     req.Secret,
		IsActive:   true,
		RetryCount: retryCount,
		Timeout:    timeout,
		CreatedBy:  userID.(uint),
	}

	if err := database.DB.Create(&webhook).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create webhook"})
		return
	}

	logAudit(c, "webhook", webhook.ID, models.AuditActionCreate, "", "Created webhook: "+webhook.Name, userID.(uint))

	c.JSON(http.StatusCreated, webhook)
}

func (h *WebhookHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook ID"})
		return
	}

	var webhook models.Webhook
	if err := database.DB.First(&webhook, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		return
	}

	var req CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{
		"name":        req.Name,
		"url":         req.URL,
		"events":      strings.Join(req.Events, ","),
		"retry_count": req.RetryCount,
		"timeout":     req.Timeout,
	}

	if req.Secret != "" {
		updates["secret"] = req.Secret
	}

	if err := database.DB.Model(&webhook).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update webhook"})
		return
	}

	c.JSON(http.StatusOK, webhook)
}

func (h *WebhookHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook ID"})
		return
	}

	userID, _ := c.Get("user_id")

	var webhook models.Webhook
	if err := database.DB.First(&webhook, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		return
	}

	if err := database.DB.Delete(&webhook).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete webhook"})
		return
	}

	logAudit(c, "webhook", webhook.ID, models.AuditActionDelete, "", "Deleted webhook: "+webhook.Name, userID.(uint))

	c.JSON(http.StatusOK, gin.H{"message": "Webhook deleted"})
}

func (h *WebhookHandler) Toggle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook ID"})
		return
	}

	var webhook models.Webhook
	if err := database.DB.First(&webhook, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		return
	}

	if err := database.DB.Model(&webhook).Update("is_active", !webhook.IsActive).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle webhook"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook toggled", "is_active": !webhook.IsActive})
}

func (h *WebhookHandler) GetDeliveries(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook ID"})
		return
	}

	var deliveries []models.WebhookDelivery
	database.DB.Where("webhook_id = ?", id).Order("created_at DESC").Limit(50).Find(&deliveries)

	c.JSON(http.StatusOK, deliveries)
}

func (h *WebhookHandler) Test(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook ID"})
		return
	}

	var webhook models.Webhook
	if err := database.DB.First(&webhook, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		return
	}

	payload := map[string]interface{}{
		"event":   "test",
		"message": "This is a test webhook",
		"time":    time.Now().Format(time.RFC3339),
	}

	payloadJSON, _ := json.Marshal(payload)

	delivery := models.WebhookDelivery{
		WebhookID: webhook.ID,
		Event:     "test",
		Payload:   string(payloadJSON),
	}

	client := &http.Client{Timeout: time.Duration(webhook.Timeout) * time.Second}
	req, _ := http.NewRequest("POST", webhook.URL, bytes.NewBuffer(payloadJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Event", "test")
	req.Header.Set("X-Webhook-ID", strconv.Itoa(int(webhook.ID)))

	if webhook.Secret != "" {
		mac := hmac.New(sha256.New, []byte(webhook.Secret))
		mac.Write(payloadJSON)
		signature := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Webhook-Signature", signature)
	}

	resp, err := client.Do(req)
	if err != nil {
		delivery.Success = false
		delivery.ErrorMsg = err.Error()
		database.DB.Create(&delivery)
		c.JSON(http.StatusOK, gin.H{"success": false, "error": err.Error()})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	delivery.StatusCode = resp.StatusCode
	delivery.Response = string(body)
	delivery.Success = resp.StatusCode >= 200 && resp.StatusCode < 300

	if !delivery.Success {
		delivery.ErrorMsg = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	database.DB.Create(&delivery)

	c.JSON(http.StatusOK, gin.H{
		"success":     delivery.Success,
		"status_code": resp.StatusCode,
		"response":    string(body),
	})
}

func TriggerWebhook(event string, data interface{}) {
	var webhooks []models.Webhook
	database.DB.Where("is_active = ? AND events LIKE ?", true, "%"+event+"%").Find(&webhooks)

	if len(webhooks) == 0 {
		return
	}

	payload := map[string]interface{}{
		"event": event,
		"data":  data,
		"time":  time.Now().Format(time.RFC3339),
	}

	payloadJSON, _ := json.Marshal(payload)

	for _, webhook := range webhooks {
		go func(wh models.Webhook) {
			delivery := models.WebhookDelivery{
				WebhookID: wh.ID,
				Event:     event,
				Payload:   string(payloadJSON),
			}

			client := &http.Client{Timeout: time.Duration(wh.Timeout) * time.Second}
			req, _ := http.NewRequest("POST", wh.URL, bytes.NewBuffer(payloadJSON))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Webhook-Event", event)
			req.Header.Set("X-Webhook-ID", strconv.Itoa(int(wh.ID)))

			if wh.Secret != "" {
				mac := hmac.New(sha256.New, []byte(wh.Secret))
				mac.Write(payloadJSON)
				signature := hex.EncodeToString(mac.Sum(nil))
				req.Header.Set("X-Webhook-Signature", signature)
			}

			resp, err := client.Do(req)
			if err != nil {
				delivery.Success = false
				delivery.ErrorMsg = err.Error()
				database.DB.Create(&delivery)
				return
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			delivery.StatusCode = resp.StatusCode
			delivery.Response = string(body)
			delivery.Success = resp.StatusCode >= 200 && resp.StatusCode < 300

			if !delivery.Success {
				delivery.ErrorMsg = fmt.Sprintf("HTTP %d", resp.StatusCode)
			}

			database.DB.Create(&delivery)
		}(webhook)
	}
}

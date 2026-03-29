package handlers

import (
	"net/http"

	"inventory-ims/internal/database"
	"inventory-ims/internal/models"

	"github.com/gin-gonic/gin"
)

type SettingHandler struct{}

func NewSettingHandler() *SettingHandler {
	return &SettingHandler{}
}

func (h *SettingHandler) GetAll(c *gin.Context) {
	var settings []models.Setting
	database.DB.Find(&settings)

	settingMap := make(map[string]string)
	for _, s := range settings {
		settingMap[s.Key] = s.Value
	}

	var result []models.SettingGroup
	for _, group := range models.DefaultSettings {
		newGroup := models.SettingGroup{
			Key:         group.Key,
			Name:        group.Name,
			Description: group.Description,
			Settings:    []models.SettingItem{},
		}

		for _, item := range group.Settings {
			value := settingMap[item.Key]
			if value == "" {
				value = item.Value
			}
			item.Value = value
			newGroup.Settings = append(newGroup.Settings, item)
		}

		result = append(result, newGroup)
	}

	c.JSON(http.StatusOK, result)
}

func (h *SettingHandler) Get(c *gin.Context) {
	key := c.Param("key")

	var setting models.Setting
	if err := database.DB.First(&setting, key).Error; err != nil {
		for _, group := range models.DefaultSettings {
			for _, item := range group.Settings {
				if item.Key == key {
					c.JSON(http.StatusOK, gin.H{"key": key, "value": item.Value})
					return
				}
			}
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Setting not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"key": setting.Key, "value": setting.Value})
}

func (h *SettingHandler) Update(c *gin.Context) {
	var req struct {
		Key   string `json:"key" binding:"required"`
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	setting := models.Setting{
		Key:   req.Key,
		Value: req.Value,
	}

	result := database.DB.Where("key = ?", req.Key).Assign(setting).FirstOrCreate(&setting)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update setting"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Setting updated", "key": req.Key, "value": req.Value})
}

func (h *SettingHandler) UpdateMultiple(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for key, value := range req {
		setting := models.Setting{
			Key:   key,
			Value: value,
		}
		database.DB.Where("key = ?", key).Assign(setting).FirstOrCreate(&setting)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Settings updated"})
}

func (h *SettingHandler) GetCurrency(c *gin.Context) {
	result := make(map[string]string)

	var setting models.Setting
	if err := database.DB.Where("key = ?", "currency_symbol").First(&setting).Error; err == nil {
		result["symbol"] = setting.Value
	} else {
		result["symbol"] = "$"
	}

	if err := database.DB.Where("key = ?", "currency_position").First(&setting).Error; err == nil {
		result["position"] = setting.Value
	} else {
		result["position"] = "before"
	}

	if err := database.DB.Where("key = ?", "default_tax_rate").First(&setting).Error; err == nil {
		result["tax_rate"] = setting.Value
	} else {
		result["tax_rate"] = "0"
	}

	c.JSON(http.StatusOK, result)
}

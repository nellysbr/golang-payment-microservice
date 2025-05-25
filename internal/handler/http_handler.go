package handler

import (
	"net/http"
	"strconv"

	"golang-payment-microservice/internal/model"
	"golang-payment-microservice/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type HTTPHandler struct {
	paymentService service.PaymentService
	logger         *logrus.Logger
}

func NewHTTPHandler(paymentService service.PaymentService, logger *logrus.Logger) *HTTPHandler {
	return &HTTPHandler{
		paymentService: paymentService,
		logger:         logger,
	}
}

func (h *HTTPHandler) SetupRoutes() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	
	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(h.corsMiddleware())

	// Health check
	router.GET("/health", h.healthCheck)

	// Payment routes
	v1 := router.Group("/api/v1")
	{
		v1.POST("/payments", h.createPayment)
		v1.GET("/payments/:id", h.getPayment)
		v1.GET("/merchants/:merchant_id/payments", h.getPaymentsByMerchant)
	}

	return router
}

func (h *HTTPHandler) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "payment-microservice",
	})
}

func (h *HTTPHandler) createPayment(c *gin.Context) {
	var req model.PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Validação básica
	if req.CardNumber == "" || req.Amount <= 0 || req.Currency == "" || req.MerchantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required fields",
		})
		return
	}

	response, err := h.paymentService.CreatePayment(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create payment")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *HTTPHandler) getPayment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid payment ID",
		})
		return
	}

	payment, err := h.paymentService.GetPayment(c.Request.Context(), id)
	if err != nil {
		h.logger.WithError(err).WithField("payment_id", id).Error("Failed to get payment")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Payment not found",
		})
		return
	}

	c.JSON(http.StatusOK, payment)
}

func (h *HTTPHandler) getPaymentsByMerchant(c *gin.Context) {
	merchantID := c.Param("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Merchant ID is required",
		})
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	payments, err := h.paymentService.GetPaymentsByMerchant(c.Request.Context(), merchantID, limit, offset)
	if err != nil {
		h.logger.WithError(err).WithField("merchant_id", merchantID).Error("Failed to get payments by merchant")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve payments",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"payments": payments,
		"limit":    limit,
		"offset":   offset,
		"count":    len(payments),
	})
}

func (h *HTTPHandler) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
} 
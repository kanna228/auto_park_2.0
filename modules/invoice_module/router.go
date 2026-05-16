package invoice_module

import (
	"auto_park/internal/config"
	"auto_park/middleware"
	"auto_park/modules/invoice_module/handlers"
	"auto_park/modules/invoice_module/service"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, cfg *config.Config, svc service.InvoiceService) {
	h := handlers.NewInvoiceHandler(svc)

	api := r.Group("/api/warehouse/invoices")
	api.Use(middleware.AuthJWT(cfg), middleware.RequireRoles(1, 5))
	{
		api.GET("", h.ListInvoices)
		api.GET("/export", h.ExportInvoices)
		api.GET("/:id", h.GetInvoiceByID)
		api.POST("", h.CreateInvoice)
	}
}

package router

// thuận
import (
	"net/http"

	_ "sms-gateway/docs"
	"sms-gateway/src/api/middleware"
	"sms-gateway/src/internal/config"
	"sms-gateway/src/internal/handler"
	"sms-gateway/src/internal/service"

	httpSwagger "github.com/swaggo/http-swagger"
)

// NewRouter creates a new HTTP router with all routes configured
func NewRouter(cfg *config.Config, smsService *service.SMSService) http.Handler {
	mux := http.NewServeMux()
	// Create handlers
	smsHandler := handler.NewSMSHandler(cfg, smsService)

	// API v1 routes
	mux.HandleFunc("/", smsHandler.HandleRoot)
	mux.HandleFunc("/api/v1/health", smsHandler.HandleHealth)
	mux.HandleFunc("/api/v1/sms/send", smsHandler.HandleSendSMS)
	mux.HandleFunc("/api/v1/modem/test", smsHandler.HandleTestModem)
	mux.HandleFunc("/api/v1/ports", smsHandler.HandleListPorts)
	mux.HandleFunc("/api/v1/ports/status", smsHandler.HandlePortStatus)
	mux.HandleFunc("/api/v1/modem/info", smsHandler.HandleModemInfo)

	// Swagger documentation
	mux.HandleFunc("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	))

	// Legacy routes for backward compatibility
	mux.HandleFunc("/health", smsHandler.HandleHealth)
	mux.HandleFunc("/send", smsHandler.HandleSendSMS)
	mux.HandleFunc("/port/status", smsHandler.HandlePortStatus)

	// Apply middleware
	var handler http.Handler = mux
	handler = middleware.Recovery(handler)
	handler = middleware.CORS(handler)
	handler = middleware.Logger(handler)

	return handler
}

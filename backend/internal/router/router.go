package router

import (
	"log"
	"net/http"
	"time"

	"github.com/FTATM/iot-monitor-interface/internal/handler"
	"github.com/golang-jwt/jwt/v5"
)

// AllHandlers bundles everything together so Setup() stays clean
type AllHandlers struct {
	Widget     *handler.WidgetHandler
	Canvas     *handler.CanvasHandler
	WidgetType *handler.WidgetTypeHandler
	User       *handler.UserHandler
	// Add more as the app grows...
}

// Setup now takes the bundle and applies middleware
func Setup(handlers AllHandlers) *http.ServeMux {
	mux := http.NewServeMux()

	widgetMux := http.NewServeMux()
	mux.Handle("/widget/", http.StripPrefix("/widget", widgetMux))
	widgetMux.HandleFunc("GET /getbyid/{id}", handlers.Widget.GetById)
	widgetMux.HandleFunc("POST /create", handlers.Widget.Create)
	widgetMux.HandleFunc("PUT /update", handlers.Widget.Update)
	widgetMux.HandleFunc("DELETE /delete", handlers.Widget.Delete)

	widgetTypeMux := http.NewServeMux()
	mux.Handle("/widgettype/", http.StripPrefix("/widgettype", widgetTypeMux))
	widgetTypeMux.HandleFunc("GET /getbyid", handlers.WidgetType.GetById)
	widgetTypeMux.HandleFunc("GET /getall", handlers.WidgetType.GetAll)

	canvasMux := http.NewServeMux()
	mux.Handle("/canvas/", http.StripPrefix("/canvas", canvasMux))
	canvasMux.HandleFunc("GET /getdetailbyid/{id}", handlers.Canvas.GetDetailById)
	canvasMux.HandleFunc("GET /getallbyuserid/{id}", handlers.Canvas.GetAllByUserId)

	userMux := http.NewServeMux()
	mux.Handle("/user/", http.StripPrefix("/user", userMux))
	userMux.HandleFunc("POST /create", handlers.User.Create)
	userMux.HandleFunc("POST /login", handlers.User.Login)

	return mux
}

// LoggingMiddleware is a standard way to wrap handlers in Go
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

// AuthMiddleware intercepts the request and checks the token
func AuthMiddleware(jwtKey []byte, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Public route
		if r.URL.Path == "/user/login" {
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie("authToken")
		if err != nil {
			// If there is no cookie, they are not logged in
			http.Error(w, "Unauthorized: Missing cookie", http.StatusUnauthorized)
			return
		}

		// Extract the raw JWT string from the cookie
		tokenString := cookie.Value

		// Parse and validate the token using the injected key
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			// Ensure the signing method is what you expect
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, http.ErrAbortHandler
			}
			return jwtKey, nil // Return the key to the parser
		})

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized: Invalid Token", http.StatusUnauthorized)
			return
		}

		// If valid, let the request pass through to the Handler
		next.ServeHTTP(w, r)
	})
}

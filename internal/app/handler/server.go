package handler

import (
	"net/http"
	"toptal/internal/app/handler/middleware"

	httpSwagger "github.com/swaggo/http-swagger"
)

type Server struct {
	router          *http.ServeMux
	bookService     BookService
	categoryService CategoryService
	authService     AuthService
	cartService     CartService
	healthService   HealthService
}

func NewServer(
	bookService BookService,
	categoryService CategoryService,
	authService AuthService,
	cartService CartService,
	healthService HealthService,
) *Server {
	server := &Server{
		router:          http.NewServeMux(),
		bookService:     bookService,
		categoryService: categoryService,
		authService:     authService,
		cartService:     cartService,
		healthService:   healthService,
	}

	server.setupRoutes()
	return server
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) setupRoutes() {
	// Root handler
	s.router.HandleFunc("GET /", s.handleRoot)

	// System routes
	s.router.HandleFunc("GET /health", s.handleHealth)

	// Swagger endpoints
	s.router.HandleFunc("GET /swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))
	s.router.HandleFunc("GET /swagger/doc.json", s.handleSwaggerJSON)

	// Book routes
	s.router.HandleFunc("GET /book/{id}", s.handleGetBookById)
	s.router.HandleFunc("GET /book", s.handleGetBooks)
	s.router.HandleFunc("POST /book", middleware.JWTMiddleware(s.handleCreateBook))
	s.router.HandleFunc("PUT /book/{id}", middleware.JWTMiddleware(s.handleUpdateBook))
	s.router.HandleFunc("DELETE /book/{id}", middleware.JWTMiddleware(s.handleDeleteBook))

	// Category routes
	s.router.HandleFunc("GET /category/{id}", s.handleGetCategoryById)
	s.router.HandleFunc("GET /category", s.handleGetCategories)
	s.router.HandleFunc("POST /category", middleware.JWTMiddleware(s.handleCreateCategory))
	s.router.HandleFunc("PUT /category", middleware.JWTMiddleware(s.handleUpdateCategory))
	s.router.HandleFunc("DELETE /category/{id}", middleware.JWTMiddleware(s.handleDeleteCategory))

	// Cart routes
	s.router.HandleFunc("GET /cart", middleware.JWTMiddleware(s.handleGetCart))
	s.router.HandleFunc("POST /cart/add", middleware.JWTMiddleware(s.handleAddToCart))
	s.router.HandleFunc("POST /cart/remove", middleware.JWTMiddleware(s.handleRemoveFromCart))
	s.router.HandleFunc("POST /cart/purchase", middleware.JWTMiddleware(s.handlePurchase))

	// User routes
	s.router.HandleFunc("POST /login", s.handleLogin)
	s.router.HandleFunc("POST /register", s.handleRegister)
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Book Shop API v1.0"))
}

func (s *Server) handleSwaggerJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	http.ServeFile(w, r, "docs/swagger.json")
}

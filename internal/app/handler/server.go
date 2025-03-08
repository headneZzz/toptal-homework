package handler

import (
	"net/http"
	"toptal/internal/app/auth"

	httpSwagger "github.com/swaggo/http-swagger"
)

// Server представляет собой HTTP-сервер с внедренными сервисами
type Server struct {
	router          *http.ServeMux
	bookService     BookService
	categoryService CategoryService
	userService     UserService
	cartService     CartService
	healthService   HealthService
}

// NewServer создает новый сервер с внедренными сервисами
func NewServer(
	bookService BookService,
	categoryService CategoryService,
	userService UserService,
	cartService CartService,
	healthService HealthService,
) *Server {
	server := &Server{
		router:          http.NewServeMux(),
		bookService:     bookService,
		categoryService: categoryService,
		userService:     userService,
		cartService:     cartService,
		healthService:   healthService,
	}

	server.setupRoutes()
	return server
}

// Handler возвращает обработчик HTTP-запросов
func (s *Server) Handler() http.Handler {
	return s.router
}

// setupRoutes настраивает маршруты сервера
func (s *Server) setupRoutes() {
	// Root handler
	s.router.HandleFunc("GET /", s.handleRoot)

	// System routes
	s.router.HandleFunc("GET /health", s.handleHealth)

	// Swagger endpoints
	s.router.HandleFunc("GET /swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))
	s.router.HandleFunc("GET /swagger/doc.json", s.handleSwaggerJSON)

	// Book routes
	s.router.HandleFunc("GET /book/{id}", s.handleGetBookById)
	s.router.HandleFunc("GET /book", s.handleGetBooks)
	s.router.HandleFunc("POST /book", auth.JWTMiddleware(s.handleCreateBook))
	s.router.HandleFunc("PUT /book/{id}", auth.JWTMiddleware(s.handleUpdateBook))
	s.router.HandleFunc("DELETE /book/{id}", auth.JWTMiddleware(s.handleDeleteBook))

	// Category routes
	s.router.HandleFunc("GET /category/{id}", s.handleGetCategoryById)
	s.router.HandleFunc("GET /category", s.handleGetCategories)
	s.router.HandleFunc("POST /category", auth.JWTMiddleware(s.handleCreateCategory))
	s.router.HandleFunc("PUT /category", auth.JWTMiddleware(s.handleUpdateCategory))
	s.router.HandleFunc("DELETE /category/{id}", auth.JWTMiddleware(s.handleDeleteCategory))

	// Cart routes
	s.router.HandleFunc("GET /cart", auth.JWTMiddleware(s.handleGetCart))
	s.router.HandleFunc("POST /cart/add", auth.JWTMiddleware(s.handleAddToCart))
	s.router.HandleFunc("POST /cart/remove", auth.JWTMiddleware(s.handleRemoveFromCart))
	s.router.HandleFunc("POST /cart/purchase", auth.JWTMiddleware(s.handlePurchase))

	// User routes
	s.router.HandleFunc("POST /login", s.handleLogin)
	s.router.HandleFunc("POST /register", s.handleRegister)
}

// Root handler
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Book Shop API v1.0"))
}

// Swagger handler
func (s *Server) handleSwaggerJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	http.ServeFile(w, r, "docs/swagger.json")
}

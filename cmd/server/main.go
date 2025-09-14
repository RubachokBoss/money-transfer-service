package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi"
	"github.com/joho/godotenv"
	"money-transfer-service/internal/cache"
	"money-transfer-service/internal/handler"
	"money-transfer-service/internal/middleware"
	"money-transfer-service/internal/repository"
	"money-transfer-service/internal/service"
	"money-transfer-service/pkg/postgres"
)

func main() {
	_ = godotenv.Load()

	db, err := postgres.Connect(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	redisClient := cache.NewRedisClient(os.Getenv("REDIS_ADDR"))

	repo := repository.NewRepository(db)
	serv := service.NewService(repo, redisClient)
	h := handler.NewHandler(serv)
	authHandler := handler.NewAuthHandler(repo)

	// Создаем роутер
	r := chi.NewRouter()

	// Middleware логгирования (должно быть первым)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Request: %s %s", r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	})

	// Serve static files
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "static"))
	fileServer := http.FileServer(filesDir)

	// Статические файлы
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// Главная страница
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, filepath.Join(workDir, "static", "index.html"))
			return
		}
		fileServer.ServeHTTP(w, r)
	})

	// Public routes
	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login", authHandler.Login)

	// Protected routes - создаем подроутер с middleware аутентификации
	r.Route("/api", func(r chi.Router) {
		// Middleware аутентификации только для API routes
		r.Use(middleware.AuthMiddleware(repo))

		r.Get("/balance", h.GetBalance)
		r.Post("/transfer", h.TransferMoney)
		r.Post("/deposit", h.DepositMoney)
		r.Get("/transfers", h.GetTransfersHistory)
	})

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

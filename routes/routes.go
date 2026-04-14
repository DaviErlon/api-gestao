package routes

import (
	"log"
	"net/http"

	"github.com/DaviErlon/api-gestao/auth"
	"github.com/DaviErlon/api-gestao/handlers"
	"github.com/DaviErlon/api-gestao/repository"
)

func Run() error {
	repository.Connect()

	mux := http.NewServeMux()

	// ========================
	// 🔓 API PÚBLICA
	// ========================
	mux.HandleFunc("/api/login", auth.LoginHandler)

	// ========================
	// 🔐 API USER
	// ========================
	userRoutes := http.NewServeMux()

	userRoutes.HandleFunc("/users", handlers.ProfileHandler)
	userRoutes.HandleFunc("/users/", handlers.ProfileHandler)

	userRoutes.HandleFunc("/empresas", handlers.EmpresasHandler)
	userRoutes.HandleFunc("/empresas/", handlers.EmpresasHandler)

	userRoutes.HandleFunc("/decisoes", handlers.DecisoesHandler)
	userRoutes.HandleFunc("/decisoes/", handlers.DecisoesHandler)

	userRoutes.HandleFunc("/reunioes", handlers.ReunioesHandler)
	userRoutes.HandleFunc("/reunioes/", handlers.ReunioesHandler)

	userRoutes.HandleFunc("/posts", handlers.PostsHandler)
	userRoutes.HandleFunc("/posts/", handlers.PostsHandler)

	mux.Handle("/api/user/",
		auth.AuthMiddleware(
			http.StripPrefix("/api/user", userRoutes),
		),
	)

	// ========================
	// 🔐 API ADMIN
	// ========================
	adminRoutes := http.NewServeMux()

	adminRoutes.HandleFunc("/users", handlers.AdminProfileHandler)
	adminRoutes.HandleFunc("/empresas", handlers.AdminEmpresasHandler)
	adminRoutes.HandleFunc("/reunioes", handlers.AdminReunioesHandler)
	adminRoutes.HandleFunc("/posts", handlers.AdminPostsHandler)
	adminRoutes.HandleFunc("/ciclos", handlers.CiclosHandler)
	adminRoutes.HandleFunc("/decisoes", handlers.DecisoesHandler)

	mux.Handle("/api/admin/",
		auth.AuthMiddleware(
			auth.AdminMiddleware(
				http.StripPrefix("/api/admin", adminRoutes),
			),
		),
	)

	mux.Handle("/", handlers.StaticsHandler())

	log.Println("Servidor rodando na porta :8080")
	return http.ListenAndServe(":8080", mux)
}
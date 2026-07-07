package main

import (
	"log"

	"github.com/alexroel/gopost-api/config"
	"github.com/alexroel/gopost-api/database"
	"github.com/alexroel/gopost-api/handlers"
	"github.com/alexroel/gopost-api/middleware"
	"github.com/alexroel/gopost-api/repositories"
	"github.com/alexroel/gopost-api/server"
	"github.com/alexroel/gopost-api/services"
)

// health responde al health check en /health. No verifica el estado real
// de la base de datos ni de dependencias externas, solo que el proceso
// atiende peticiones.
func health(c *server.Context) {
	c.Send("Servidor corriendo")
}

// main es la composición raíz de la aplicación: carga config, conecta la
// base de datos y construye la cadena repository -> service -> handler para
// cada recurso antes de registrar las rutas.
func main() {

	config := config.LoadConfig()

	if err := database.Connect(config.DatabaseURL); err != nil {
		log.Fatal("Error al conectar a la base de datos: ", err)
	}

	defer database.Close()

	userRepo := repositories.NewUserRepository(database.DB)
	postRepo := repositories.NewPostRepository(database.DB)

	userService := services.NewUserService(userRepo)
	postService := services.NewPostService(postRepo)

	userHandler := handlers.NewUserHandler(userService)
	postHandler := handlers.NewPostHandler(postService)

	app := server.NewApp()

	app.Get("/health", health)
	app.Post("/signup", userHandler.SignUpHandler)
	app.Post("/login", userHandler.LoginHandler)

	app.Get("/me", middleware.AuthMiddleware(userHandler.MeHandler))

	// Rutas de posts públicas: lectura sin autenticación.
	app.Get("/posts", postHandler.GetPostsHandler)
	app.Get("/posts/{id}", postHandler.GetPostHandler)
	app.Get("/users/{id}/posts", postHandler.GetPostsByUserIDHandler)

	// Rutas de posts protegidas: cada una envuelta individualmente con
	// middleware.AuthMiddleware. No hay middleware global sobre app; la
	// autenticación se decide ruta por ruta aquí, así que cualquier
	// endpoint nuevo que deba requerir login tiene que envolverse
	// explícitamente o quedará público por defecto.
	app.Post("/posts", middleware.AuthMiddleware(postHandler.CreatePostHandler))
	app.Get("/posts/me", middleware.AuthMiddleware(postHandler.GetPostsMeHandler))
	app.Put("/posts/{id}", middleware.AuthMiddleware(postHandler.UpdatePostHandler))
	app.Delete("/posts/{id}", middleware.AuthMiddleware(postHandler.DeletePostHandler))

	if err := app.RunServer(config.Port); err != nil {
		log.Fatal("Error al iniciar el servidor: ", err)
	}
}

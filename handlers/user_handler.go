package handlers

import (
	"net/http"

	"github.com/alexroel/gopost-api/models"
	"github.com/alexroel/gopost-api/server"
	"github.com/alexroel/gopost-api/services"
)

// UserHandler expone los endpoints HTTP de usuarios, delegando toda la
// lógica de negocio en services.UserService.
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler crea un UserHandler sobre userService.
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// SignUpHandler maneja POST /signup: registra un usuario nuevo y devuelve
// sus datos públicos (sin password). No devuelve un token; el cliente debe
// llamar a /login por separado tras registrarse.
func (h *UserHandler) SignUpHandler(c *server.Context) {
	var req models.SignUpInput
	if err := c.BindJSON(&req); err != nil {
		server.RespondError(c, server.NewAppError("Datos inválidos", http.StatusBadRequest))
		return
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		server.RespondError(c, server.NewAppError("Todos los campos son obligatorios", http.StatusBadRequest))
		return
	}

	user, err := h.userService.SignUp(c.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		server.RespondError(c, server.NewAppError(err.Error(), http.StatusBadRequest))
		return
	}

	c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "Usuario creado exitosamente",
		"user": map[string]interface{}{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	})

}

// LoginHandler maneja la solicitud de inicio de sesión de un usuario
func (h *UserHandler) LoginHandler(c *server.Context) {
	var req models.LoginInput

	if err := c.BindJSON(&req); err != nil {
		server.RespondError(c, server.NewAppError("Datos inválidos", http.StatusBadRequest))
		return
	}

	if req.Email == "" || req.Password == "" {
		server.RespondError(c, server.NewAppError("Email y contraseña son obligatorios", http.StatusBadRequest))
		return
	}

	token, err := h.userService.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		server.RespondError(c, server.NewAppError(err.Error(), http.StatusUnauthorized))
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Inicio de sesión exitoso",
		"token":   token,
	})
}

// MeHandler maneja GET /me: devuelve los datos del usuario autenticado.
// Solo se registra detrás de middleware.AuthMiddleware, así que en la
// práctica userID nunca es 0 aquí; el chequeo es una salvaguarda por si
// esta ruta llegara a registrarse sin el middleware por error.
func (h *UserHandler) MeHandler(c *server.Context) {
	userID := c.GetUserID()
	if userID == 0 {
		server.RespondError(c, server.NewAppError("Usuario no autenticado", http.StatusUnauthorized))
		return
	}

	user, err := h.userService.GetUserByID(c.Context(), userID)
	if err != nil {
		server.RespondError(c, server.NewAppError("Usuario no encontrado", http.StatusNotFound))
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	})
}

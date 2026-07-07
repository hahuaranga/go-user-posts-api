package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/alexroel/gopost-api/server"
	"github.com/alexroel/gopost-api/services"
)

// PostHandler expone los endpoints HTTP de posts, delegando toda la lógica
// de negocio (validación, autorización) en services.PostService.
type PostHandler struct {
	postService *services.PostService
}

// NewPostHandler crea un PostHandler sobre postService.
func NewPostHandler(postService *services.PostService) *PostHandler {
	return &PostHandler{postService: postService}
}

// respondPostServiceError traduce un error de PostService al status HTTP
// correcto usando errors.Is contra los sentinela services.ErrPostNotFound /
// services.ErrForbidden, y cae a fallbackCode (normalmente 400) para
// errores de validación sin tipo específico.
func respondPostServiceError(c *server.Context, err error, fallbackCode int) {
	switch {
	case errors.Is(err, services.ErrPostNotFound):
		server.RespondError(c, server.NewAppError(err.Error(), http.StatusNotFound))
	case errors.Is(err, services.ErrForbidden):
		server.RespondError(c, server.NewAppError(err.Error(), http.StatusForbidden))
	default:
		server.RespondError(c, server.NewAppError(err.Error(), fallbackCode))
	}
}

// CreatePostHandler maneja POST /posts (ruta protegida): crea un post para
// el usuario autenticado, tomando userID del contexto (fijado por
// AuthMiddleware), no del body de la petición.
func (h *PostHandler) CreatePostHandler(c *server.Context) {
	userID := c.GetUserID()

	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := c.BindJSON(&req); err != nil {
		server.RespondError(c, server.NewAppError("Datos inválidos", http.StatusBadRequest))
		return
	}

	if req.Title == "" || req.Content == "" {
		server.RespondError(c, server.NewAppError("El título y contenido son obligatorios", http.StatusBadRequest))
		return
	}

	post, err := h.postService.CreatePost(c.Context(), userID, req.Title, req.Content)
	if err != nil {
		server.RespondError(c, server.NewAppError(err.Error(), http.StatusBadRequest))
		return
	}

	c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "Post creado exitosamente",
		"post":    post,
	})
}

// GetPostsHandler maneja GET /posts (ruta pública): lista todos los posts
// de todos los usuarios, sin paginar.
func (h *PostHandler) GetPostsHandler(c *server.Context) {
	posts, err := h.postService.GetAllPosts(c.Context())
	if err != nil {
		log.Printf("GetPostsHandler: %v", err)
		server.RespondError(c, server.NewAppError("Error al obtener los posts", http.StatusInternalServerError))
		return
	}

	c.JSON(http.StatusOK, posts)
}

// GetPostHandler maneja GET /posts/{id} (ruta pública).
func (h *PostHandler) GetPostHandler(c *server.Context) {
	idStr := c.Request.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		server.RespondError(c, server.NewAppError("ID inválido", http.StatusBadRequest))
		return
	}

	post, err := h.postService.GetPostByID(c.Context(), uint(id))
	if err != nil {
		respondPostServiceError(c, err, http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, post)
}

// GetPostsMeHandler maneja GET /posts/me (ruta protegida): lista los posts
// del usuario autenticado, tomando userID del contexto. Reutiliza el mismo
// PostService.GetPostsByUserID que GetPostsByUserIDHandler, pero aquí el ID
// viene del JWT y no de la URL.
func (h *PostHandler) GetPostsMeHandler(c *server.Context) {
	userID := c.GetUserID()

	posts, err := h.postService.GetPostsByUserID(c.Context(), userID)
	if err != nil {
		log.Printf("GetPostsMeHandler: %v", err)
		server.RespondError(c, server.NewAppError("Error al obtener tus posts", http.StatusInternalServerError))
		return
	}

	c.JSON(http.StatusOK, posts)
}

// GetPostsByUserIDHandler maneja GET /users/{id}/posts (ruta pública):
// lista los posts de cualquier usuario indicado en la URL, sin requerir
// autenticación ni verificar que exista ese usuario (un id inexistente
// simplemente devuelve una lista vacía).
func (h *PostHandler) GetPostsByUserIDHandler(c *server.Context) {
	idStr := c.Request.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		server.RespondError(c, server.NewAppError("ID inválido", http.StatusBadRequest))
		return
	}

	posts, err := h.postService.GetPostsByUserID(c.Context(), uint(id))
	if err != nil {
		log.Printf("GetPostsByUserIDHandler: %v", err)
		server.RespondError(c, server.NewAppError("Error al obtener los posts del usuario", http.StatusInternalServerError))
		return
	}

	c.JSON(http.StatusOK, posts)
}

// UpdatePostHandler maneja PUT /posts/{id} (ruta protegida). Los errores de
// validación responden 400; "post no encontrado" y "sin permiso" responden
// 404/403 vía respondPostServiceError.
func (h *PostHandler) UpdatePostHandler(c *server.Context) {
	idStr := c.Request.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		server.RespondError(c, server.NewAppError("ID inválido", http.StatusBadRequest))
		return
	}

	userID := c.GetUserID()

	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := c.BindJSON(&req); err != nil {
		server.RespondError(c, server.NewAppError("Datos inválidos", http.StatusBadRequest))
		return
	}

	if req.Title == "" || req.Content == "" {
		server.RespondError(c, server.NewAppError("El título y contenido son obligatorios", http.StatusBadRequest))
		return
	}

	if err := h.postService.UpdatePost(c.Context(), uint(id), userID, req.Title, req.Content); err != nil {
		respondPostServiceError(c, err, http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, map[string]string{
		"message": "Post actualizado exitosamente",
	})
}

// DeletePostHandler maneja DELETE /posts/{id} (ruta protegida). Mismo
// mapeo de errores que UpdatePostHandler vía respondPostServiceError.
func (h *PostHandler) DeletePostHandler(c *server.Context) {
	idStr := c.Request.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		server.RespondError(c, server.NewAppError("ID inválido", http.StatusBadRequest))
		return
	}

	userID := c.GetUserID()

	if err := h.postService.DeletePost(c.Context(), uint(id), userID); err != nil {
		respondPostServiceError(c, err, http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, map[string]string{
		"message": "Post eliminado exitosamente",
	})
}

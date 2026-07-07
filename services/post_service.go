package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/alexroel/gopost-api/models"
	"github.com/alexroel/gopost-api/repositories"
)

// Errores sentinela que permiten a los handlers distinguir, con errors.Is,
// entre "no encontrado" y "sin permiso" para mapearlos a 404/403 en vez de
// tratar todo error de PostService como un simple 400.
var (
	ErrPostNotFound = errors.New("post no encontrado")
	ErrForbidden    = errors.New("no tienes permiso para esta acción")
)

// PostService contiene las reglas de negocio de posts: validación de campos
// y verificación de que solo el autor de un post pueda modificarlo o
// eliminarlo. Las queries SQL viven solo en repositories.PostRepository.
type PostService struct {
	repo *repositories.PostRepository
}

// NewPostService crea un PostService sobre repo.
func NewPostService(repo *repositories.PostRepository) *PostService {
	return &PostService{repo: repo}
}

// CreatePost valida título y contenido y crea un post asociado a userID.
func (s *PostService) CreatePost(ctx context.Context, userID uint, title, content string) (*models.Post, error) {
	if title == "" {
		return nil, fmt.Errorf("el título es requerido")
	}

	if content == "" {
		return nil, fmt.Errorf("el contenido es requerido")
	}

	post := &models.Post{
		UserID:  userID,
		Title:   title,
		Content: content,
	}

	if err := s.repo.Create(ctx, post); err != nil {
		return nil, fmt.Errorf("error al crear el post: %w", err)
	}

	return post, nil
}

// GetAllPosts devuelve todos los posts existentes, sin filtrar por autor.
func (s *PostService) GetAllPosts(ctx context.Context) ([]models.Post, error) {
	posts, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("error al obtener los posts: %w", err)
	}

	return posts, nil
}

// GetPostByID obtiene un post por su ID. No exige autenticación: el
// endpoint público GET /posts/{id} lo usa directamente.
func (s *PostService) GetPostByID(ctx context.Context, id uint) (*models.Post, error) {
	post, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrPostNotFound
	}

	return post, nil
}

// GetPostsByUserID obtiene todos los posts de userID. Se reutiliza tanto
// para GET /posts/me (userID sacado del JWT) como para GET /users/{id}/posts
// (userID de la URL, ruta pública): la propia firma no distingue "mis
// posts" de "los posts de un tercero", esa decisión la toma el handler que
// llama a este método.
func (s *PostService) GetPostsByUserID(ctx context.Context, userID uint) ([]models.Post, error) {
	posts, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error al obtener los posts del usuario: %w", err)
	}

	return posts, nil
}

// UpdatePost valida título y contenido, confirma que postID pertenece a
// userID (autorización a nivel de servicio, no de middleware) y persiste
// los cambios.
//
// Los errores de "no encontrado" y "sin permiso" envuelven ErrPostNotFound
// y ErrForbidden respectivamente, para que el handler los distinga con
// errors.Is y responda 404/403 en vez de un genérico 400.
func (s *PostService) UpdatePost(ctx context.Context, postID, userID uint, title, content string) error {
	if title == "" {
		return fmt.Errorf("el título es requerido")
	}

	if content == "" {
		return fmt.Errorf("el contenido es requerido")
	}

	existingPost, err := s.repo.FindByID(ctx, postID)
	if err != nil {
		return ErrPostNotFound
	}

	if existingPost.UserID != userID {
		return fmt.Errorf("no tienes permiso para actualizar este post: %w", ErrForbidden)
	}

	existingPost.Title = title
	existingPost.Content = content

	if err := s.repo.Update(ctx, existingPost); err != nil {
		return fmt.Errorf("error al actualizar el post: %w", err)
	}

	return nil
}

// DeletePost confirma que postID pertenece a userID (misma verificación de
// autorización y mismos errores sentinela que UpdatePost) y lo elimina.
func (s *PostService) DeletePost(ctx context.Context, postID, userID uint) error {
	existingPost, err := s.repo.FindByID(ctx, postID)
	if err != nil {
		return ErrPostNotFound
	}

	if existingPost.UserID != userID {
		return fmt.Errorf("no tienes permiso para eliminar este post: %w", ErrForbidden)
	}

	if err := s.repo.Delete(ctx, postID); err != nil {
		return fmt.Errorf("error al eliminar el post: %w", err)
	}

	return nil
}

package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/alexroel/gopost-api/models"
)

// PostRepository encapsula el acceso a la tabla posts. No valida datos de
// negocio ni permisos (título/contenido vacíos, propiedad del post): eso es
// responsabilidad de services.PostService.
type PostRepository struct {
	db *sql.DB
}

// NewPostRepository crea un PostRepository sobre un *sql.DB ya conectado
// (ver database.Connect).
func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{db: db}
}

// Create inserta post y rellena post.ID con el ID autogenerado por MySQL.
func (r *PostRepository) Create(ctx context.Context, post *models.Post) error {
	query := "INSERT INTO posts (user_id, title, content) VALUES (?, ?, ?)"
	result, err := r.db.ExecContext(ctx, query, post.UserID, post.Title, post.Content)
	if err != nil {
		return fmt.Errorf("error al crear post: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error al obtener ID: %w", err)
	}

	post.ID = uint(id)
	return nil
}

// FindAll devuelve todos los posts, más recientes primero. Sin paginación:
// en una tabla grande esto carga el resultado completo en memoria.
func (r *PostRepository) FindAll(ctx context.Context) ([]models.Post, error) {
	query := "SELECT id, user_id, title, content, created_at, updated_at FROM posts ORDER BY created_at DESC"
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error al obtener posts: %w", err)
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		if err := rows.Scan(&post.ID, &post.UserID, &post.Title,
			&post.Content, &post.CreatedAt, &post.UpdatedAt); err != nil {
			return nil, fmt.Errorf("error al escanear post: %w", err)
		}
		posts = append(posts, post)
	}

	return posts, nil
}

// FindByID busca un post por su ID.
func (r *PostRepository) FindByID(ctx context.Context, id uint) (*models.Post, error) {
	post := &models.Post{}
	query := "SELECT id, user_id, title, content, created_at, updated_at FROM posts WHERE id = ?"

	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("post no encontrado")
		}
		return nil, fmt.Errorf("error al buscar post: %w", err)
	}

	return post, nil
}

// FindByUserID devuelve los posts de userID, más recientes primero. Igual
// que FindAll, no pagina el resultado.
func (r *PostRepository) FindByUserID(ctx context.Context, userID uint) ([]models.Post, error) {
	query := "SELECT id, user_id, title, content, created_at, updated_at FROM posts WHERE user_id = ? ORDER BY created_at DESC"
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("error al obtener posts del usuario: %w", err)
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		if err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Content,
			&post.CreatedAt, &post.UpdatedAt); err != nil {
			return nil, fmt.Errorf("error al escanear post: %w", err)
		}
		posts = append(posts, post)
	}

	return posts, nil
}

// Update sobrescribe título y contenido de post.ID.
//
// GOTCHA: MySQL, por defecto (sin el flag de protocolo CLIENT_FOUND_ROWS),
// reporta en RowsAffected las filas realmente *modificadas*, no las filas
// que hicieron match en el WHERE. Si un cliente reenvía el mismo título y
// contenido que ya tenía el post, MySQL devuelve 0 filas afectadas aunque
// el post exista y el UPDATE haya sido válido, y este método responderá
// erróneamente "post no encontrado". El repositorio no puede distinguir ese
// caso de un ID inexistente sin una consulta adicional (o sin habilitar
// CLIENT_FOUND_ROWS en el DSN).
func (r *PostRepository) Update(ctx context.Context, post *models.Post) error {
	query := "UPDATE posts SET title = ?, content = ? WHERE id = ?"
	result, err := r.db.ExecContext(ctx, query, post.Title, post.Content, post.ID)
	if err != nil {
		return fmt.Errorf("error al actualizar post: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error al verificar actualización: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("post no encontrado")
	}

	return nil
}

// Delete elimina el post con el id dado. A diferencia de Update, aquí
// RowsAffected sí refleja de forma fiable si la fila existía: un DELETE no
// tiene la ambigüedad "afectada vs. coincidente" propia del UPDATE.
func (r *PostRepository) Delete(ctx context.Context, id uint) error {
	query := "DELETE FROM posts WHERE id = ?"
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error al eliminar post: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error al verificar eliminación: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("post no encontrado")
	}

	return nil
}

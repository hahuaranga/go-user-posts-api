package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/alexroel/gopost-api/models"
)

// UserRepository encapsula el acceso a la tabla users. No valida datos de
// negocio (formato de email, longitud de password, etc.): eso es
// responsabilidad de services.UserService.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository crea un UserRepository sobre un *sql.DB ya conectado
// (ver database.Connect).
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserta user y rellena user.ID con el ID autogenerado por MySQL.
// Asume que user.Password ya llega hasheado; este método no hashea nada.
func (r *UserRepository) Create(cxt context.Context, user *models.User) error {
	query := "INSERT INTO users (name, email, password) VALUES (?, ?, ?)"
	result, err := r.db.ExecContext(cxt, query, user.Name, user.Email, user.Password)
	if err != nil {
		return fmt.Errorf("error al crear usuario: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error al obtener el ID del usuario creado: %w", err)
	}

	user.ID = uint(id)
	return nil
}

// FindByID busca un usuario por ID. Deliberadamente no selecciona la
// columna password (a diferencia de FindByEmail): este método alimenta
// respuestas como /me, que nunca deben tener el hash en memoria.
func (r *UserRepository) FindByID(cxt context.Context, id uint) (*models.User, error) {
	user := &models.User{}
	query := "SELECT id, name, email FROM users WHERE id = ?"
	err := r.db.QueryRowContext(cxt, query, id).Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("usuario no encontrado")
		}

		return nil, fmt.Errorf("error al buscar usuario por ID: %w", err)
	}

	return user, nil
}

// FindByEmail busca un usuario por email, incluyendo el hash de password:
// es el único método pensado para el flujo de login
// (services.UserService.Login), donde se necesita para comparar con bcrypt.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	query := "SELECT id, name, email, password FROM users WHERE email = ?"

	err := r.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Name, &user.Email, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("usuario no encontrado")
		}
		return nil, fmt.Errorf("error al buscar usuario: %w", err)
	}

	return user, nil
}

// EmailExists indica si ya existe un usuario con ese email. Se usa antes de
// Create para devolver un error de negocio claro en vez de depender del
// UNIQUE constraint de la tabla y tener que parsear el error del driver.
func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM users WHERE email = ?"

	err := r.db.QueryRowContext(ctx, query, email).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("error al verificar email: %w", err)
	}

	return count > 0, nil
}

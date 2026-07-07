package services

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/alexroel/gopost-api/config"
	"github.com/alexroel/gopost-api/models"
	"github.com/alexroel/gopost-api/repositories"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// UserService contiene las reglas de negocio de usuarios: validación,
// hashing de contraseñas y emisión de JWT. Las queries SQL viven solo en
// repositories.UserRepository.
type UserService struct {
	repo *repositories.UserRepository
}

// NewUserService crea un UserService sobre repo.
func NewUserService(repo *repositories.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// ValidateEmail valida que email tenga un formato sintácticamente correcto.
// No verifica que el dominio exista ni que la casilla sea entregable.
func ValidateEmail(email string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("formato de email inválido")
	}

	return nil
}

// ValidatePassword valida que la contraseña tenga al menos 6 caracteres
func ValidatePassword(password string) error {
	if len(password) < 6 {
		return fmt.Errorf("la contraseña debe tener al menos 6 caracteres")
	}
	return nil
}

// SignUp registra un nuevo usuario: valida email y contraseña, comprueba
// que el email no esté en uso, hashea la contraseña con bcrypt y persiste
// el usuario. El *models.User devuelto trae Password ya hasheado, nunca la
// contraseña original.
func (s *UserService) SignUp(ctx context.Context, name, email, password string) (*models.User, error) {
	if err := ValidateEmail(email); err != nil {
		return nil, err
	}

	if err := ValidatePassword(password); err != nil {
		return nil, err
	}

	exists, err := s.repo.EmailExists(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("el email ya está registrado")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error al hashear la contraseña: %w", err)
	}

	user := &models.User{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// generateToken firma un JWT HS256 con el claim user_id y expiración a 72h.
//
// Depende del singleton config.AppConfig, que debe estar poblado por
// config.LoadConfig antes de la primera llamada; si AppConfig es nil esto
// entra en panic por nil pointer dereference (no hay chequeo defensivo).
func (s *UserService) generateToken(userId uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userId,
		"exp":     jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppConfig.JWTSecret))
}

// Login valida credenciales contra la contraseña hasheada almacenada y, si
// coinciden, devuelve un JWT firmado. Los dos casos de fallo (email
// inexistente y contraseña incorrecta) devuelven el mismo mensaje genérico
// a propósito, para no revelar a un atacante si un email está registrado.
func (s *UserService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("credenciales incorrectas")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", fmt.Errorf("credenciales incorrectas")
	}

	toke, err := s.generateToken(user.ID)
	if err != nil {
		return "", fmt.Errorf("error al generar el token: %w", err)
	}

	return toke, nil
}

// GetUserByID obtiene un usuario por su ID, delegando directamente en el
// repositorio (sin lógica de negocio adicional).
func (s *UserService) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	return s.repo.FindByID(ctx, id)
}

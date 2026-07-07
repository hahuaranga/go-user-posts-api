package models

// User representa un usuario registrado. Password guarda el hash bcrypt,
// nunca la contraseña en texto plano, y su tag `json:"-"` impide que se
// serialice aunque el struct completo se pase a Context.JSON por error.
type User struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

// SignUpInput es el cuerpo esperado en POST /signup.
type SignUpInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginInput es el cuerpo esperado en POST /login.
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

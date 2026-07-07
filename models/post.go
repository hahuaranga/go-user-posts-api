package models

// Post representa una publicación de un usuario.
//
// CreatedAt/UpdatedAt son string, no time.Time, porque el driver MySQL
// devuelve las columnas TIMESTAMP como texto mientras el DSN no incluya
// parseTime=true (ver database/database.go). Si en el futuro se agrega ese
// parámetro al DSN, el driver empezará a entregar time.Time y el Scan de
// estos campos en los repositorios fallará en tiempo de ejecución: hay que
// cambiar el tipo aquí y en los repositorios a la vez.
type Post struct {
	ID        uint   `json:"id"`
	UserID    uint   `json:"user_id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

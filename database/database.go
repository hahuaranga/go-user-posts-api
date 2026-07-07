package database

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// DB es el pool de conexiones global a MySQL, usado directamente por todos
// los repositorios. Permanece en nil hasta que Connect se ejecuta con éxito;
// no es seguro usarlo antes de eso.
var DB *sql.DB

// Connect abre el pool de conexiones a MySQL y verifica con Ping que la
// base de datos responde antes de devolver el control a main.
//
// sql.Open no abre ninguna conexión real por sí solo (solo valida el DSN);
// el Ping explícito es lo que detecta credenciales u host inválidos al
// arrancar, en vez de que el primer error aparezca recién en la primera
// petición HTTP.
func Connect(dsn string) error {
	var err error

	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("error al abrir la conexión: %w", err)
	}

	err = DB.Ping()
	if err != nil {
		return fmt.Errorf("error al conectar a la base de datos: %w", err)
	}

	// Límites del pool para no agotar max_connections de MySQL; mantener
	// conexiones inactivas listas evita el costo de reabrir TCP+auth en
	// cada request.
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(10)

	return nil
}

// Close libera el pool de conexiones. Es seguro llamarlo aunque Connect
// nunca se haya ejecutado (DB seguiría siendo nil).
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

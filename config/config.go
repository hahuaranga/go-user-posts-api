package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config agrupa los valores de configuración de la aplicación,
// cargados desde variables de entorno (ver LoadConfig).
type Config struct {
	Port        string
	JWTSecret   string
	DatabaseURL string
}

// AppConfig es la instancia global de configuración, poblada por LoadConfig.
// Es nil hasta que LoadConfig se ejecuta; otros paquetes (middleware, services)
// la leen directamente, así que LoadConfig debe llamarse una única vez al
// inicio de main antes de construir cualquier componente que dependa de ella.
var AppConfig *Config

// LoadConfig lee el archivo .env (si existe) y las variables de entorno,
// y devuelve la configuración resultante en AppConfig.
//
// Termina el proceso con log.Fatal si JWT_SECRET no está definido, ya que
// no existe un valor por defecto seguro para firmar tokens. Por esta razón
// no debe invocarse desde tests unitarios sin definir antes esa variable.
func LoadConfig() *Config {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("No se encontró archivo .env")
	}

	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET es requerido. Por favor configura esta variable de entorno.")
	}

	AppConfig = &Config{
		Port:        getEnv("PORT", ":5050"),
		JWTSecret:   jwtSecret,
		DatabaseURL: getEnv("DATABASE_URL", "root:password@tcp(localhost:3306)/gopost"),
	}
	return AppConfig
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

package middleware

import (
	"net/http"
	"strings"

	"github.com/alexroel/gopost-api/config"
	"github.com/alexroel/gopost-api/server"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware decora next exigiendo un JWT válido en el header
// "Authorization: Bearer <token>". Si la validación pasa, extrae el claim
// user_id y lo deja disponible para next vía c.SetUserID; si falla en
// cualquier paso, corta la cadena respondiendo 401 y sin invocar next.
//
// Se aplica ruta por ruta en main.go (no como middleware global de App),
// para poder dejar públicos los GET de posts y proteger solo las mutaciones.
func AuthMiddleware(next server.HandleFunc) server.HandleFunc {
	return func(c *server.Context) {
		authHeader := c.Request.Header.Get("Authorization")

		if authHeader == "" {
			server.RespondError(c, server.NewAppError("Token no proporcionado", http.StatusUnauthorized))
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			server.RespondError(c, server.NewAppError("Formato de token inválido", http.StatusUnauthorized))
			return
		}

		tokenString := parts[1]

		// La comprobación de SigningMethodHMAC es la defensa contra el ataque
		// de "confusión de algoritmo": sin ella, un token construido con
		// alg=none o con un algoritmo asimétrico (RS256 usando la clave
		// pública como si fuera el secreto HMAC) podría colarse como válido.
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, server.NewAppError("Método de firma inesperado", http.StatusUnauthorized)
			}
			return []byte(config.AppConfig.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			server.RespondError(c, server.NewAppError("Token inválido", http.StatusUnauthorized))
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			server.RespondError(c, server.NewAppError("Claims inválidos", http.StatusUnauthorized))
			return
		}

		// jwt.MapClaims decodifica el JSON del token con encoding/json
		// estándar, que representa todo número como float64 (no existe int
		// en JSON). Por eso el type assertion es a float64 y no a uint o int:
		// generateToken codifica user_id como número, así que aquí siempre
		// vuelve como float64, nunca como un entero de Go.
		userID, ok := claims["user_id"].(float64)
		if !ok {
			server.RespondError(c, server.NewAppError("User ID no encontrado en el token", http.StatusUnauthorized))
			return
		}

		c.SetUserID(uint(userID))
		next(c)

	}
}

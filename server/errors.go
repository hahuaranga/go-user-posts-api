package server

import "net/http"

// ErrorResponse es el formato JSON uniforme para toda respuesta de error
// de la API.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// AppError adjunta un código de estado HTTP a un mensaje de error, para que
// RespondError sepa con qué status responder. Handlers y middleware la usan
// como wrapper en el borde HTTP; los errores de dominio (services) siguen
// siendo errors/fmt.Errorf planos y cada llamador decide a qué AppError
// mapearlos.
//
// Vive en server (no en handlers) para que middleware pueda construir
// respuestas de error sin depender del paquete handlers.
type AppError struct {
	Message string
	Code    int
}

// Error implementa la interfaz error.
func (e *AppError) Error() string {
	return e.Message
}

// NewAppError construye un AppError con el mensaje y código HTTP dados.
func NewAppError(message string, code int) *AppError {
	return &AppError{
		Message: message,
		Code:    code,
	}
}

// RespondError escribe appErr como ErrorResponse en el código de estado
// indicado. El campo Error se deriva de http.StatusText(appErr.Code), así
// que un Code que no sea un status HTTP estándar producirá un Error vacío.
func RespondError(c *Context, appErr *AppError) {
	c.JSON(appErr.Code, ErrorResponse{
		Error:   http.StatusText(appErr.Code),
		Message: appErr.Message,
		Code:    appErr.Code,
	})
}

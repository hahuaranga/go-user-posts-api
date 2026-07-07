package server

import (
	"context"
	"encoding/json"
	"net/http"
)

// Context envuelve una petición HTTP individual y se pasa a cada HandleFunc
// registrado en App. userID queda deliberadamente sin exportar: solo se
// puede escribir mediante SetUserID (llamado por AuthMiddleware tras validar
// el JWT), evitando que un handler la fije por error a partir de datos no
// verificados.
type Context struct {
	RWriter http.ResponseWriter
	Request *http.Request
	Cxt     context.Context
	userID  uint
}

// Send escribe text tal cual en el cuerpo de la respuesta, sin fijar
// Content-Type ni código de estado (usa el 200 por defecto de net/http).
func (c *Context) Send(text string) {
	c.RWriter.Write([]byte(text))
}

// Status escribe únicamente el código de estado HTTP. Como con cualquier
// http.ResponseWriter, debe llamarse antes de escribir el cuerpo: una vez
// enviado un byte, net/http fija el código en 200 de forma implícita y ya
// no puede cambiarse.
func (c *Context) Status(code int) {
	c.RWriter.WriteHeader(code)
}

// JSON serializa data como JSON y lo escribe en la respuesta con el código
// indicado.
//
// El código de estado se escribe antes de codificar el body, así que si
// json.Encode falla a mitad de la escritura (p. ej. un tipo no serializable)
// el cliente ya recibió el header con code y no hay forma de corregirlo:
// el error devuelto solo sirve para logging, no para reintentar la respuesta.
func (c *Context) JSON(code int, data interface{}) error {
	c.RWriter.Header().Set("Content-Type", "application/json")
	c.RWriter.WriteHeader(code)
	return json.NewEncoder(c.RWriter).Encode(data)
}

// BindJSON decodifica el cuerpo JSON de la petición en dest (debe ser un
// puntero). No cierra el body; net/http lo hace automáticamente al finalizar
// el handler.
func (c *Context) BindJSON(dest interface{}) error {
	return json.NewDecoder(c.Request.Body).Decode(dest)
}

// SetUserID asocia el ID del usuario autenticado a esta petición. Pensado
// para ser invocado únicamente por AuthMiddleware tras validar el JWT.
func (c *Context) SetUserID(id uint) {
	c.userID = id
}

// GetUserID devuelve el ID del usuario autenticado, o 0 si SetUserID nunca
// se llamó (rutas públicas, sin AuthMiddleware). Los handlers protegidos no
// necesitan validar este caso porque siempre están envueltos en
// AuthMiddleware, que ya rechaza la petición sin un token válido.
func (c *Context) GetUserID() uint {
	return c.userID
}

// Context devuelve el context.Context de la petición HTTP subyacente, para
// propagar cancelación/deadline a llamadas a la base de datos u otras
// operaciones sensibles al contexto.
func (c *Context) Context() context.Context {
	return c.Cxt
}

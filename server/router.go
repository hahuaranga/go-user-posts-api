package server

import "net/http"

// HandleFunc es la firma que deben cumplir los handlers registrados en App.
// Existe como alias del literal func(*Context) para que middleware como
// AuthMiddleware pueda expresar su firma "decorator" (HandleFunc -> HandleFunc)
// sin repetir el tipo función completo.
type HandleFunc func(c *Context)

// Get registra handler para peticiones GET en path. path sigue la sintaxis
// de patrones de http.ServeMux (Go 1.22+), incluyendo comodines de segmento
// como "/posts/{id}" recuperables luego con Request.PathValue("id").
func (app *App) Get(path string, handler func(*Context)) {
	app.mux.HandleFunc("GET "+path, func(w http.ResponseWriter, r *http.Request) {
		handler(&Context{
			RWriter: w,
			Request: r,
			Cxt:     r.Context(),
		})
	})

	app.handlerCount++
}

// Post registra handler para peticiones POST en path. Ver Get para la
// sintaxis de path.
func (app *App) Post(path string, handler func(*Context)) {
	app.mux.HandleFunc("POST "+path, func(w http.ResponseWriter, r *http.Request) {
		handler(&Context{
			RWriter: w,
			Request: r,
			Cxt:     r.Context(),
		})
	})

	app.handlerCount++
}

// Put registra handler para peticiones PUT en path. Ver Get para la
// sintaxis de path.
func (app *App) Put(path string, handler func(*Context)) {
	app.mux.HandleFunc("PUT "+path, func(w http.ResponseWriter, r *http.Request) {
		handler(&Context{
			RWriter: w,
			Request: r,
			Cxt:     r.Context(),
		})
	})

	app.handlerCount++
}

// Delete registra handler para peticiones DELETE en path. Ver Get para la
// sintaxis de path.
func (app *App) Delete(path string, handler func(*Context)) {
	app.mux.HandleFunc("DELETE "+path, func(w http.ResponseWriter, r *http.Request) {
		handler(&Context{
			RWriter: w,
			Request: r,
			Cxt:     r.Context(),
		})
	})

	app.handlerCount++
}

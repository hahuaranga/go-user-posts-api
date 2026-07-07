package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// App es un envoltorio mínimo sobre http.ServeMux que registra el número
// de rutas dadas de alta, solo para mostrarlo en el banner de arranque.
type App struct {
	mux          *http.ServeMux
	handlerCount int
}

// NewApp crea una App lista para registrar rutas con Get/Post/Put/Delete
// (ver router.go).
func NewApp() *App {
	return &App{
		mux:          http.NewServeMux(),
		handlerCount: 0,
	}
}

// RunServer inicia el servidor HTTP en port y bloquea hasta que reciba
// SIGINT/SIGTERM, momento en el que intenta un apagado ordenado (hasta 10s
// para drenar peticiones en curso) y retorna. Esto permite que el defer
// database.Close() de main se ejecute en el camino normal de apagado, no
// solo cuando ListenAndServe falla al arrancar.
func (app *App) RunServer(port string) error {

	app.printBanner(port)

	httpServer := &http.Server{
		Addr:         port,
		Handler:      app.mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serveErr := make(chan error, 1)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serveErr <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serveErr:
		return err
	case <-stop:
		log.Println("Apagando servidor...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return httpServer.Shutdown(ctx)
	}
}

// printBanner imprime en consola la URL base y el número de rutas
// registradas, a modo de confirmación visual de que el servidor arrancó.
func (app *App) printBanner(port string) {

	urlBase := fmt.Sprintf("http://localhost%s", port)
	handlerCount := fmt.Sprintf("Handlers ......: %d", app.handlerCount)

	fmt.Println("┌───────────────────────────────────────────────────┐")
	fmt.Printf("|%s|\n", textCenter("MyServer v1.0.0", 51))
	fmt.Printf("|%s|\n", textCenter(urlBase, 51))
	fmt.Printf("|%s|\n", strings.Repeat(" ", 51))
	fmt.Printf("|%s|\n", textCenter(handlerCount, 51))
	fmt.Println("└───────────────────────────────────────────────────┘")

}

// textCenter centra text dentro de un campo de width caracteres, o lo trunca
// si ya lo excede. La división entera del padding puede dejar un espacio de
// más a la derecha en anchos impares; es solo estético, sin efecto funcional.
func textCenter(text string, width int) string {
	if len(text) >= width {
		return text[:width]
	}

	padding := (width - len(text)) / 2
	return strings.Repeat(" ", padding) + text + strings.Repeat(" ", width-len(text)-padding)
}

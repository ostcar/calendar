package web

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ostcar/calendar/model"
	"github.com/ostcar/calendar/web/template"
)

//go:generate templ generate -path template

//go:embed files
var publicFiles embed.FS

// Run starts the server.
func Run(ctx context.Context, addr string, model *model.Model) error {
	handler := newServer(model)

	httpSRV := &http.Server{
		Addr:        addr,
		Handler:     handler,
		BaseContext: func(net.Listener) context.Context { return ctx },
	}

	// Shutdown logic in separate goroutine.
	wait := make(chan error)
	go func() {
		// Wait for the context to be closed.
		<-ctx.Done()

		if err := httpSRV.Shutdown(context.WithoutCancel(ctx)); err != nil {
			wait <- fmt.Errorf("HTTP server shutdown: %w", err)
			return
		}
		wait <- nil
	}()

	fmt.Printf("Listen webserver on: %s\n", addr)
	if err := httpSRV.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("HTTP Server failed: %v", err)
	}

	return <-wait
}

type server struct {
	http.Handler
	model *model.Model
}

func newServer(model *model.Model) server {
	srv := server{
		model: model,
	}
	srv.registerHandlers()

	return srv
}

func (s *server) registerHandlers() {
	router := mux.NewRouter()

	router.PathPrefix("/assets").Handler(handleStatic())

	router.Handle("/", handleError(s.handleHome))

	s.Handler = router
}

func (s server) handleHome(w http.ResponseWriter, r *http.Request) error {
	month := s.model.ThisMonth()
	if attr := r.URL.Query().Get("month"); attr != "" {
		var err error
		month, err = s.model.MonthFromAttr(attr)
		if err != nil {
			return err
		}
	}
	return template.Month(month).Render(r.Context(), w)
}

func handleStatic() http.Handler {
	files, err := fs.Sub(publicFiles, "files")
	if err != nil {
		// This only happens on startup time.
		panic(err)
	}

	return http.FileServer(http.FS(files))
}

func handleError(handler func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			log.Printf("Error: %v", err)
			if r.Header.Get("HX-Request") != "" {
				// TODO
				return
			}

			http.Error(w, "Ups, something went wrong", 500)
		}
	}
}

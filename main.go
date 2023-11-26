package main

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Handler func(w http.ResponseWriter, r *http.Request) error

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {
		// handle returned error here.
		w.WriteHeader(503)
		w.Write([]byte("bad"))
	}
}

func main() {
	r := chi.NewRouter()
	//--
	// This line adds the RequestID middleware to your router. The RequestID middleware generates a unique ID for each HTTP request. This is useful for logging and tracing requests through your system. If an ID is already present in the request header, it will use that, otherwise, it will generate a new one.
	r.Use(middleware.RequestID)
	//--
	//Here, the Logger middleware is added to the router. This middleware logs the start and end of each request with the elapsed processing time, status code, and similar request details. It's useful for monitoring and debugging the behavior of your web application by providing insights into the traffic it's handling.
	r.Use(middleware.Logger)
	//--
	//This middleware recovers from panics anywhere in the chain, prevents the panic from crashing the server, and logs the panic. This is a safety feature to ensure that if your application encounters an unexpected error during request processing, it can recover gracefully without crashing.
	r.Use(middleware.Recoverer)
	// --
	// w (of type http.ResponseWriter): This is used to write the response that will be sent back to the client. The ResponseWriter interface is used to send HTTP responses.
	// r (of type *http.Request): This represents the HTTP request received by the server. It contains details like the request URL, headers, query parameters, etc.
	// w.Write([]byte("hello world"))

	// Inside the handler function, the Write method of ResponseWriter (w) is called to send a response back to the client.
	// The argument to Write is []byte("hello world"), which converts the string "hello world" to a byte slice ([]byte). In Go, HTTP responses are sent as byte slices.
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world"))
	})

	// Example of customHandler being used when a user hits the /picture endpoint.
	r.Method("GET", "/picture", Handler(customHandler))

	// Create a route along /files that will serve contents from
	// the ./data/ folder.
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "data"))
	FileServer(r, "/files", filesDir)

	http.ListenAndServe(":3333", r)
}

// Example of a custom handler function.
func customHandler(w http.ResponseWriter, r *http.Request) error {
	q := r.URL.Query().Get("err")

	if q != "" {
		return errors.New(q)
	}

	w.Write([]byte("A whole bunch of messages and such"))
	return nil
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}

//Notes:
//1. You can create your own custom HTTP Methods(i.e GET, POST...), however be aware that when creating a custom

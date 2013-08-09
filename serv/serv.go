// Serv is a tiny command-line webserver for serving static directories.
// It is intended for temporary debugging use only.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

var port = flag.Int("port", 9999, "Port to serve on.")
var root = flag.String("root", ".", "Target directory to serve.  Current dir by default.")

type LoggingHandler struct {
	http.Handler
}

type responseCodeCapturer struct {
	Code int
	http.ResponseWriter
}

func (r *responseCodeCapturer) WriteHeader(code int) {
	r.Code = code
	r.ResponseWriter.WriteHeader(code)
}

func (l *LoggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	wrapper := &responseCodeCapturer{0, w}
	l.Handler.ServeHTTP(wrapper, r)
	log.Printf("%10v %3d %s %s", time.Since(start), wrapper.Code, r.Method, r.RequestURI)
}

func main() {
	flag.Parse()
	log.SetFlags(log.Ldate | log.Ltime)
	http.Handle("/", &LoggingHandler{http.FileServer(http.Dir(*root))})
	log.Println("Listening on port", *port)
	http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
}

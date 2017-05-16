package application

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/gorilla/handlers"
	"github.com/rs/cors"
	"gopkg.in/tylerb/graceful.v1"
)

// Run sets up the server and starts listening
func Run(config *Config) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	app := NewApplication(config)

	cors := cors.New(cors.Options{AllowedOrigins: config.Origins})

	http.Handle("/", app.stats.Handler(handlers.ProxyHeaders(handlers.LoggingHandler(os.Stdout, cors.Handler(app.Router())))))

	fmt.Println("Starting server", config.Version, "on", config.Bind)

	// starts a server that waits until all current requests are finished
	// before shutting down. The server stops listening
	// immediately though, so a new process can listen on the port and start
	// serving requests straight away
	graceful.Run(config.Bind, 30*time.Second, nil)
}

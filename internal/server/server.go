package server

import (
	"fmt"
	"net/http"
	"net/url"
	"reverse-proxy-learn/internal/configs"
	"sync"
	"time"

	"embed"

	"golang.org/x/time/rate"
)

//go:embed templates/*
var templates embed.FS

var limiter = rate.NewLimiter(1, 4)

var mutex = &sync.Mutex{}

func limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			return
		}
		if !limiter.Allow() {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

var config configs.Configuration

// Run start the server on defined port
func Run() error {
	// load configurations from config file
	configo, err := configs.NewConfiguration()
	if err != nil {
		return fmt.Errorf("could not load configuration: %v", err)
	}
	config = *configo

	// Creates a new router
	mux := http.NewServeMux()

	// register health check endpoint
	mux.HandleFunc("/ping", ping)

	mux.HandleFunc("/login", login)

	// Iterating through the configuration resource and registering them
	// into the router.
	for _, resource := range config.Resources {
		url, _ := url.Parse(resource.Destination_URL)
		proxy := NewProxy(url)
		mux.HandleFunc(resource.Endpoint, ProxyRequestHandler(proxy, url, &resource))
	}

	// Running proxy server
	fmt.Printf("%s [ PROXY SERVER ] Starting at PORT %s\n", time.Now().UTC(), config.Server.Listen_port)
	if err := http.ListenAndServe(config.Server.Host+":"+config.Server.Listen_port, limit(mux)); err != nil {
		return fmt.Errorf("could not start the server: %v", err)
	}

	return nil
}

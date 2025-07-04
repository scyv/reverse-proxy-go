package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"reverse-proxy-learn/internal/configs"
	"strings"
	"time"
)

func NewProxy(target *url.URL) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)
	return proxy
}

func ProxyRequestHandler(proxy *httputil.ReverseProxy, url *url.URL, resource *configs.Resource) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s [ PROXY SERVER ] Request received at %s\n", time.Now().UTC(), r.URL)

		var AccessAllowed = false

		if len(resource.Access) == 0 {
			AccessAllowed = true
		} else {
			sessionCookie, err := r.Cookie("X-SCSESS")
			if sessionCookie != nil && err == nil {
				mutex.Lock()
				session, ok := sessions[sessionCookie.Value]
				if ok {
					if session.expires.After(time.Now()) {
						AccessAllowed = true
					} else {
						delete(sessions, sessionCookie.Value)
					}
				}
				mutex.Unlock()
			}
		}

		if !AccessAllowed {
			http.Redirect(w, r, fmt.Sprintf("/login?redirect=%s", r.URL.Path), http.StatusFound)
			return
		}

		path := r.URL.Path
		r.URL.Path = strings.TrimLeft(path, resource.Endpoint)

		// Update the headers to allow for SSL redirection
		r.URL.Host = url.Host
		r.URL.Scheme = url.Scheme
		r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
		r.Host = url.Host

		// Note that ServeHttp is non blocking and uses a go routine under the hood
		fmt.Printf("%s [ PROXY SERVER ] Proxying request to %s\n", time.Now().UTC(), r.URL)
		proxy.ServeHTTP(w, r)
	}
}

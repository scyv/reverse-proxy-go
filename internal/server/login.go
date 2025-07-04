package server

import (
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type Session struct {
	expires time.Time
	name    string
}

var chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func sessionId() string {
	result := make([]byte, 128)
	for i := 0; i < 128; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

var sessions = make(map[string]Session)

func login(w http.ResponseWriter, r *http.Request) {

	var redirectEndpoint = ""
	if r.Method == http.MethodPost {
		redirect := r.PostFormValue("redirect")
		key := r.PostFormValue("key")

		for _, resource := range config.Resources {
			if resource.Endpoint == redirect {
				for _, access := range resource.Access {
					if access.Key == key {
						mutex.Lock()
						sessionId := sessionId()
						now := time.Now()
						duration, _ := time.ParseDuration("1h")
						expiration := now.Add(duration)
						sessionCookie := http.Cookie{
							Name:     "X-SCSESS",
							Value:    sessionId,
							Path:     "/",
							MaxAge:   int(duration.Seconds()),
							HttpOnly: true,
							Secure:   true,
							SameSite: http.SameSiteStrictMode,
						}
						sessions[sessionId] = Session{
							name:    access.Name,
							expires: expiration,
						}
						mutex.Unlock()
						http.SetCookie(w, &sessionCookie)
						http.Redirect(w, r, redirect, http.StatusFound)
						return
					}
				}
			}
		}

	} else {
		if r.URL.Query().Has("redirect") {
			redirectEndpoint = r.URL.Query().Get("redirect")
		}
	}

	t, err := template.ParseFS(templates, "templates/login.html")
	if err != nil {
		log.Fatalf("error loading file: %s", err)
	}
	t.Execute(w, redirectEndpoint)
}

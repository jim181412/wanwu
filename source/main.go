package main

import (
	"log"
	"net/http"
	"os"

	"institute.supwisdom.com/authx-demo-cas-go/config"
	"institute.supwisdom.com/authx-demo-cas-go/session"
	_ "institute.supwisdom.com/authx-demo-cas-go/session/memory"
	"institute.supwisdom.com/authx-demo-cas-go/sso"
	"institute.supwisdom.com/authx-demo-cas-go/web"
)

func main() {
	cfg := config.Load()

	sessions, err := session.NewSessionManager("memory", "goSessionid", 3600)
	if err != nil {
		log.Fatal(err)
	}
	go sessions.GC()

	webHandler, err := web.NewHandler(cfg.ContextPath, cfg.AuthMode, sessions)
	if err != nil {
		log.Fatal(err)
	}
	ssoHandler := sso.NewHandler(cfg.AuthMode, cfg.CASServerURL, cfg.AppServerURL, sessions)
	rootHandler := func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("ticket") != "" || query.Get("returnUrl") != "" {
			ssoHandler.LoginHandler(w, r)
			return
		}

		webHandler.IndexHandler(w, r)
	}

	mux := http.NewServeMux()
	mux.HandleFunc(cfg.ContextPath, rootHandler)
	mux.HandleFunc(cfg.ContextPath+"/", rootHandler)
	mux.HandleFunc(cfg.ContextPath+"/index", webHandler.IndexHandler)

	mux.HandleFunc(cfg.ContextPath+"/sso/login", ssoHandler.LoginHandler)
	mux.HandleFunc(cfg.ContextPath+"/sso/logout", ssoHandler.LogoutHandler)
	mux.HandleFunc(cfg.ContextPath+"/sso/userOnlineDetect", ssoHandler.UserOnlineDetectHandler)
	mux.HandleFunc(cfg.ContextPath+"/sso/slo", ssoHandler.SloHandler)

	if _, err := os.Stat("demo"); err == nil {
		mux.Handle(cfg.ContextPath+"/demo/", http.StripPrefix(cfg.ContextPath+"/demo/", http.FileServer(http.Dir("demo"))))
	}

	log.Printf("server starting on %s contextPath=%s authMode=%s", cfg.ListenAddr, cfg.ContextPath, cfg.AuthMode)
	log.Fatal(http.ListenAndServe(cfg.ListenAddr, mux))
}

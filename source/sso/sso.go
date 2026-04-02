package sso

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"institute.supwisdom.com/authx-demo-cas-go/internal/cas"
	"institute.supwisdom.com/authx-demo-cas-go/session"
)

const sessionCasUserKey = "casUser"

var jsonpCallbackPattern = regexp.MustCompile(`^[a-zA-Z_$][0-9a-zA-Z_$.]*$`)

type Handler struct {
	authMode     string
	casServerURL string
	appServerURL string
	sessions     *session.Manager
}

type responseEnvelope struct {
	Code    int         `json:"code"`
	Message interface{} `json:"message"`
	Data    interface{} `json:"data"`
}

func NewHandler(authMode string, casServerURL string, appServerURL string, sessions *session.Manager) *Handler {
	return &Handler{
		authMode:     authMode,
		casServerURL: casServerURL,
		appServerURL: appServerURL,
		sessions:     sessions,
	}
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid query", http.StatusBadRequest)
		return
	}

	service := h.appServerURL
	returnURL := r.FormValue("returnUrl")
	if returnURL != "" {
		service += "?returnUrl=" + url.QueryEscape(returnURL)
	}

	log.Println("sso.go", "returnURL", returnURL)

	if h.useMockAuth() {
		h.loginMockUser(w, r, returnURL)
		return
	}

	casUser, redirect := cas.Login(h.casServerURL, service, r.Form)
	if redirect != "" {
		http.Redirect(w, r, redirect, http.StatusFound)
		return
	}

	if casUser == nil {
		http.Error(w, "cas service validate failed", http.StatusBadGateway)
		return
	}

	if casUser.IsEmpty() {
		http.Error(w, "cas service validate failed", http.StatusUnauthorized)
		return
	}

	sess := h.sessions.SessionStart(w, r)
	if err := sess.Set(sessionCasUserKey, casUser); err != nil {
		http.Error(w, "save session failed", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, h.resolveReturnURL(returnURL), http.StatusFound)
}

func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid query", http.StatusBadRequest)
		return
	}

	service := h.appServerURL + "/sso/logout"
	returnURL := r.FormValue("returnUrl")
	if returnURL != "" {
		service += "?returnUrl=" + url.QueryEscape(returnURL)
	}

	if h.useMockAuth() {
		h.sessions.SessionDestroy(w, r)
		http.Redirect(w, r, h.resolveReturnURL(returnURL), http.StatusFound)
		return
	}

	isLogout, redirect := cas.Logout(h.casServerURL, service, r.Form)
	if !isLogout {
		http.Redirect(w, r, redirect, http.StatusFound)
		return
	}

	h.sessions.SessionDestroy(w, r)
	http.Redirect(w, r, h.resolveReturnURL(returnURL), http.StatusFound)
}

// 检测是否在线
func (h *Handler) UserOnlineDetectHandler(w http.ResponseWriter, r *http.Request) {
	isAlive := false

	sess := h.sessions.SessionStart(w, r)
	if casUser, ok := sess.Get(sessionCasUserKey).(*cas.CasUser); ok {
		if h.useMockAuth() {
			isAlive = !casUser.IsEmpty()
		} else {
			isAlive = cas.UserOnlineDetect(h.casServerURL, casUser)
		}
	}

	writeJSON(w, http.StatusOK, responseEnvelope{
		Code:    0,
		Message: nil,
		Data: map[string]bool{
			"isAlive": isAlive,
		},
	})
}

func (h *Handler) SloHandler(w http.ResponseWriter, r *http.Request) {
	callback := ""
	if err := r.ParseForm(); err == nil {
		callback = r.FormValue("callback")
	}

	h.sessions.SessionDestroy(w, r)

	payload := responseEnvelope{
		Code:    0,
		Message: nil,
		Data: map[string]string{
			"success": "注销成功",
		},
	}

	if callback == "" {
		writeJSON(w, http.StatusOK, payload)
		return
	}

	if !jsonpCallbackPattern.MatchString(callback) {
		http.Error(w, "invalid callback", http.StatusBadRequest)
		return
	}

	responseJSON, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "encode response failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/javascript;charset=UTF-8")
	if _, err := w.Write([]byte(callback + "(" + string(responseJSON) + ");")); err != nil {
		log.Printf("write jsonp response failed: %v", err)
	}
}

func (h *Handler) resolveReturnURL(raw string) string {
	if raw == "" {
		return h.appServerURL
	}

	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}

	if strings.HasPrefix(raw, "/") {
		return h.appServerURL + raw
	}

	return h.appServerURL + "/" + raw
}

func (h *Handler) useMockAuth() bool {
	return h.authMode == "mock" || h.casServerURL == ""
}

func (h *Handler) loginMockUser(w http.ResponseWriter, r *http.Request, returnURL string) {
	userName := strings.TrimSpace(r.FormValue("user"))
	if userName == "" {
		userName = "demo-user"
	}

	casUser := &cas.CasUser{
		Service: h.appServerURL,
		Ticket:  "mock-ticket",
		User:    userName,
		Attributes: map[string][]string{
			"name":             {"本地演示用户"},
			"userName":         {userName},
			"identityTypeName": {"Mock Login"},
		},
	}

	sess := h.sessions.SessionStart(w, r)
	if err := sess.Set(sessionCasUserKey, casUser); err != nil {
		http.Error(w, "save session failed", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, h.resolveReturnURL(returnURL), http.StatusFound)
}

func writeJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write json response failed: %v", err)
	}
}

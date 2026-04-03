package casauth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	invalidRequest           = "INVALID_REQUEST"
	invalidTicketSpec        = "INVALID_TICKET_SPEC"
	unauthorizedService      = "UNAUTHORIZED_SERVICE"
	unauthorizedServiceProxy = "UNAUTHORIZED_SERVICE_PROXY"
	invalidProxyCallback     = "INVALID_PROXY_CALLBACK"
	invalidTicket            = "INVALID_TICKET"
	invalidService           = "INVALID_SERVICE"
	internalError            = "INTERNAL_ERROR"
	defaultRequestTimeout    = 10 * time.Second
)

var defaultHTTPClient = &http.Client{
	Timeout: defaultRequestTimeout,
}

type User struct {
	Service    string
	Ticket     string
	User       string
	Attributes map[string][]string
}

func (u *User) IsEmpty() bool {
	return u == nil || strings.TrimSpace(u.User) == ""
}

func NormalizeServerURL(raw string) string {
	baseURL := strings.TrimRight(strings.TrimSpace(raw), "/")
	switch {
	case strings.HasSuffix(baseURL, "/login"):
		return strings.TrimSuffix(baseURL, "/login")
	case strings.HasSuffix(baseURL, "/logout"):
		return strings.TrimSuffix(baseURL, "/logout")
	case strings.HasSuffix(baseURL, "/serviceValidate"):
		return strings.TrimSuffix(baseURL, "/serviceValidate")
	default:
		return baseURL
	}
}

func LoginURL(serverURL, service string) string {
	return NormalizeServerURL(serverURL) + "/login?service=" + url.QueryEscape(service)
}

func LogoutURL(serverURL, service string) string {
	return NormalizeServerURL(serverURL) + "/logout?service=" + url.QueryEscape(service)
}

func ServiceValidate(serverURL, service, ticket string) (*User, error) {
	serviceValidateURL := NormalizeServerURL(serverURL) + "/serviceValidate?service=" + url.QueryEscape(service) + "&ticket=" + url.QueryEscape(ticket)
	responseXML, err := httpText(http.MethodGet, serviceValidateURL)
	if err != nil {
		return nil, err
	}
	success, err := ParseServiceResponse([]byte(responseXML))
	if err != nil {
		return nil, err
	}
	return &User{
		Service:    service,
		Ticket:     ticket,
		User:       success.User,
		Attributes: success.Attributes,
	}, nil
}

func UserOnlineDetect(serverURL string, user *User) bool {
	if user == nil || user.IsEmpty() {
		return false
	}
	return UserOnlineDetect2(serverURL, user.Service, user.Ticket, user.User)
}

func UserOnlineDetect2(serverURL, service, ticket, username string) bool {
	userOnlineDetectURL := NormalizeServerURL(serverURL) + "/login/userOnlineDetect?service=" + url.QueryEscape(service) + "&ticket=" + url.QueryEscape(ticket) + "&username=" + url.QueryEscape(username)
	responseJSON, err := httpText(http.MethodPost, userOnlineDetectURL)
	if err != nil {
		return false
	}
	var payload jsonResponse
	if err := json.Unmarshal([]byte(responseJSON), &payload); err != nil {
		return false
	}
	if payload.Error != nil || payload.Data == nil {
		return false
	}
	return payload.Code == 0 && payload.Data.IsAlive
}

type AuthenticationError struct {
	Code    string
	Message string
}

func (e *AuthenticationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AuthenticationError) AuthenticationError() bool {
	return true
}

type jsonResponse struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    *jsonData  `json:"data"`
	Error   *jsonError `json:"error"`
}

type jsonData struct {
	IsAlive bool `json:"isAlive"`
}

type jsonError struct {
	Error string `json:"error"`
}

func httpText(method, target string) (string, error) {
	req, err := http.NewRequest(method, target, nil)
	if err != nil {
		return "", err
	}
	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("unexpected %s status: %s", method, resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

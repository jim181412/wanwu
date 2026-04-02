package cas

import (
	"encoding/json"
	"log"
	"net/url"
	"strings"
)

type CasUser struct {
	Service string
	Ticket  string
	User    string

	// Attributes 保存 CAS 返回的扩展用户信息。
	Attributes map[string][]string
}

func NewCasUser() *CasUser {
	return &CasUser{}
}

func (casUser CasUser) IsEmpty() bool {
	return casUser.User == ""
}

func Login(casServerHostURL string, service string, requestParams map[string][]string) (*CasUser, string) {
	ticket := firstQueryValue(requestParams, "ticket")
	log.Println(ticket)
	if ticket == "" {
		redirect := casServerHostURL + "/login?service=" + url.QueryEscape(service)
		log.Println("redirect to login", redirect)
		return nil, redirect
	}

	log.Println("ticket:", ticket)

	casUser := ServiceValidate(casServerHostURL, service, ticket)
	log.Println("casUser:", casUser)

	return casUser, ""
}

func Logout(casServerHostURL string, service string, requestParams map[string][]string) (bool, string) {
	logout := firstQueryValue(requestParams, "logout")
	if logout == "" {
		if !strings.Contains(service, "?") {
			service += "?"
		} else {
			service += "&"
		}
		service += "logout=logout"

		redirect := casServerHostURL + "/logout?service=" + url.QueryEscape(service)
		log.Println("redirect to logout", redirect)
		return false, redirect
	}

	return true, ""
}

func ServiceValidate(casServerHostURL string, service string, ticket string) *CasUser {
	serviceValidateURL := casServerHostURL + "/serviceValidate?service=" + url.QueryEscape(service) + "&ticket=" + ticket

	responseXML, err := httpGet(serviceValidateURL)
	if err != nil {
		log.Println("service validate request failed:", err)
		return nil
	}
	log.Println("responseXml:", responseXML)

	success, err := ParseServiceResponse([]byte(responseXML))
	if err != nil {
		log.Println("AuthenticationError:", err.Error())
		return NewCasUser()
	}

	casUser := NewCasUser()
	casUser.Service = service
	casUser.Ticket = ticket
	casUser.User = success.User
	casUser.Attributes = success.Attributes

	return casUser
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

func UserOnlineDetect(casServerHostURL string, casUser *CasUser) bool {
	return UserOnlineDetect2(casServerHostURL, casUser.Service, casUser.Ticket, casUser.User)
}

func UserOnlineDetect2(casServerHostURL string, service string, ticket string, username string) bool {
	userOnlineDetectURL := casServerHostURL + "/login/userOnlineDetect?service=" + url.QueryEscape(service) + "&ticket=" + ticket + "&username=" + username

	responseJSON, err := httpPost(userOnlineDetectURL)
	if err != nil {
		log.Println("user online detect request failed:", err)
		return false
	}
	log.Println("responseJson:", responseJSON)

	var payload jsonResponse
	if err := json.Unmarshal([]byte(responseJSON), &payload); err != nil {
		return false
	}

	if payload.Error != nil || payload.Data == nil {
		return false
	}

	if payload.Code == 0 {
		return payload.Data.IsAlive
	}

	return false
}

func firstQueryValue(params map[string][]string, key string) string {
	values := params[key]
	if len(values) == 0 {
		return ""
	}

	return values[0]
}

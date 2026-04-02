package cas

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

const (
	INVALID_REQUEST            = "INVALID_REQUEST"
	INVALID_TICKET_SPEC        = "INVALID_TICKET_SPEC"
	UNAUTHORIZED_SERVICE       = "UNAUTHORIZED_SERVICE"
	UNAUTHORIZED_SERVICE_PROXY = "UNAUTHORIZED_SERVICE_PROXY"
	INVALID_PROXY_CALLBACK     = "INVALID_PROXY_CALLBACK"
	INVALID_TICKET             = "INVALID_TICKET"
	INVALID_SERVICE            = "INVALID_SERVICE"
	INTERNAL_ERROR             = "INTERNAL_ERROR"
)

type AuthenticationError struct {
	Code    string
	Message string
}

func (e AuthenticationError) AuthenticationError() bool {
	return true
}

func (e AuthenticationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

type AuthenticationResponse struct {
	User                string
	ProxyGrantingTicket string
	Proxies             []string
	AuthenticationDate  time.Time
	IsNewLogin          bool
	IsRememberedLogin   bool
	MemberOf            []string
	Attributes          UserAttributes
}

type UserAttributes map[string][]string

func (a UserAttributes) Get(name string) string {
	if v, ok := a[name]; ok {
		return v[0]
	}

	return ""
}

func (a UserAttributes) Add(name string, value string) {
	a[name] = append(a[name], value)
}

func ParseServiceResponse(data []byte) (*AuthenticationResponse, error) {
	var parsed xmlServiceResponse
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, err
	}

	if parsed.Failure != nil {
		msg := strings.TrimSpace(parsed.Failure.Message)
		return nil, &AuthenticationError{Code: parsed.Failure.Code, Message: msg}
	}

	response := &AuthenticationResponse{
		User:                parsed.Success.User,
		ProxyGrantingTicket: parsed.Success.ProxyGrantingTicket,
		Attributes:          make(UserAttributes),
	}

	if proxies := parsed.Success.Proxies; proxies != nil {
		response.Proxies = proxies.Proxies
	}

	if attributes := parsed.Success.Attributes; attributes != nil {
		response.AuthenticationDate = attributes.AuthenticationDate
		response.IsRememberedLogin = attributes.LongTermAuthenticationRequestTokenUsed
		response.IsNewLogin = attributes.IsFromNewLogin
		response.MemberOf = attributes.MemberOf

		if attributes.UserAttributes != nil {
			for _, attribute := range attributes.UserAttributes.Attributes {
				if attribute.Name == "" {
					continue
				}

				response.Attributes.Add(attribute.Name, strings.TrimSpace(attribute.Value))
			}

			for _, attribute := range attributes.UserAttributes.AnyAttributes {
				response.Attributes.Add(attribute.XMLName.Local, strings.TrimSpace(attribute.Value))
			}
		}

		if attributes.ExtraAttributes != nil {
			for _, attribute := range attributes.ExtraAttributes {
				response.Attributes.Add(attribute.XMLName.Local, strings.TrimSpace(attribute.Value))
			}
		}
	}

	return response, nil
}

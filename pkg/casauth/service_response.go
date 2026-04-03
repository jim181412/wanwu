package casauth

import (
	"encoding/xml"
	"strings"
	"time"
)

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
	if len(a[name]) == 0 {
		return ""
	}
	return a[name][0]
}

func (a UserAttributes) Add(name, value string) {
	a[name] = append(a[name], value)
}

func ParseServiceResponse(data []byte) (*AuthenticationResponse, error) {
	var parsed xmlServiceResponse
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return nil, err
	}
	if parsed.Failure != nil {
		return nil, &AuthenticationError{
			Code:    parsed.Failure.Code,
			Message: strings.TrimSpace(parsed.Failure.Message),
		}
	}
	response := &AuthenticationResponse{
		User:       parsed.Success.User,
		Attributes: make(UserAttributes),
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

type xmlServiceResponse struct {
	XMLName xml.Name `xml:"http://www.yale.edu/tp/cas serviceResponse"`

	Failure *xmlAuthenticationFailure
	Success *xmlAuthenticationSuccess
}

type xmlAuthenticationFailure struct {
	XMLName xml.Name `xml:"authenticationFailure"`
	Code    string   `xml:"code,attr"`
	Message string   `xml:",innerxml"`
}

type xmlAuthenticationSuccess struct {
	XMLName             xml.Name           `xml:"authenticationSuccess"`
	User                string             `xml:"user"`
	ProxyGrantingTicket string             `xml:"proxyGrantingTicket,omitempty"`
	Proxies             *xmlProxies        `xml:"proxies"`
	Attributes          *xmlAttributes     `xml:"attributes"`
	ExtraAttributes     []*xmlAnyAttribute `xml:",any"`
}

type xmlProxies struct {
	XMLName xml.Name `xml:"proxies"`
	Proxies []string `xml:"proxy"`
}

type xmlAttributes struct {
	XMLName                                xml.Name `xml:"attributes"`
	AuthenticationDate                     time.Time
	LongTermAuthenticationRequestTokenUsed bool `xml:"longTermAuthenticationRequestTokenUsed"`
	IsFromNewLogin                         bool `xml:"isFromNewLogin"`
	MemberOf                               []string
	UserAttributes                         *xmlUserAttributes
	ExtraAttributes                        []*xmlAnyAttribute `xml:",any"`
}

type xmlUserAttributes struct {
	XMLName       xml.Name             `xml:"userAttributes"`
	Attributes    []*xmlNamedAttribute `xml:"attribute"`
	AnyAttributes []*xmlAnyAttribute   `xml:",any"`
}

type xmlNamedAttribute struct {
	XMLName xml.Name `xml:"attribute"`
	Name    string   `xml:"name,attr,omitempty"`
	Value   string   `xml:",innerxml"`
}

type xmlAnyAttribute struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

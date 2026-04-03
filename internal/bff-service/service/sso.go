package service

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/casauth"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/gin-gonic/gin"
)

const (
	ssoModeMock = "mock"
	ssoModeCAS  = "cas"
	ssoFlag     = "1"
	mockTicket  = "mock-ticket"
)

func LoginBySSO(ctx *gin.Context, callbackURL, mockUser string) (string, error) {
	if !config.Cfg().UnifiedAuth.Enabled {
		return "", grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_sso_disabled")
	}
	serviceURL, err := buildSSOServiceURL(callbackURL)
	if err != nil {
		return "", err
	}
	if unifiedAuthMode() == ssoModeMock {
		redirectURL, err := appendMockTicket(serviceURL, firstNonEmpty(mockUser, config.Cfg().UnifiedAuth.MockUser))
		if err != nil {
			return "", grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_sso_login_redirect", err.Error())
		}
		return redirectURL, nil
	}
	if config.Cfg().UnifiedAuth.CASServerURL == "" {
		return "", grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_sso_cas_server_empty")
	}
	return casauth.LoginURL(config.Cfg().UnifiedAuth.CASServerURL, serviceURL), nil
}

func ExchangeSSOLogin(ctx *gin.Context, callbackURL, ticket, language string) (*response.Login, error) {
	if !config.Cfg().UnifiedAuth.Enabled {
		return nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_sso_disabled")
	}
	serviceURL, err := buildSSOServiceURL(callbackURL)
	if err != nil {
		return nil, err
	}
	casUser, err := getSSOUser(serviceURL, callbackURL, ticket)
	if err != nil {
		return nil, err
	}
	username, err := resolveSSOUsername(casUser)
	if err != nil {
		return nil, err
	}
	userID, err := findSSOLocalUserID(ctx, username)
	if err != nil {
		return nil, err
	}
	return buildSSOLoginResp(ctx, userID, language)
}

func LogoutBySSO(ctx *gin.Context, callbackURL string) (string, error) {
	if !config.Cfg().UnifiedAuth.Enabled {
		return "", grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_sso_disabled")
	}
	redirectURL, err := normalizeSSOCallbackURL(callbackURL)
	if err != nil {
		return "", err
	}
	if unifiedAuthMode() == ssoModeMock {
		return redirectURL, nil
	}
	if config.Cfg().UnifiedAuth.CASServerURL == "" {
		return "", grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_sso_cas_server_empty")
	}
	return casauth.LogoutURL(config.Cfg().UnifiedAuth.CASServerURL, redirectURL), nil
}

func buildSSOLoginResp(ctx *gin.Context, userID, language string) (*response.Login, error) {
	orgs, err := iam.GetOrgSelect(ctx.Request.Context(), &iam_service.GetOrgSelectReq{UserId: userID})
	if err != nil {
		return nil, err
	}
	orgID, permission, err := selectFirstAvailableOrgPermission(ctx, userID, orgs.Selects)
	if err != nil {
		return nil, err
	}
	if language != "" {
		if _, err := iam.ChangeUserLanguage(ctx.Request.Context(), &iam_service.ChangeUserLanguageReq{
			UserId:   userID,
			Language: language,
		}); err != nil {
			return nil, err
		}
	}
	userInfo, err := iam.GetUserInfo(ctx.Request.Context(), &iam_service.GetUserInfoReq{
		UserId: userID,
		OrgId:  orgID,
	})
	if err != nil {
		return nil, err
	}
	if language != "" {
		userInfo.Language = language
	}
	return buildLoginResp(ctx, userInfo, permission, orgs.Selects)
}

func selectFirstAvailableOrgPermission(ctx *gin.Context, userID string, orgs []*iam_service.IDName) (string, *iam_service.UserPermission, error) {
	if len(orgs) == 0 && userID == config.SystemAdminUserID {
		orgs = []*iam_service.IDName{{Id: config.TopOrgID}}
	}
	for _, org := range orgs {
		permission, err := iam.GetUserPermission(ctx.Request.Context(), &iam_service.GetUserPermissionReq{
			UserId: userID,
			OrgId:  org.Id,
		})
		if err == nil {
			return org.Id, permission, nil
		}
	}
	return "", nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_sso_org_permission_empty", userID)
}

func findSSOLocalUserID(ctx *gin.Context, username string) (string, error) {
	users, err := iam.GetUserList(ctx.Request.Context(), &iam_service.GetUserListReq{
		OrgId:    config.TopOrgID,
		UserName: username,
		PageNo:   1,
		PageSize: 20,
	})
	if err != nil {
		return "", err
	}
	var matched []*iam_service.UserInfo
	for _, user := range users.Users {
		if user.GetUserName() == username {
			matched = append(matched, user)
		}
	}
	switch len(matched) {
	case 0:
		return "", grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_sso_user_not_found", username)
	case 1:
		return matched[0].GetUserId(), nil
	default:
		return "", grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_sso_user_duplicate", username)
	}
}

func resolveSSOUsername(casUser *casauth.User) (string, error) {
	if casUser == nil || casUser.IsEmpty() {
		return "", grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_sso_user_empty")
	}
	attrKeys := []string{config.Cfg().UnifiedAuth.UsernameAttr, "userName", "username", "name"}
	for _, attrKey := range attrKeys {
		attrKey = strings.TrimSpace(attrKey)
		if attrKey == "" {
			continue
		}
		values := casUser.Attributes[attrKey]
		if len(values) == 0 {
			continue
		}
		if username := strings.TrimSpace(values[0]); username != "" {
			return username, nil
		}
	}
	if username := strings.TrimSpace(casUser.User); username != "" {
		return username, nil
	}
	return "", grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_sso_username_empty")
}

func getSSOUser(serviceURL, callbackURL, ticket string) (*casauth.User, error) {
	switch unifiedAuthMode() {
	case ssoModeMock:
		callback, parseErr := url.Parse(callbackURL)
		if parseErr != nil {
			return nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_sso_callback_invalid", parseErr.Error())
		}
		mockUser := firstNonEmpty(callback.Query().Get("mockUser"), config.Cfg().UnifiedAuth.MockUser, "demo-user")
		return &casauth.User{
			Service: serviceURL,
			Ticket:  firstNonEmpty(ticket, mockTicket),
			User:    mockUser,
			Attributes: map[string][]string{
				"userName": {mockUser},
			},
		}, nil
	case ssoModeCAS:
		if ticket == "" {
			return nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_sso_ticket_empty")
		}
		casUser, err := casauth.ServiceValidate(config.Cfg().UnifiedAuth.CASServerURL, serviceURL, ticket)
		if err != nil {
			return nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_sso_validate_failed", err.Error())
		}
		return casUser, nil
	default:
		return nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_sso_mode_invalid", config.Cfg().UnifiedAuth.Mode)
	}
}

func buildSSOServiceURL(raw string) (string, error) {
	callbackURL, err := normalizeSSOCallbackURL(raw)
	if err != nil {
		return "", err
	}
	parsed, err := url.Parse(callbackURL)
	if err != nil {
		return "", grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_sso_callback_invalid", err.Error())
	}
	query := parsed.Query()
	query.Set("sso", ssoFlag)
	query.Del("ticket")
	parsed.RawQuery = query.Encode()
	parsed.Fragment = ""
	return parsed.String(), nil
}

func normalizeSSOCallbackURL(raw string) (string, error) {
	callbackURL := strings.TrimSpace(raw)
	if callbackURL == "" {
		callbackURL = strings.TrimRight(config.Cfg().Server.WebBaseUrl, "/") + "/aibase/login"
	}
	parsed, err := url.Parse(callbackURL)
	if err != nil || !parsed.IsAbs() {
		return "", grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_sso_callback_invalid", callbackURL)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_sso_callback_invalid", callbackURL)
	}
	if !strings.HasSuffix(parsed.Path, "/aibase/login") {
		return "", grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_sso_callback_invalid", callbackURL)
	}
	if !isAllowedSSOCallbackHost(parsed.Hostname()) {
		return "", grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_sso_callback_host_invalid", parsed.Hostname())
	}
	return parsed.String(), nil
}

func isAllowedSSOCallbackHost(host string) bool {
	if host == "" {
		return false
	}
	webBaseHost := hostOnly(config.Cfg().Server.WebBaseUrl)
	if strings.EqualFold(host, webBaseHost) {
		return true
	}
	return isLocalHost(host)
}

func isLocalHost(host string) bool {
	switch strings.ToLower(strings.TrimSpace(host)) {
	case "localhost", "127.0.0.1", "::1":
		return true
	default:
		return net.ParseIP(host) != nil && net.ParseIP(host).IsLoopback()
	}
}

func hostOnly(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return ""
	}
	return parsed.Hostname()
}

func appendMockTicket(serviceURL, mockUser string) (string, error) {
	parsed, err := url.Parse(serviceURL)
	if err != nil {
		return "", err
	}
	query := parsed.Query()
	query.Set("ticket", mockTicket)
	if mockUser != "" {
		query.Set("mockUser", mockUser)
	}
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func unifiedAuthMode() string {
	switch strings.ToLower(strings.TrimSpace(config.Cfg().UnifiedAuth.Mode)) {
	case ssoModeMock:
		return ssoModeMock
	case ssoModeCAS:
		return ssoModeCAS
	default:
		if config.Cfg().UnifiedAuth.CASServerURL == "" {
			return ssoModeMock
		}
		return ssoModeCAS
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func formatSSOErr(format string, args ...interface{}) error {
	return grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_sso_general", fmt.Sprintf(format, args...))
}

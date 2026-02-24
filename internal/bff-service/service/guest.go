package service

import (
	"fmt"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	operate_service "github.com/UnicomAI/wanwu/api/proto/operate-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/gin-gonic/gin"
)

const (
	customModeLight = "light"
	customModeDark  = "dark"
)

func GetLanguageSelect() *response.LanguageSelect {
	language := make([]response.Language, len(config.Cfg().I18n.Langs))
	for i, lang := range config.Cfg().I18n.Langs {
		language[i].Code = lang.Code
		language[i].Name = lang.Name
	}
	return &response.LanguageSelect{
		Languages:       language,
		DefaultLanguage: getLanguageByCode(config.Cfg().I18n.DefaultLang),
	}
}

func GetLogoCustomInfo(ctx *gin.Context, mode string) (response.LogoCustomInfo, error) {
	ret := response.LogoCustomInfo{}
	var theme string
	switch mode {
	case customModeLight:
		theme = customModeLight
	case customModeDark:
		theme = customModeDark
	default:
		theme = config.Cfg().CustomInfo.DefaultMode
	}
	for _, mode := range config.Cfg().CustomInfo.Modes {
		if theme != mode.Mode {
			continue
		}
		ret = response.LogoCustomInfo{
			Login: response.CustomLogin{
				Background:       request.Avatar{Path: mode.Login.BackgroundPath},
				Logo:             request.Avatar{Path: mode.Login.LogoPath},
				LoginButtonColor: mode.Login.LoginButtonColor,
				WelcomeText:      gin_util.I18nKey(ctx, mode.Login.WelcomeText),
			},
			Home: response.CustomHome{
				Logo:            request.Avatar{Path: mode.Home.LogoPath},
				Title:           gin_util.I18nKey(ctx, mode.Home.Title),
				BackgroundColor: mode.Home.BackgroundColor,
			},
			Tab: response.CustomTab{
				Logo:  request.Avatar{Path: mode.Tab.TabLogoPath},
				Title: gin_util.I18nKey(ctx, mode.Tab.TabTitle),
			},
			About: response.CustomAbout{
				LogoPath:  mode.About.LogoPath,
				Version:   config.Cfg().CustomInfo.Version,
				Copyright: gin_util.I18nKey(ctx, mode.About.Copyright),
			},
			LinkList:      config.Cfg().DocCenter.GetDocs(),
			Register:      response.CustomRegister{Email: response.CustomEmail{Status: config.Cfg().CustomInfo.RegisterByEmail != 0}},
			ResetPassword: response.CustomResetPassword{Email: response.CustomEmail{Status: config.Cfg().CustomInfo.ResetPasswordByEmail != 0}},
			LoginEmail:    response.CustomLoginEmail{Email: response.CustomEmail{Status: config.Cfg().CustomInfo.LoginByEmail != 0}},
			DefaultIcon: response.CustomDefaultIcon{
				RagIcon:      config.Cfg().DefaultIcon.RagIcon,
				AgentIcon:    config.Cfg().DefaultIcon.AgentIcon,
				WorkflowIcon: config.Cfg().DefaultIcon.WorkflowIcon,
				PromptIcon:   config.Cfg().DefaultIcon.PromptIcon,
				ChatflowIcon: config.Cfg().DefaultIcon.ChatflowIcon,
			},
		}
		break
	}
	custom, err := operate.GetSystemCustom(ctx.Request.Context(), &operate_service.GetSystemCustomReq{Mode: theme})
	if err != nil {
		return ret, err
	}
	if custom.Tab.TabLogoPath != "" {
		ret.Tab.Logo = CacheAvatar(ctx, custom.Tab.TabLogoPath, false)
	}
	if custom.Tab.TabTitle != "" {
		ret.Tab.Title = custom.Tab.TabTitle
	}
	if custom.Login.LoginBgPath != "" {
		ret.Login.Background = CacheAvatar(ctx, custom.Login.LoginBgPath, false)
	}
	if custom.Login.LoginLogo != "" {
		ret.Login.Logo = CacheAvatar(ctx, custom.Login.LoginLogo, false)
	}
	if custom.Login.LoginButtonColor != "" {
		ret.Login.LoginButtonColor = custom.Login.LoginButtonColor
	}
	if custom.Login.LoginWelcomeText != "" {
		ret.Login.WelcomeText = custom.Login.LoginWelcomeText
	}
	if custom.Home.HomeName != "" {
		ret.Home.Title = custom.Home.HomeName
	}
	if custom.Home.HomeLogoPath != "" {
		ret.Home.Logo = CacheAvatar(ctx, custom.Home.HomeLogoPath, false)
	}
	if custom.Home.HomeBgColor != "" {
		ret.Home.BackgroundColor = custom.Home.HomeBgColor
	}
	return ret, nil
}

func GetCaptcha(ctx *gin.Context, key string) (*response.Captcha, error) {
	resp, err := iam.GetCaptcha(ctx.Request.Context(), &iam_service.GetCaptchaReq{
		Key: key,
	})
	if err != nil {
		return nil, err
	}
	return &response.Captcha{
		Key: key,
		B64: resp.B64,
	}, nil
}

func RegisterByEmail(ctx *gin.Context, register *request.RegisterByEmail) error {
	if config.Cfg().CustomInfo.RegisterByEmail == 0 {
		return grpc_util.ErrorStatus(errs.Code_BFFRegisterDisable)
	}
	_, err := iam.RegisterByEmail(ctx.Request.Context(), &iam_service.RegisterByEmailReq{
		UserName: register.Username,
		Email:    register.Email,
		Code:     register.Code,
	})
	return err
}

func RegisterSendEmailCode(ctx *gin.Context, username, email string) error {
	if config.Cfg().CustomInfo.RegisterByEmail == 0 {
		return grpc_util.ErrorStatus(errs.Code_BFFRegisterDisable)
	}
	_, err := iam.RegisterSendEmailCode(ctx.Request.Context(), &iam_service.RegisterSendEmailCodeReq{
		Email:    email,
		UserName: username,
	})
	return err
}

// --- reset password---
func ResetPasswordSendEmailCode(ctx *gin.Context, email string) error {
	if config.Cfg().CustomInfo.ResetPasswordByEmail == 0 {
		return grpc_util.ErrorStatus(errs.Code_BFFResetPasswordDisable)
	}
	_, err := iam.ResetPasswordSendEmailCode(ctx.Request.Context(), &iam_service.ResetPasswordSendEmailCodeReq{
		Email: email,
	})
	return err
}

func ResetPasswordByEmail(ctx *gin.Context, reset *request.ResetPasswordByEmail) error {
	if config.Cfg().CustomInfo.ResetPasswordByEmail == 0 {
		return grpc_util.ErrorStatus(errs.Code_BFFResetPasswordDisable)
	}
	password, err := decryptPD(reset.Password)
	if err != nil {
		return fmt.Errorf("decrypt password err: %v", err)
	}
	_, err = iam.ResetPasswordByEmail(ctx.Request.Context(), &iam_service.ResetPasswordByEmailReq{
		Email:    reset.Email,
		Password: password,
		Code:     reset.Code,
	})
	return err
}

// --- internal ---

func getLanguageByCode(languageCode string) response.Language {
	langs := config.Cfg().I18n.Langs
	language := response.Language{Code: languageCode}
	for _, lang := range langs {
		if lang.Code == languageCode {
			language.Name = lang.Name
		}
	}
	return language
}

func toOrgPermission(ctx *gin.Context, orgPerm *iam_service.UserPermission) response.UserOrgPermission {
	return response.UserOrgPermission{
		IsAdmin:     orgPerm.IsAdmin,
		IsSystem:    orgPerm.IsSystem,
		Org:         toOrgIDName(ctx, orgPerm.Org),
		Roles:       toRoleIDNames(ctx, orgPerm.Roles),
		Permissions: toPermissions(orgPerm.IsAdmin, orgPerm.IsSystem, orgPerm.Perms),
	}
}

func toPermissions(isAdmin, isSystem bool, perms []*iam_service.Perm) []response.Permission {
	routes := mid.CollectPerms()
	var ret []response.Permission
	if isAdmin {
		for _, r := range routes {
			if isSystem && r.Tag == "permission.role" {
				continue
			}
			if !isSystem && r.Tag == "setting" {
				continue
			}
			if !isSystem && r.Tag == "statistic_client" || config.Cfg().WorkflowTemplate.ServerMode == "remote" {
				continue
			}
			if !isSystem && r.Tag == "operation" {
				continue
			}
			if !isSystem && r.Tag == "operation.oauth" {
				continue
			}
			ret = append(ret, response.Permission{
				Perm: r.Tag,
				Name: r.Name,
			})
		}
		return ret
	}
	for _, r := range routes {
		if r.Tag == "setting" {
			continue
		}
		if r.Tag == "statistic_client" {
			continue
		}
		if r.Tag == "oauth" {
			continue
		}
		for _, perm := range perms {
			if perm.Perm == r.Tag {
				ret = append(ret, response.Permission{
					Perm: r.Tag,
					Name: r.Name,
				})
				break
			}
		}
	}
	return ret
}

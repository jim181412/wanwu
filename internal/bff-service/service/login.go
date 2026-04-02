package service

import (
	"fmt"
	"strconv"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	jwt_util "github.com/UnicomAI/wanwu/pkg/jwt-util"
	"github.com/gin-gonic/gin"
)

// 普通的登陆方法
func Login(ctx *gin.Context, login *request.Login, language string) (*response.Login, error) {
	if config.Cfg().CustomInfo.LoginByEmail != 0 {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFSingleLoginDisable)
	}
	password, err := decryptPD(login.Password)
	if err != nil {
		return nil, fmt.Errorf("decrypt password err: %v", err)
	}
	resp, err := iam.Login(ctx.Request.Context(), &iam_service.LoginReq{
		UserName: login.Username,
		Password: password,
		Key:      login.Key,
		Code:     login.Code,
		Language: language,
	})
	if err != nil {
		return nil, err
	}
	return getLoginResp(ctx, resp)
}

// 邮箱登录
func LoginByEmail(ctx *gin.Context, login *request.Login) (*response.LoginByEmail, error) {
	if config.Cfg().CustomInfo.LoginByEmail == 0 {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFLoginDisable)
	}
	password, err := decryptPD(login.Password)
	if err != nil {
		return nil, fmt.Errorf("decrypt password err: %v", err)
	}
	resp, err := iam.LoginByEmail(ctx.Request.Context(), &iam_service.LoginByEmailReq{
		UserName: login.Username,
		Password: password,
		Key:      login.Key,
		Code:     login.Code,
	})
	if err != nil {
		return nil, err
	}
	// jwt token
	claims, token, err := jwt_util.GenerateToken(
		resp.UserId,
		jwt_util.UserLoginTokenTimeout,
	)
	if err != nil {
		return nil, err
	}
	ctx.Set(gin_util.CLAIMS, &claims)
	// resp
	return &response.LoginByEmail{
		IsEmailCheck:     resp.IsEmailCheck,
		Token:            token,
		IsUpdatePassword: resp.LastUpdatePasswordAt != 0,
	}, nil
}

// 邮箱检测
func LoginEmailCheck(ctx *gin.Context, login *request.LoginEmailCheck, language, userId string) (*response.Login, error) {
	if config.Cfg().CustomInfo.LoginByEmail == 0 {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFLoginDisable)
	}
	resp, err := iam.LoginEmailCheck(ctx.Request.Context(), &iam_service.LoginEmailCheckReq{
		UserId:   userId,
		Email:    login.Email,
		Code:     login.Code,
		Language: language,
	})
	if err != nil {
		return nil, err
	}
	return getLoginResp(ctx, resp)
}

// 更换密码
func ChangeUserPasswordByEmail(ctx *gin.Context, login *request.ChangeUserPasswordByEmail, language, userId string) (*response.Login, error) {
	if config.Cfg().CustomInfo.LoginByEmail == 0 {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFLoginDisable)
	}
	oldPassword, err := decryptPD(login.OldPassword)
	if err != nil {
		return nil, fmt.Errorf("decrypt password err: %v", err)
	}
	newPassword, err := decryptPD(login.NewPassword)
	if err != nil {
		return nil, fmt.Errorf("decrypt password err: %v", err)
	}
	resp, err := iam.ChangeUserPasswordByEmail(ctx.Request.Context(), &iam_service.ChangeUserPasswordByEmailReq{
		UserId:      userId,
		OldPassword: oldPassword,
		NewPassword: newPassword,
		Email:       login.Email,
		Code:        login.Code,
		Language:    language,
	})
	if err != nil {
		return nil, err
	}
	return getLoginResp(ctx, resp)
}

// --- login email code---
func LoginSendEmailCode(ctx *gin.Context, email string) error {
	if config.Cfg().CustomInfo.LoginByEmail == 0 {
		return grpc_util.ErrorStatus(err_code.Code_BFFLoginDisable)
	}
	_, err := iam.LoginSendEmailCode(ctx.Request.Context(), &iam_service.LoginSendEmailCodeReq{
		Email: email,
	})
	return err
}

func getLoginResp(ctx *gin.Context, resp *iam_service.LoginResp) (*response.Login, error) {
	orgs, err := iam.GetOrgSelect(ctx.Request.Context(), &iam_service.GetOrgSelectReq{UserId: resp.User.GetUserId()})
	if err != nil {
		return nil, err
	}
	return buildLoginResp(ctx, resp.User, resp.Permission, orgs.Selects)
}

func buildLoginResp(ctx *gin.Context, user *iam_service.UserInfo, permission *iam_service.UserPermission, orgs []*iam_service.IDName) (*response.Login, error) {
	// jwt token
	claims, token, err := jwt_util.GenerateToken(
		user.GetUserId(),
		jwt_util.UserTokenTimeout,
	)
	if err != nil {
		return nil, err
	}
	ctx.Set(gin_util.CLAIMS, &claims)
	// resp
	return &response.Login{
		UID:              user.GetUserId(),
		Username:         user.GetUserName(),
		Nickname:         user.GetNickName(),
		Token:            token,
		ExpiresAt:        claims.ExpiresAt * 1000, // 超时事件戳毫秒
		ExpireIn:         strconv.FormatInt(jwt_util.UserTokenTimeout, 10),
		Orgs:             toOrgIDNames(ctx, orgs, user.GetUserId() == config.SystemAdminUserID),
		OrgPermission:    toOrgPermission(ctx, permission),
		Language:         getLanguageByCode(user.Language),
		IsUpdatePassword: permission.LastUpdatePasswordAt != 0,
	}, nil
}

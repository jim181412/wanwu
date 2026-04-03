package v1

import (
	"net/http"

	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// SSOLogin
//
//	@Tags		guest
//	@Summary	统一认证登录跳转
//	@Accept		json
//	@Produce	json
//	@Param		callbackUrl	query		string	false	"登录完成后回跳的前端地址"
//	@Param		mockUser	query		string	false	"mock 模式使用的用户名"
//	@Success	302			{string}	string	"重定向到统一认证中心"
//	@Router		/base/sso/login [get]
func SSOLogin(ctx *gin.Context) {
	var req request.SSOLogin
	if !gin_util.BindQuery(ctx, &req) {
		return
	}

	redirectURL, err := service.LoginBySSO(ctx, req.CallbackURL, req.MockUser)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	ctx.Redirect(http.StatusFound, redirectURL)
}

// SSOExchange
//
//	@Tags		guest
//	@Summary	统一认证票据换取本地登录态
//	@Accept		json
//	@Produce	json
//	@Param		X-Language	header		string	false	"语言"
//	@Param		callbackUrl	query		string	false	"登录完成后回跳的前端地址"
//	@Param		ticket		query		string	true	"CAS 返回票据"
//	@Success	200			{object}	response.Response{data=response.Login}
//	@Router		/base/sso/exchange [get]
func SSOExchange(ctx *gin.Context) {
	var req request.SSOExchange
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.ExchangeSSOLogin(ctx, req.CallbackURL, req.Ticket, getLanguage(ctx))
	gin_util.Response(ctx, resp, err)
}

// SSOLogout
//
//	@Tags		guest
//	@Summary	统一认证退出跳转
//	@Accept		json
//	@Produce	json
//	@Param		callbackUrl	query		string	false	"退出完成后回跳的前端地址"
//	@Success	302			{string}	string	"重定向到统一认证中心退出地址"
//	@Router		/base/sso/logout [get]
func SSOLogout(ctx *gin.Context) {
	var req request.SSOLogout
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	redirectURL, err := service.LogoutBySSO(ctx, req.CallbackURL)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	ctx.Redirect(http.StatusFound, redirectURL)
}

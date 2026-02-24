package v1

import (
	"net/http"

	v1 "github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/v1"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func registerOauth(apiV1 *gin.RouterGroup) {
	mid.Sub("operation.oauth").Reg(apiV1, "/oauth/app", http.MethodPost, v1.CreateOauthApp, "创建OAuth应用")
	mid.Sub("operation.oauth").Reg(apiV1, "/oauth/app", http.MethodDelete, v1.DeleteOauthApp, "删除OAuth应用")
	mid.Sub("operation.oauth").Reg(apiV1, "/oauth/app", http.MethodPut, v1.UpdateOauthApp, "修改OAuth应用信息")
	mid.Sub("operation.oauth").Reg(apiV1, "/oauth/app/list", http.MethodGet, v1.GetOauthAppList, "获取OAuth应用列表")
	mid.Sub("operation.oauth").Reg(apiV1, "/oauth/app/status", http.MethodPut, v1.UpdateOauthAppStatus, "更新OAuth应用状态")
}

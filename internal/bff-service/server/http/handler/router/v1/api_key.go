package v1

import (
	"net/http"

	v1 "github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/v1"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func registerAPIKey(apiV1 *gin.RouterGroup) {
	mid.Sub("api_key.api_key_management").Reg(apiV1, "/api/key", http.MethodPost, v1.CreateAPIKey, "创建API密钥")
	mid.Sub("api_key.api_key_management").Reg(apiV1, "/api/key", http.MethodDelete, v1.DeleteAPIKey, "删除API密钥")
	mid.Sub("api_key.api_key_management").Reg(apiV1, "/api/key/list", http.MethodGet, v1.ListAPIKeys, "获取API密钥列表")
	mid.Sub("api_key.api_key_management").Reg(apiV1, "/api/key", http.MethodPut, v1.UpdateAPIKey, "更新API密钥")
	mid.Sub("api_key.api_key_management").Reg(apiV1, "/api/key/status", http.MethodPut, v1.UpdateAPIKeyStatus, "更新API密钥状态")
}

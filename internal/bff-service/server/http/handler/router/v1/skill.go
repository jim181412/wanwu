package v1

import (
	"net/http"

	v1 "github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/v1"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func registerAgentSkill(apiV1 *gin.RouterGroup) {
	// skills 模板
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/list", http.MethodGet, v1.GetAgentSkillList, "获取skill模板列表")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/detail", http.MethodGet, v1.GetAgentSkillDetail, "获取skill模板详情")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/download", http.MethodGet, v1.DownloadAgentSkill, "下载skill模板")
}

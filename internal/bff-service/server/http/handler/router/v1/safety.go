package v1

import (
	"net/http"

	v1 "github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/v1"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func registerSafety(apiV1 *gin.RouterGroup) {
	// safety
	mid.Sub("resource.safety").Reg(apiV1, "/safe/sensitive/table", http.MethodPost, v1.CreateSensitiveWordTable, "创建敏感词表")
	mid.Sub("resource.safety").Reg(apiV1, "/safe/sensitive/table", http.MethodPut, v1.UpdateSensitiveWordTable, "编辑敏感词表")
	mid.Sub("resource.safety").Reg(apiV1, "/safe/sensitive/table", http.MethodGet, v1.GetSensitiveWordTable, "获取敏感词表")
	mid.Sub("resource.safety").Reg(apiV1, "/safe/sensitive/table/reply", http.MethodPut, v1.UpdateSensitiveWordTableReply, "编辑回复设置")
	mid.Sub("resource.safety").Reg(apiV1, "/safe/sensitive/table", http.MethodDelete, v1.DeleteSensitiveWordTable, "删除敏感词表")
	mid.Sub("resource.safety").Reg(apiV1, "/safe/sensitive/table/list", http.MethodGet, v1.GetSensitiveWordTableList, "获取敏感词表列表")
	mid.Sub("resource.safety").Reg(apiV1, "/safe/sensitive/word/list", http.MethodGet, v1.GetSensitiveVocabularyList, "获取词表数据列表")
	mid.Sub("resource.safety").Reg(apiV1, "/safe/sensitive/word", http.MethodPost, v1.UploadSensitiveVocabulary, "上传敏感词")
	mid.Sub("resource.safety").Reg(apiV1, "/safe/sensitive/word", http.MethodDelete, v1.DeleteSensitiveVocabulary, "删除敏感词")
}

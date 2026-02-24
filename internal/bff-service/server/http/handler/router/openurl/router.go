package openurl

import (
	"net/http"

	"github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/openurl"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func Register(openUrl *gin.RouterGroup) {
	// --- openurl ---
	mid.Sub("openurl").Reg(openUrl, "/agent/:suffix", http.MethodGet, openurl.GetUrlAgentDetail, "获取智能体Url信息")

	// --- conversation ---
	mid.Sub("openurl").Reg(openUrl, "/agent/:suffix/conversation", http.MethodPost, openurl.UrlConversationCreate, "创建智能体对话")
	mid.Sub("openurl").Reg(openUrl, "/agent/:suffix/conversation", http.MethodDelete, openurl.UrlConversationDelete, "删除智能体对话")
	mid.Sub("openurl").Reg(openUrl, "/agent/:suffix/conversation/list", http.MethodGet, openurl.GetUrlConversationList, "获取智能体对话列表")
	mid.Sub("openurl").Reg(openUrl, "/agent/:suffix/conversation/detail", http.MethodGet, openurl.GetUrlConversationDetailList, "智能体对话详情历史列表")
	mid.Sub("openurl").Reg(openUrl, "/agent/:suffix/stream", http.MethodPost, openurl.AssistantUrlConversionStream, "智能体流式问答")
	mid.Sub("openurl").Reg(openUrl, "/agent/:suffix/recommend", http.MethodPost, openurl.AssistantUrlQuestionRecommend, "智能体推荐问题")
}

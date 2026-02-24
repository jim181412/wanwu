package v1

import (
	"net/http"

	v1 "github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/v1"
	"github.com/UnicomAI/wanwu/internal/bff-service/server/http/middleware"
	"github.com/UnicomAI/wanwu/pkg/constant"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func registerExploration(apiV1 *gin.RouterGroup) {
	mid.Sub("exploration.app").Reg(apiV1, "/exploration/app/list", http.MethodGet, v1.GetExplorationAppList, "获取应用广场应用")
	mid.Sub("exploration.app").Reg(apiV1, "/exploration/app/favorite", http.MethodPost, v1.ChangeExplorationAppFavorite, "更改App收藏状态")

	// rag 相关接口
	mid.Sub("exploration.app").Reg(apiV1, "/appspace/rag", http.MethodGet, v1.GetPublishedRag, "获取已发布rag详情")
	mid.Sub("exploration.app").Reg(apiV1, "/rag/chat", http.MethodPost, v1.ChatPublishedRag, "已发布rag流式接口", middleware.AppHistoryRecord("ragId", constant.AppTypeRag))

	// agent 相关接口
	mid.Sub("exploration.app").Reg(apiV1, "/assistant", http.MethodGet, v1.GetPublishedAssistantInfo, "查看已发布智能体详情")
	mid.Sub("exploration.app").Reg(apiV1, "/assistant/conversation", http.MethodPost, v1.ConversationCreate, "创建智能体对话")
	mid.Sub("exploration.app").Reg(apiV1, "/assistant/conversation", http.MethodDelete, v1.ConversationDelete, "删除智能体对话")
	mid.Sub("exploration.app").Reg(apiV1, "/assistant/conversation/list", http.MethodGet, v1.GetConversationList, "智能体对话列表")
	mid.Sub("exploration.app").Reg(apiV1, "/assistant/conversation/detail", http.MethodGet, v1.GetConversationDetailList, "智能体对话详情历史列表")
	mid.Sub("exploration.app").Reg(apiV1, "/assistant/stream", http.MethodPost, v1.PublishedAssistantConversionStream, "已发布智能体流式问答", middleware.AppHistoryRecord("assistantId", constant.AppTypeAgent))

	// workflow 相关接口
	mid.Sub("exploration.app").Reg(apiV1, "/workflow/run", http.MethodPost, v1.PublishedWorkflowRun, "已发布工作流运行接口", middleware.AppHistoryRecord("workflow_id", constant.AppTypeWorkflow))
	mid.Sub("exploration.app").Reg(apiV1, "/appspace/workflow/export", http.MethodGet, v1.ExportWorkflow, "导出workflow")
	mid.Sub("exploration.app").Reg(apiV1, "/appspace/workflow/copy", http.MethodPost, v1.CopyWorkflow, "拷贝workflow")

	// chatflow 相关接口
	mid.Sub("exploration.app").Reg(apiV1, "/chatflow/application/list", http.MethodPost, v1.ChatflowApplicationList, "应用广场对话流关联应用", middleware.AppHistoryRecord("workflow_id", constant.AppTypeChatflow))
	mid.Sub("exploration.app").Reg(apiV1, "/chatflow/application/info", http.MethodPost, v1.ChatflowApplicationInfo, "应用广场对话流关联应用信息")
	mid.Sub("exploration.app").Reg(apiV1, "/chatflow/conversation/delete", http.MethodDelete, v1.DeleteChatflowConversation, "删除对话流会话")
	mid.Sub("exploration.app").Reg(apiV1, "/appspace/chatflow/copy", http.MethodPost, v1.CopyChatflow, "拷贝chatflow")

	mid.Sub("exploration.template").Reg(apiV1, "/prompt/template/list", http.MethodGet, v1.GetPromptTemplateList, "获取提示词模板列表")
	mid.Sub("exploration.template").Reg(apiV1, "/prompt/template/detail", http.MethodGet, v1.GetPromptTemplateDetail, "获取提示词模板详情")
}

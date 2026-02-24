package openapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	"github.com/UnicomAI/wanwu/pkg/constant"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	sse_util "github.com/UnicomAI/wanwu/pkg/sse-util"
	"github.com/gin-gonic/gin"
)

//	@title		AI Agent Productivity Platform - Open API
//	@version	v0.0.1

//	@BasePath	/openapi/v1

// CreateAgentConversation
//
//	@Tags			openapi
//	@Summary		智能体创建对话OpenAPI
//	@Description	智能体创建对话OpenAPI
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.OpenAPIAgentCreateConversationRequest	true	"请求参数"
//	@Success		400		{object}	response.Response{data=response.OpenAPIAgentCreateConversationResponse}
//	@Router			/agent/conversation [post]
func CreateAgentConversation(ctx *gin.Context) {
	var req request.OpenAPIAgentCreateConversationRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	userID := getUserID(ctx)
	orgID := getOrgID(ctx)
	appID, err := service.GetAssistantIdByUuid(ctx, req.UUID)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	resp, err := service.ConversationCreate(ctx, userID, orgID, request.ConversationCreateRequest{
		AssistantId: appID,
		Prompt:      req.Title,
	}, constant.ConversationTypeOpenAPI)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	gin_util.Response(ctx, response.OpenAPIAgentCreateConversationResponse{ConversationID: resp.ConversationId}, nil)
}

// ChatAgent
//
//	@Tags			openapi
//	@Summary		智能体对话OpenAPI
//	@Description	智能体对话OpenAPI
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.OpenAPIAgentChatRequest	true	"请求参数"
//	@Success		200		{object}	response.OpenAPIAgentChatResponse
//	@Success		400		{object}	response.Response
//	@Router			/agent/chat [post]
func ChatAgent(ctx *gin.Context) {
	var req request.OpenAPIAgentChatRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	userID := getUserID(ctx)
	orgID := getOrgID(ctx)
	appID, err := service.GetAssistantIdByUuid(ctx, req.UUID)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	// 流式
	if req.Stream {
		if err := service.AssistantConversionStream(ctx, userID, orgID, request.ConversionStreamRequest{
			AssistantId:    appID,
			ConversationId: req.ConversationID,
			Prompt:         req.Query,
			FileInfo:       []request.ConversionStreamFile{},
		}, true); err != nil {
			gin_util.Response(ctx, nil, err)
		}
		return
	}
	// 非流式
	chatCh, err := service.CallAssistantConversationStream(ctx, userID, orgID, request.ConversionStreamRequest{
		AssistantId:    appID,
		ConversationId: req.ConversationID,
		Prompt:         req.Query,
		FileInfo:       []request.ConversionStreamFile{},
	}, true)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	var output string
	resp := &response.OpenAPIAgentChatResponse{}
	for chat := range chatCh {
		// 注意这里智能体的原始流式返回没有data:前缀
		if strings.TrimSpace(chat) == "" {
			continue
		}
		curr := &response.OpenAPIAgentChatResponse{}
		if err := json.Unmarshal([]byte(strings.TrimPrefix(chat, "data:")), curr); err != nil {
			log.Errorf("[Agent] %v conversation %v user %v org %v unmarshal %v err: %v", appID, req.ConversationID, userID, orgID, err)
			continue
		}
		resp = curr
		output += curr.Response
	}
	resp.Response = output
	b, _ := json.Marshal(resp)
	status := http.StatusOK
	ctx.Set(gin_util.STATUS, status)
	ctx.Set(gin_util.RESULT, string(b))
	ctx.JSON(status, resp)
}

// ChatRag
//
//	@Tags			openapi
//	@Summary		文本问答OpenAPI
//	@Description	文本问答OpenAPI
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.OpenAPIRagChatRequest	true	"请求参数"
//	@Success		200		{object}	response.OpenAPIRagChatResponse
//	@Success		400		{object}	response.Response
//	@Router			/rag/chat [post]
func ChatRag(ctx *gin.Context) {
	var req request.OpenAPIRagChatRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	userID := getUserID(ctx)
	orgID := getOrgID(ctx)

	// 流式
	if req.Stream {
		if err := service.ChatRagStream(ctx, userID, orgID, request.ChatRagRequest{RagID: req.UUID, Question: req.Query, History: req.History}, true); err != nil {
			gin_util.Response(ctx, nil, err)
		}
		return
	}
	// 非流式
	chatCh, err := service.CallRagChatStream(ctx, userID, orgID, request.ChatRagRequest{RagID: req.UUID, Question: req.Query, History: req.History}, true)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	var output string
	resp := &response.OpenAPIRagChatResponse{}
	for chat := range chatCh {
		if !strings.HasPrefix(chat, "data:") || strings.HasPrefix(chat, strings.TrimSpace(sse_util.DONE_MSG)) {
			continue
		}
		curr := &response.OpenAPIRagChatResponse{}
		if err := json.Unmarshal([]byte(strings.TrimPrefix(chat, "data:")), curr); err != nil {
			log.Errorf("[RAG] %v user %v org %v unmarshal %v err: %v", req.UUID, userID, orgID, err)
			continue
		}
		resp = curr
		output += curr.Data.Output
	}
	resp.Data.Output = output
	b, _ := json.Marshal(resp)
	status := http.StatusOK
	ctx.Set(gin_util.STATUS, status)
	ctx.Set(gin_util.RESULT, string(b))
	ctx.JSON(status, resp)
}

// --- internal ---

// 获取当前用户ID
func getUserID(ctx *gin.Context) string {
	return ctx.GetString(gin_util.USER_ID)
}

// 获取当前组织ID
func getOrgID(ctx *gin.Context) string {
	return ctx.GetString(gin_util.X_ORG_ID)
}

// 获取当前appID
func getAppID(ctx *gin.Context) string {
	return ctx.GetString(gin_util.APP_ID)
}

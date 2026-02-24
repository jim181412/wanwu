package local_agent

import (
	"context"

	"github.com/UnicomAI/wanwu/internal/agent-service/model/request"
	service_model "github.com/UnicomAI/wanwu/internal/agent-service/service/service-model"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
)

type LocalAgentService interface {
	//CreateChatModel 创建chatModel
	CreateChatModel(ctx context.Context, req *request.AgentChatParams, agentChatInfo *service_model.AgentChatInfo) (model.ToolCallingChatModel, error)
	//BuildAgentInput 构造会话消息
	BuildAgentInput(ctx context.Context, req *request.AgentChatParams, agentChatInfo *service_model.AgentChatInfo, agentInput *adk.AgentInput, generator *adk.AsyncGenerator[*adk.AgentEvent]) (*adk.AgentInput, error)
}

func CreateLocalAgentService(ctx context.Context, req *request.AgentChatParams, agentChatInfo *service_model.AgentChatInfo, chatContext *request.AgentChatContext) LocalAgentService {
	////如果有特殊输出或者逻辑的模型可以仿照，vision_chat 实现，不过目前主流的的vision_chat 都是openai格式,ChatAgent都可兼容
	//if agentChatInfo.VisionSupport {
	//	return &VisionChatAgent{}
	//}
	return &ChatAgent{ChatContext: chatContext}
}

func CreateChatModel(ctx context.Context, agentChatInfo *service_model.AgentChatInfo, req *request.AgentChatParams) (model.ToolCallingChatModel, error) {
	modelInfo := agentChatInfo.ModelInfo
	modelConfig := modelInfo.Config
	params := req.ModelParams
	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:           modelConfig.ApiKey,
		BaseURL:          modelConfig.EndpointUrl,
		Model:            modelInfo.Model,
		Temperature:      params.Temperature,
		TopP:             params.TopP,
		FrequencyPenalty: params.FrequencyPenalty,
		PresencePenalty:  params.PresencePenalty,
	})
}

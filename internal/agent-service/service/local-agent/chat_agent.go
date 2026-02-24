package local_agent

import (
	"context"

	"github.com/UnicomAI/wanwu/internal/agent-service/model/request"
	"github.com/UnicomAI/wanwu/internal/agent-service/pkg/config"
	agent_message_flow "github.com/UnicomAI/wanwu/internal/agent-service/service/agent-message-flow"
	service_model "github.com/UnicomAI/wanwu/internal/agent-service/service/service-model"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type ChatAgent struct {
	ChatContext *request.AgentChatContext
}

func (a *ChatAgent) CreateChatModel(ctx context.Context, req *request.AgentChatParams, agentChatInfo *service_model.AgentChatInfo) (model.ToolCallingChatModel, error) {
	fillInternalToolConfig(req, agentChatInfo)
	return CreateChatModel(ctx, agentChatInfo, req)
}

// BuildAgentInput 构造会话消息
func (a *ChatAgent) BuildAgentInput(ctx context.Context, req *request.AgentChatParams, agentChatInfo *service_model.AgentChatInfo, agentInput *adk.AgentInput, generator *adk.AsyncGenerator[*adk.AgentEvent]) (*adk.AgentInput, error) {
	userInput, messages := splitUserInput(req, agentInput.Messages)
	req.Input = userInput

	agentChatContext := &request.AgentChatContext{AgentChatReq: req, AgentChatInfo: agentChatInfo, Generator: generator}
	//1.创建前置消息准备
	messageBuilder, err := createMessageBuilder(ctx, agentChatContext)
	if err != nil {
		return nil, err
	}
	//2.生成前置消息
	createMessages, err := messageBuilder.Invoke(ctx, agentChatContext)
	if err != nil {
		return nil, err
	}
	createMessages = append(createMessages, messages...)
	//3.知识库信息记录
	if a.ChatContext != nil {
		a.ChatContext.KnowledgeHitData = agentChatContext.KnowledgeHitData
	}
	return &adk.AgentInput{
		Messages:        createMessages,
		EnableStreaming: agentInput.EnableStreaming,
	}, nil
}

func splitUserInput(req *request.AgentChatParams, messages []*schema.Message) (string, []*schema.Message) {
	if len(messages) > 0 {
		var retMessages []*schema.Message
		var userInput string
		var gotUserMessage = false
		for _, message := range messages {
			if !gotUserMessage && message.Role == schema.User && len(message.Content) > 0 {
				userInput = message.Content
				gotUserMessage = true
			} else {
				retMessages = append(retMessages, message)
			}
		}
		return userInput, retMessages
	}
	return req.Input, messages
}
func createMessageBuilder(ctx context.Context, req *request.AgentChatContext) (compose.Runnable[*request.AgentChatContext, []*schema.Message], error) {
	graph := agent_message_flow.NewAgentMessageFlow(req.AgentChatReq.MultiAgent)
	return graph.Compile(ctx)
}

// fillInternalToolConfig 配置内置文件工具
func fillInternalToolConfig(req *request.AgentChatParams, agentChatInfo *service_model.AgentChatInfo) {
	if !agentChatInfo.FunctionCalling {
		req.ToolParams = nil
	}
	//如果用户使用的不是多模态模型，但是又上传了文件，则通过工具对文件进行解析，
	//但是目前只支持一个文件，具体处理逻在node_prompt_variables.go
	if !agentChatInfo.VisionSupport {
		templateConfig := config.GetToolTemplateConfig()
		if len(templateConfig.ConfigPluginToolList) > 0 && agentChatInfo.UploadUrl {
			params := req.ToolParams
			if params != nil {
				params.PluginToolList = append(params.PluginToolList, templateConfig.ConfigPluginToolList...)
			} else {
				params = &request.ToolParams{
					PluginToolList: templateConfig.ConfigPluginToolList,
				}
			}
			req.ToolParams = params
		}
	}
}

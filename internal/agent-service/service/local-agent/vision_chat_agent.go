package local_agent

import (
	"context"
	"path/filepath"

	"github.com/UnicomAI/wanwu/internal/agent-service/model/request"
	"github.com/UnicomAI/wanwu/internal/agent-service/pkg/config"
	agent_util "github.com/UnicomAI/wanwu/internal/agent-service/pkg/util"
	chat_model "github.com/UnicomAI/wanwu/internal/agent-service/service/agent-message-flow/chat-model"
	minio_service "github.com/UnicomAI/wanwu/internal/agent-service/service/minio-service"
	service_model "github.com/UnicomAI/wanwu/internal/agent-service/service/service-model"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type VisionChatAgent struct {
}

func (a *VisionChatAgent) CreateChatModel(ctx context.Context, req *request.AgentChatParams, agentChatInfo *service_model.AgentChatInfo) (model.ToolCallingChatModel, error) {
	req.KnowledgeParams = nil
	req.ToolParams = nil

	//1.创建chatModel
	info := agentChatInfo.ModelInfo
	chatModel := chat_model.CreateYuanjingVLModel(info.Model, info.Config.ApiKey, info.Config.EndpointUrl)
	return chatModel, nil
}

// BuildAgentInput 构造会话消息
func (a *VisionChatAgent) BuildAgentInput(ctx context.Context, req *request.AgentChatParams, agentChatInfo *service_model.AgentChatInfo, agentInput *adk.AgentInput, generator *adk.AsyncGenerator[*adk.AgentEvent]) (*adk.AgentInput, error) {
	var messages []*schema.Message
	messages = agentInput.Messages
	if agentChatInfo.UploadUrl {
		var parts []schema.MessageInputPart
		for _, minioFilePath := range req.UploadFile {
			message, err := buildFileMessage(ctx, minioFilePath)
			if err != nil {
				return nil, err
			}
			parts = append(parts, *message)
		}
		messages = append(messages, &schema.Message{
			Role:                  schema.User,
			UserInputMultiContent: parts,
		})
	}
	return &adk.AgentInput{
		Messages:        messages,
		EnableStreaming: req.Stream,
	}, nil
}

// buildFileMessage 构建文件消息
func buildFileMessage(ctx context.Context, minioFilePath string) (*schema.MessageInputPart, error) {
	//1.下载压缩文件到本地
	var localFilePath = agent_util.BuildFilePath(config.GetConfig().AgentFileConfig.LocalFilePath, filepath.Ext(minioFilePath))
	err := minio_service.DownloadFileToLocal(ctx, minioFilePath, localFilePath)
	if err != nil {
		return nil, err
	}
	//2.图片转base64
	base64, err := agent_util.Img2base64(localFilePath)
	if err != nil {
		return nil, err
	}
	return &schema.MessageInputPart{
		Type: schema.ChatMessagePartTypeImageURL,
		Image: &schema.MessageInputImage{
			MessagePartCommon: schema.MessagePartCommon{
				Base64Data: &base64,
			},
		},
	}, nil
}

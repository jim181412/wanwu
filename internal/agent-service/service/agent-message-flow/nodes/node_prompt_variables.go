package nodes

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/UnicomAI/wanwu/internal/agent-service/model/request"
	"github.com/UnicomAI/wanwu/internal/agent-service/pkg/config"
	agent_util "github.com/UnicomAI/wanwu/internal/agent-service/pkg/util"
	"github.com/UnicomAI/wanwu/internal/agent-service/service/agent-message-flow/prompt"
	minio_service "github.com/UnicomAI/wanwu/internal/agent-service/service/minio-service"
	"github.com/cloudwego/eino/schema"
)

const (
	placeholderOfUserInput   = "_user_input"
	placeholderOfChatHistory = "_chat_history"
)

type PromptVariables struct {
	Avs map[string]string
}

func (p *PromptVariables) AssemblePromptVariables(ctx context.Context, reqContext *request.AgentChatContext) (variables map[string]any, err error) {
	req := reqContext.AgentChatReq
	variables = make(map[string]any)

	variables[prompt.PlaceholderOfTime] = time.Now().Format("Monday 2006/01/02 15:04:05 -07")
	variables[prompt.PlaceholderOfAgentName] = req.AgentBaseParams.Name

	input, err := buildUserInput(reqContext)
	if err != nil {
		return nil, err
	}
	variables[placeholderOfUserInput] = input

	// Handling conversation history
	if len(req.ModelParams.History) > 0 {
		// Add chat history to variable
		variables[placeholderOfChatHistory] = buildHistory(req.ModelParams.History, req.ModelParams.MaxHistory)
	}

	if p.Avs != nil {
		var memoryVariablesList []string
		for k, v := range p.Avs {
			variables[k] = v
			memoryVariablesList = append(memoryVariablesList, fmt.Sprintf("%s: %s\n", k, v))
		}
		variables[prompt.PlaceholderOfVariables] = memoryVariablesList
	}

	subAgentInfoList := reqContext.AgentChatReq.SubAgentInfoList
	if reqContext.AgentChatReq.MultiAgent && len(subAgentInfoList) > 0 {
		variables[prompt.PlaceholderOfSubAgentCount] = strconv.Itoa(len(subAgentInfoList))
	}

	return variables, nil
}

func buildHistory(history []request.AssistantConversionHistory, maxHistory int) []*schema.Message {
	var historyList []*schema.Message

	// 处理所有历史记录
	for _, conversionHistory := range history {
		historyList = append(historyList, schema.UserMessage(conversionHistory.Query))
		//todo 先不传ToolCall(后续版本考虑传进去)
		historyList = append(historyList, schema.AssistantMessage(conversionHistory.Response, nil))
	}
	if maxHistory <= 0 {
		return historyList
	}
	// 每条记录占用2个位置(问/答)
	maxHistory = maxHistory * 2
	// 只返回最后maxHistory条
	if len(historyList) > maxHistory {
		return historyList[len(historyList)-maxHistory:]
	}
	return historyList
}

func buildUserInput(reqContext *request.AgentChatContext) ([]*schema.Message, error) {
	req := reqContext.AgentChatReq
	agentChatInfo := reqContext.AgentChatInfo
	var input = req.Input

	var messages []*schema.Message
	if agentChatInfo.VisionSupport && agentChatInfo.UploadUrl { // 视觉模型，传了url
		var parts []schema.MessageInputPart
		for _, minioFilePath := range req.UploadFile {
			message, err := buildFileMessage(minioFilePath)
			if err != nil {
				return nil, err
			}
			parts = append(parts, *message)
		}
		parts = append(parts, schema.MessageInputPart{
			Type: schema.ChatMessagePartTypeText,
			Text: req.Input,
		})
		messages = append(messages, &schema.Message{
			Role:                  schema.User,
			UserInputMultiContent: parts,
		})
	} else if agentChatInfo.UploadUrl { //非视觉模型，传了url
		input += "\n用户上传的文档连接为:" + req.UploadFile[0]
		messages = append(messages, schema.UserMessage(input))
	} else {
		messages = append(messages, schema.UserMessage(input))
	}
	return messages, nil
}

// buildFileMessage 构建文件消息
func buildFileMessage(minioFilePath string) (*schema.MessageInputPart, error) {
	//1.下载压缩文件到本地
	var localFilePath = agent_util.BuildFilePath(config.GetConfig().AgentFileConfig.LocalFilePath, filepath.Ext(minioFilePath))
	err := minio_service.DownloadFileToLocal(context.Background(), minioFilePath, localFilePath)
	if err != nil {
		return nil, err
	}
	//2.图片转base64
	mimeType, base64, err := agent_util.Img2base64Data(localFilePath)
	if err != nil {
		return nil, err
	}
	return &schema.MessageInputPart{
		Type: schema.ChatMessagePartTypeImageURL,
		Image: &schema.MessageInputImage{
			MessagePartCommon: schema.MessagePartCommon{
				Base64Data: &base64,
				MIMEType:   mimeType,
			},
		},
	}, nil
}

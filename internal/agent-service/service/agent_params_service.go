package service

import (
	"encoding/json"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/internal/agent-service/model/request"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/samber/lo"
)

// BuildMultiAgentParams 构建多智能体问答请求
func BuildMultiAgentParams(multiAgentChatParams *request.MultiAgentChatParams, multiAgentConfig *assistant_service.MultiAssistantDetailResp) *request.MultiAgentChatReq {
	return &request.MultiAgentChatReq{
		Input:           multiAgentChatParams.Input,
		UploadFile:      multiAgentChatParams.UploadFile,
		Stream:          multiAgentChatParams.Stream,
		ModelParams:     buildModelParams(multiAgentConfig.MultiAgent.ModelParams),
		AgentBaseParams: buildAgentBaseParams(multiAgentConfig.MultiAgent),
		AgentList:       buildSubAgentParamsList(multiAgentConfig.SubAgents),
	}
}

func buildAgentBaseParams(assistantDetail *assistant_service.AgentDetail) *request.AgentBaseParams {
	baseParams := assistantDetail.AgentBaseParams
	return &request.AgentBaseParams{
		AgentId:     baseParams.AgentId,
		Description: baseParams.Description,
		Instruction: baseParams.Instruction,
		Name:        baseParams.Name,
		Avatar:      baseParams.Avatar,
		CallDetail:  true,
	}
}

func buildSubAgentParamsList(subAgents []*assistant_service.AgentDetail) []*request.AgentChatBaseParams {
	return lo.Map(subAgents, func(item *assistant_service.AgentDetail, index int) *request.AgentChatBaseParams {
		return buildAgentChatBaseParams(item)
	})
}

func BuildAgentParams(req *request.AgentChatReq, assistantDetail *assistant_service.AssistantDetailResp) *request.AgentChatParams {
	params := buildAgentChatBaseParams(assistantDetail.GetAgentDetail())
	return &request.AgentChatParams{
		AgentChatBaseParams: *params,
		Input:               req.Input,
		Stream:              req.Stream,
		UploadFile:          req.UploadFile,
	}
}

func buildAgentChatBaseParams(assistantDetail *assistant_service.AgentDetail) *request.AgentChatBaseParams {
	return &request.AgentChatBaseParams{
		AgentBaseParams: &request.AgentBaseParams{
			Description: assistantDetail.AgentBaseParams.Description,
			Instruction: assistantDetail.AgentBaseParams.Instruction,
			Name:        assistantDetail.AgentBaseParams.Name,
		},
		KnowledgeParams: buildKnowledgeParams(assistantDetail.KnowledgeParams),
		ModelParams:     buildModelParams(assistantDetail.ModelParams),
		ToolParams:      buildToolParams(assistantDetail.ToolParams),
	}
}

func buildKnowledgeParams(knowledgeParams string) *request.KnowledgeParams {
	if knowledgeParams == "" {
		return nil
	}
	var params = &request.KnowledgeParams{}
	err := json.Unmarshal([]byte(knowledgeParams), &params)
	if err != nil {
		log.Errorf("buildAgentParams knowledgeParams %s err %s", knowledgeParams, err)
		return nil
	}
	return params
}

// buildModelParams 构造模型参数
func buildModelParams(req *assistant_service.ModelParams) *request.ModelParams {
	output := &request.ModelParams{
		ModelId:    req.ModelId,
		MaxHistory: int(req.MaxHistory),
		History:    buildHistory(req.History),
	}

	// 转换 float64 到 float32
	if req.Temperature != nil {
		temp := float32(*req.Temperature)
		output.Temperature = &temp
	}

	if req.TopP != nil {
		topP := float32(*req.TopP)
		output.TopP = &topP
	}

	if req.FrequencyPenalty != nil {
		freqPenalty := float32(*req.FrequencyPenalty)
		output.FrequencyPenalty = &freqPenalty
	}

	if req.PresencePenalty != nil {
		presencePenalty := float32(*req.PresencePenalty)
		output.PresencePenalty = &presencePenalty
	}

	// 转换 int32 到 int
	if req.MaxTokens != nil {
		maxTokens := int(*req.MaxTokens)
		output.MaxTokens = &maxTokens
	}

	return output
}

func buildToolParams(toolParams *assistant_service.ToolParams) *request.ToolParams {
	return &request.ToolParams{
		McpToolList:    buildMCPToolList(toolParams.McpToolList),
		PluginToolList: buildPluginToolList(toolParams.PluginToolList),
	}
}
func buildMCPToolList(mcpToolList []*assistant_service.MCPToolInfo) []*request.MCPToolInfo {
	if len(mcpToolList) == 0 {
		return nil
	}
	return lo.Map(mcpToolList, func(item *assistant_service.MCPToolInfo, index int) *request.MCPToolInfo {
		return &request.MCPToolInfo{
			URL:          item.Url,
			Transport:    item.Transport,
			ToolNameList: item.ToolNameList,
		}
	})
}

func buildPluginToolList(pluginTool []string) []*request.PluginToolInfo {
	if len(pluginTool) == 0 {
		return nil
	}
	var pluginToolList []*request.PluginToolInfo
	for _, tool := range pluginTool {
		var toolInfo = &request.PluginToolInfo{}
		err := json.Unmarshal([]byte(tool), &toolInfo)
		if err != nil {
			log.Errorf("buildPluginToolList pluginTool %s err %s", tool, err)
			continue
		}
		pluginToolList = append(pluginToolList, toolInfo)
	}
	return pluginToolList
}

func buildHistory(history []*assistant_service.ConversionHistory) []request.AssistantConversionHistory {
	if len(history) == 0 {
		return nil
	}
	return lo.Map(history, func(item *assistant_service.ConversionHistory, index int) request.AssistantConversionHistory {
		return request.AssistantConversionHistory{
			Query:         item.Query,
			Response:      item.Response,
			UploadFileUrl: item.UploadFileUrl,
		}
	})
}

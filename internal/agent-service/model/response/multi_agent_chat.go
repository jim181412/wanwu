package response

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/UnicomAI/wanwu/internal/agent-service/model/request"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)

const (
	AgentStartLabel    = "transfer_to_agent"
	defaultAgentAvatar = "/v1/static/icon/agent-default-icon.png"
)

type AgentText struct {
	AgentName string
	Text      string
}

type AgentInfo struct {
	AgentName string `json:"agent_name"`
}

// MultiNewAgentChatRespWithTool 多智能体涉及各子智能体展示切换，而且在最后很可能不输出finish_reason = stop 而是输出以tool==exit，如果现有的工具名==exit 可能有bug,todo 输出方法考虑重构
func MultiNewAgentChatRespWithTool(chatMessage *schema.Message, respContext *AgentChatRespContext, req *request.AgentChatContext) ([]string, error) {
	contentList, subAgentEventData, notStop := buildMultiContentWithTool(chatMessage, respContext, req)
	var outputList = make([]string, 0)
	if len(contentList) == 0 && subAgentEventData != nil {
		return buildSubAgentEventInfo(req, chatMessage, subAgentEventData)
	}
	for _, content := range contentList {
		var agentChatResp = &AgentChatResp{
			Code:           agentSuccessCode,
			Message:        "success",
			Response:       content,
			EventType:      buildEventType(subAgentEventData),
			EventData:      subAgentEventData,
			GenFileUrlList: []interface{}{},
			History:        []interface{}{},
			QaType:         buildQaType(req),
			SearchList:     buildSearchList(req),
			Finish:         buildFinish(chatMessage, notStop),
			Usage:          buildUsage(chatMessage),
		}
		respString, err := buildRespString(agentChatResp)
		if err != nil {
			return nil, err
		}
		outputList = append(outputList, respString)
	}
	return outputList, nil
}

func buildSubAgentEventInfo(respContext *request.AgentChatContext, chatMessage *schema.Message, subAgentEventData *SubEventData) ([]string, error) {
	var outputList = make([]string, 0)
	var agentChatResp = &AgentChatResp{
		Code:           agentSuccessCode,
		Message:        "success",
		Response:       "",
		EventType:      buildEventType(subAgentEventData),
		EventData:      subAgentEventData,
		GenFileUrlList: []interface{}{},
		History:        []interface{}{},
		SearchList:     buildSubAgentSearchList(subAgentEventData, respContext),
		Finish:         buildFinish(chatMessage, true),
		Usage:          buildUsage(chatMessage),
	}
	respString, err := buildRespString(agentChatResp)
	if err != nil {
		return nil, err
	}
	outputList = append(outputList, respString)
	return outputList, nil
}

func buildMultiContentWithTool(chatMessage *schema.Message, respContext *AgentChatRespContext, req *request.AgentChatContext) ([]string, *SubEventData, bool) {
	if filterMultiMessage(chatMessage) {
		return make([]string, 0), nil, false
	}
	first, start := agentStart(chatMessage, respContext)
	var contentList = make([]string, 0)
	//子智能体开始
	if start {
		if first {
			return contentList, nil, false
		}
		if len(chatMessage.ToolCalls) > 0 {
			toolCall := chatMessage.ToolCalls[0]
			if len(toolCall.Function.Arguments) > 0 {
				respContext.AgentTempMessage.WriteString(toolCall.Function.Arguments)
				if !toolParamsEnd(chatMessage) {
					return contentList, nil, false
				}

			}
		}
		if toolParamsEnd(chatMessage) {
			respContext.AgentStart = false
			agentName := buildAgentName(respContext.AgentTempMessage.String())
			respContext.AgentTempMessage = strings.Builder{}

			respContext.CurrentAgent = agentName
			respContext.CurrentAgentId = uuid.New().String()
			respContext.CurrentAgentAvatar = buildAgentAvatar(agentName, req)
			return contentList, BuildStartSubAgent(respContext), false
		}
	}
	//子智能体结束
	if chatMessage.ResponseMeta != nil && chatMessage.ResponseMeta.FinishReason == "stop" && len(respContext.CurrentAgent) > 0 {
		subAgentEventData := BuildEndSubAgent(respContext, util.NowSpanToHMS(respContext.AgentStartTime))
		respContext.CurrentAgentId = ""
		respContext.CurrentAgent = ""
		return contentList, subAgentEventData, true
	}

	//supervisor 结束
	if chatMessage.Role == schema.Tool && chatMessage.ToolName == "exit" {
		respContext.ExitTool = false
		return []string{chatMessage.Content}, nil, false
	}
	if exitToolStart(chatMessage, respContext) {
		return make([]string, 0), nil, false
	}
	if toolStart(chatMessage) {
		respContext.ToolStart = true
		respContext.HasTool = true
		var retList []string

		for _, tool := range chatMessage.ToolCalls {
			newTool := isNewTool(tool, respContext)
			if !newTool && len(tool.Function.Arguments) > 0 { //只有此次不是新工具才add 参数，处理先输出方法名，后流式输出参数的情况
				retList = append(retList, tool.Function.Arguments)
				continue
			}
			if tool.Type == "function" {
				if respContext.ToolIndex == -1 {
					respContext.ToolIndex = *tool.Index
				} else if *tool.Index != respContext.ToolIndex { //模型触发并发请求工具的bad case
					respContext.ToolIndex = *tool.Index
					if respContext.ToolParamsStartCount > respContext.ToolParamsEndCount {
						respContext.ToolParamsEndCount = respContext.ToolParamsEndCount + 1
						retList = append(retList, toolParamsEndFormat)
					}
				}
				if len(tool.Function.Name) > 0 {
					toolName := fmt.Sprintf(toolStartTitleFormat, tool.Function.Name)
					retList = append(retList, toolName)
				}

				if newTool {
					respContext.ToolParamsStartCount = respContext.ToolParamsStartCount + 1
					retList = append(retList, toolStartTitle)
					retList = append(retList, toolParamsStartFormat)
					respContext.ToolCountMap[tool.ID] = 1
					if len(tool.Function.Arguments) > 0 { //处理一次性同时返回 方法名和参数的情况
						retList = append(retList, tool.Function.Arguments)
					}
					if toolParamsEnd(chatMessage) { //处理一次性同时返回 方法名和参数的情况 ,同时返回结束了
						retList = append(retList, toolParamsEndFormat)
					}
				}
			}
		}

		return retList, BuildProcessSubAgent(respContext), false
	} else if toolParamsEnd(chatMessage) {
		respContext.ToolParamsEndCount = respContext.ToolParamsEndCount + 1
		return []string{toolParamsEndFormat}, BuildProcessSubAgent(respContext), false
	} else if toolEnd(chatMessage) {
		respContext.ToolEnd = true
		toolResult := fmt.Sprintf(toolEndFormat, chatMessage.ToolName, chatMessage.Content)
		respContext.ToolCountMap[chatMessage.ToolCallID] = 0
		return []string{toolResult, toolEndTitle}, BuildProcessSubAgent(respContext), false
	} else {
		//在工具期间，不输出任何content内容
		if respContext.ToolStart && !respContext.ToolEnd {
			return []string{}, BuildProcessSubAgent(respContext), false
		}
		//替换内容准备(工具未开始，但是输出了内容)
		if !respContext.ToolStart {
			if utf8.RuneCountInString(chatMessage.Content) > 10 {
				var replaceContent = respContext.ReplaceContentStr
				if len(replaceContent) == 0 {
					replaceContent = respContext.ReplaceContent.String()
				}
				if replaceContent == chatMessage.Content {
					respContext.ReplaceContentDone = true
					respContext.ReplaceContentStr = replaceContent
					return []string{}, BuildProcessSubAgent(respContext), false
				}
			}
			if !respContext.ReplaceContentDone {
				respContext.ReplaceContent.WriteString(chatMessage.Content)
			}

		}
		return []string{chatMessage.Content}, BuildProcessSubAgent(respContext), false
	}
}

func buildAgentAvatar(agentName string, req *request.AgentChatContext) string {
	if len(req.SubAgentMap) == 0 {
		return defaultAgentAvatar
	}
	agentConfig := req.SubAgentMap[agentName]
	if agentConfig == nil || len(agentConfig.AgentAvatar) == 0 {
		return defaultAgentAvatar
	}
	return agentConfig.AgentAvatar
}

func exitToolStart(chatMessage *schema.Message, respContext *AgentChatRespContext) bool {
	if respContext.ExitTool {
		return true
	}
	if len(chatMessage.ToolCalls) > 0 {
		toolCall := chatMessage.ToolCalls[0]
		if toolCall.Function.Name == "exit" {
			respContext.ExitTool = true
			return true
		}
	}
	return false
}

// agentStart 子智能体开始
func agentStart(chatMessage *schema.Message, respContext *AgentChatRespContext) (first bool, start bool) {
	if respContext.AgentStart {

		if !subAgentParamsStart(chatMessage) {
			return false, true

		}
		//如果在开始过程中，模型抽风又触发开始，先忽略之前得清空
		respContext.AgentTempMessage.Reset()
	}
	if AgentStartLabel == chatMessage.ToolName {
		return true, true
	}
	if len(chatMessage.ToolCalls) == 0 {
		return false, false
	}
	toolCall := chatMessage.ToolCalls[0]
	if AgentStartLabel == toolCall.Function.Name {
		if toolCall.Function.Arguments != respContext.MainAgentName {
			respContext.AgentStart = true
			respContext.AgentStartTime = time.Now().UnixMilli()
			agentName := buildAgentName(toolCall.Function.Arguments)
			if len(agentName) > 0 {
				return false, true
			}
		}
		return true, true
	}
	return false, false
}

func buildAgentName(tempMessage string) string {
	if len(tempMessage) == 0 {
		return ""
	}
	if !json.Valid([]byte(tempMessage)) {
		return ""
	}
	var agentInfo = &AgentInfo{}
	_ = json.Unmarshal([]byte(tempMessage), agentInfo)
	return agentInfo.AgentName
}

// 子智能体参数开始
func subAgentParamsStart(chatMessage *schema.Message) bool {
	if len(chatMessage.ToolCalls) == 0 {
		return false
	}
	toolCall := chatMessage.ToolCalls[0]
	return AgentStartLabel == toolCall.Function.Name
}

// buildEventType 事件类型构造
func buildEventType(subEvent *SubEventData) AgentEventType {
	if subEvent == nil {
		return MainAgentEventType
	}
	return SubAgentEventType
}

func filterMultiMessage(chatMessage *schema.Message) bool {
	if !toolParamsEnd(chatMessage) && !toolEnd(chatMessage) && len(chatMessage.Content) == 0 && len(chatMessage.ToolCalls) > 0 {
		tool := chatMessage.ToolCalls[0]
		if len(tool.Function.Name) == 0 && len(tool.Function.Arguments) == 0 {
			return true
		}
	}
	return false
}

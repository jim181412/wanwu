package response

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/UnicomAI/wanwu/internal/agent-service/model/request"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/cloudwego/eino/schema"
)

const (
	toolStartTitle        = `<tool>`
	toolStartTitleFormat  = `工具名：%s`
	toolParamsStartFormat = "\n\n```工具参数：\n"
	toolParamsEndFormat   = "\n```\n\n"
	toolEndFormat         = "\n\n```工具%s调用结果：\n %s \n```\n\n"
	toolEndTitle          = `</tool>`
	endLine               = "\n\n"
	agentSuccessCode      = 0
	agentFailCode         = 1
	finish                = 1
	notFinish             = 0

	ToolNameStep         ToolStep = 0 //输出工具名阶段
	ToolParamStartStep   ToolStep = 1 //输出工具参数 开始阶段
	ToolParamStep        ToolStep = 2 //输出工具参数阶段
	ToolParamFinishStep  ToolStep = 3 //输出工具参数完成阶段
	ToolResultFinishStep ToolStep = 4 //输出工具结果完成阶段
)

type ToolStep int

type AgentTool struct {
	ToolId   string
	ToolStep ToolStep //工具阶段
	Order    int      //工具顺序
}

type AgentChatRespContext struct {
	MainAgentName      string //主智能体名称
	MultiAgent         bool   //多智能体
	AgentStart         bool   //智能体开始
	AgentStartTime     int64
	AgentTempMessage   strings.Builder
	CurrentAgent       string //当前智能体
	CurrentAgentId     string //当前智能体Id
	CurrentAgentAvatar string //当前智能体图片
	ExitTool           bool   //退出工具开始
	//上面为多智能体相关参数
	HasTool            bool // 是否包含工具
	ToolStart          bool // 是否工具已开始
	ToolEnd            bool // 是否工具已结束
	ToolIndex          int  // 工具索引
	ToolCountMap       map[string]int
	CurrentToolId      string //当前toolId
	ToolMap            map[string]*AgentTool
	ReplaceContent     strings.Builder // 替换内容，如果出现相同内则则进行替换
	ReplaceContentStr  string          // 替换内容，如果出现相同内则则进行替换
	ReplaceContentDone bool            //替换内容准备完成

	ToolParamsStartCount int //工具参数开始的数量
	ToolParamsEndCount   int //工具参数结束的数量
}

func NewAgentChatRespContext(multiAgent bool, mainAgentName string) *AgentChatRespContext {
	return &AgentChatRespContext{
		MainAgentName: mainAgentName,
		ToolCountMap:  make(map[string]int),
		ToolMap:       make(map[string]*AgentTool),
		ToolIndex:     -1,
		MultiAgent:    multiAgent,
	}
}

type AgentChatResp struct {
	Code           int             `json:"code"`
	Message        string          `json:"message"`
	Response       string          `json:"response"`
	EventType      AgentEventType  `json:"eventType"`
	EventData      *SubEventData   `json:"eventData"`
	GenFileUrlList []interface{}   `json:"gen_file_url_list"`
	History        []interface{}   `json:"history"`
	Finish         int             `json:"finish"`
	Usage          *AgentChatUsage `json:"usage"`
	SearchList     []interface{}   `json:"search_list"`
	QaType         int             `json:"qa_type"`
}

type AgentChatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func NewAgentChatRespWithTool(chatMessage *schema.Message, respContext *AgentChatRespContext, req *request.AgentChatContext) ([]string, error) {
	//glm 模型在输出finish_reason = stop 还会token 统计消息导致finish= 1后又输出finish=0
	//所以此处做过滤
	if filterMessage(chatMessage) {
		return make([]string, 0), nil
	}
	if respContext.MultiAgent { //多智能体单独处理
		return MultiNewAgentChatRespWithTool(chatMessage, respContext, req)
	}
	contentList := buildNewContentWithTool(chatMessage, respContext)
	var outputList = make([]string, 0)
	for _, content := range contentList {
		var agentChatResp = &AgentChatResp{
			Code:           agentSuccessCode,
			Message:        "success",
			Response:       content,
			GenFileUrlList: []interface{}{},
			History:        []interface{}{},
			QaType:         buildQaType(req),
			SearchList:     buildSearchList(req),
			Finish:         buildFinish(chatMessage, false),
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

func AgentChatFailResp() string {
	var agentChatResp = &AgentChatResp{
		Code:     agentFailCode,
		Message:  "智能体处理异常，请稍后重试",
		Response: "智能体处理异常，请稍后重试",
		Finish:   finish,
	}
	respString, err := buildRespString(agentChatResp)
	if err != nil {
		log.Errorf("buildRespString error: %v", err)
		return ""
	}
	return respString
}

func filterMessage(chatMessage *schema.Message) bool {
	if string(chatMessage.Role) == "" && chatMessage.Content == "" {
		return true
	}
	return false
}

func buildRespString(agentChatResp *AgentChatResp) (string, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false) // 关键：禁用 HTML 转义

	if err := encoder.Encode(agentChatResp); err != nil {
		return "", err
	}
	return "data:" + buf.String(), nil
}

// buildNewContentWithTool 构造输出内容
// 构造tool的当前步骤；如果没有步骤就直接输出内容信息
func buildNewContentWithTool(chatMessage *schema.Message, respContext *AgentChatRespContext) []string {
	stepsMap, toolIdList := buildToolStep(chatMessage, respContext)
	if len(stepsMap) == 0 { //没有工具处理
		return buildNoToolContent(chatMessage, respContext)
	} else {
		return buildToolContent(chatMessage, respContext, stepsMap, toolIdList)
	}
}

/*
*
目前工具调用有几种情况做处理
1.正常流式：先输出方法名，在流式分别输出方法对应的参数，再输出调用结果
2.并发流式：如果需要调用同一方法两次，先输出方法名，方法参数，再输出方法名方法参数，再输出结果1，再输出结果2
3.同步请求：请求一个事件，返回一个事件，没有流式
4.同步请求和返回：请求和返回都在同一个事件，没有流式
*/
func buildToolStep(chatMessage *schema.Message, respContext *AgentChatRespContext) (map[string][]ToolStep, []string) {
	var toolStepMap = make(map[string][]ToolStep)
	//构造toolId
	var toolId = buildToolId(chatMessage, respContext)

	var toolIdList []string
	if toolStart(chatMessage) {
		for _, tool := range chatMessage.ToolCalls {
			newTool := isNewTool(tool, respContext)
			if newTool { //新工具开始
				toolId = tool.ID
			}
			steps := toolStepMap[toolId]
			if len(tool.Function.Name) > 0 {
				steps = append(steps, ToolNameStep)
				if newTool {
					steps = append(steps, ToolParamStartStep)
				}
			}

			if len(tool.Function.Arguments) > 0 {
				steps = append(steps, ToolParamStep)
			}
			if toolParamsEnd(chatMessage) {
				steps = append(steps, ToolParamFinishStep)
			}
			toolStepMap[toolId] = steps
			toolIdList = append(toolIdList, toolId)
		}
	} else if toolParamsEnd(chatMessage) {
		steps := toolStepMap[toolId]
		steps = append(steps, ToolParamFinishStep)
		toolStepMap[toolId] = steps
		toolIdList = append(toolIdList, toolId)
	} else if toolEnd(chatMessage) {
		steps := toolStepMap[toolId]
		steps = append(steps, ToolResultFinishStep)
		toolStepMap[toolId] = steps
		toolIdList = append(toolIdList, toolId)
	}
	return toolStepMap, toolIdList
}

// 构造toolId
// case1:工具同步调用结果，或者模型处理较好会直接返回模型id
// case2:触发了工具的并发调用即，先输出了两此工具参数，此时输出工具调用结果，如果没有toolId就默认按顺序填充结果
// case3:参数输出过程中，或者工具同步调用结果 没有toolId 标识，则返回当前toolId（上次参数输出的toolId）
func buildToolId(chatMessage *schema.Message, respContext *AgentChatRespContext) string {
	if len(chatMessage.ToolCallID) > 0 {
		return chatMessage.ToolCallID
	}
	toolIdList := filerToolByStep(respContext, ToolResultFinishStep, false)
	if len(toolIdList) > 1 { //此处表示有多个工具并发调用了
		var agentToolList []*AgentTool
		for _, toolId := range toolIdList {
			agentToolList = append(agentToolList, respContext.ToolMap[toolId])
		}
		sort.Slice(agentToolList, func(i, j int) bool {
			return agentToolList[i].Order < agentToolList[j].Order
		})
		return agentToolList[0].ToolId
	}
	return respContext.CurrentToolId
}

// filerToolByStep,equalCondition为true 则过滤等于此类型的tool，为false 则过滤不等于此类型的tool
func filerToolByStep(respContext *AgentChatRespContext, step ToolStep, equalCondition bool) []string {
	if len(respContext.ToolMap) > 0 {
		var toolIdList []string
		for toolId, tool := range respContext.ToolMap {
			if filterToolByCondition(tool, step, equalCondition) {
				toolIdList = append(toolIdList, toolId)
			}
		}
		return toolIdList
	}
	return nil
}

func filterToolByCondition(tool *AgentTool, step ToolStep, equalCondition bool) bool {
	if equalCondition {
		return tool.ToolStep == step
	} else {
		return tool.ToolStep != step
	}
}

// buildToolContent 构造有工具的内容输出
// 需要额外判断，如果此次输出的步骤不包含当前任务的步骤，同时之前工具有参数未完成的，则补充个参数结束的内容（处理并发调用工具的情况）
func buildToolContent(chatMessage *schema.Message, respContext *AgentChatRespContext, stepsMap map[string][]ToolStep, toolIdList []string) []string {
	steps := stepsMap[respContext.CurrentToolId]
	paramsNotFinishList := filerToolByStep(respContext, ToolParamStep, true)
	var contentList []string
	if len(steps) == 0 && len(paramsNotFinishList) > 0 { //是新工具且之前工具处于参数处理未完成状态
		//增加参数处理完成结果，并更改状态
		for _, toolId := range paramsNotFinishList {
			tool := respContext.ToolMap[toolId]
			if tool == nil {
				continue
			}
			//更改状态
			tool.ToolStep = ToolParamFinishStep
			//输出结果，增加结束
			contentList = append(contentList, toolParamsEndFormat)
		}
	}
	//根据step循环构造输出的内容
	for _, toolId := range toolIdList {
		toolSteps := stepsMap[toolId]
		agentTool := respContext.ToolMap[toolId]
		if agentTool == nil {
			agentTool = &AgentTool{ToolId: toolId, Order: len(respContext.ToolMap)}
			respContext.ToolMap[toolId] = agentTool
		}
		for _, step := range toolSteps {
			agentTool.ToolStep = step
			toolContentList := buildContentByStep(chatMessage, step, toolId)
			if len(toolContentList) == 0 {
				continue
			}
			contentList = append(contentList, toolContentList...)
		}
		respContext.CurrentToolId = toolId
	}
	return contentList
}

// buildContentByStep 根据当前步骤构造需要输出的内容,构造<tool></tool>数据以及markdown格式
func buildContentByStep(chatMessage *schema.Message, step ToolStep, toolId string) []string {
	var contentList []string
	switch step {
	case ToolNameStep:
		tool := buildMessageTool(chatMessage, toolId)
		if tool == nil {
			break
		}
		toolName := fmt.Sprintf(toolStartTitleFormat, tool.Function.Name)
		contentList = append(contentList, toolName)
	case ToolParamStartStep:
		contentList = append(contentList, toolStartTitle)
		contentList = append(contentList, toolParamsStartFormat)
	case ToolParamStep:
		tool := buildMessageTool(chatMessage, toolId)
		if tool == nil {
			break
		}
		contentList = append(contentList, tool.Function.Arguments)
	case ToolParamFinishStep:
		contentList = append(contentList, toolParamsEndFormat)
	case ToolResultFinishStep:
		toolResult := fmt.Sprintf(toolEndFormat, chatMessage.ToolName, chatMessage.Content)
		contentList = append(contentList, toolResult)
		contentList = append(contentList, toolEndTitle)
	}
	return contentList
}

// buildMessageTool 构造消息工具内容数据
func buildMessageTool(chatMessage *schema.Message, toolId string) *schema.ToolCall {
	var length = len(chatMessage.ToolCalls)
	if length == 0 {
		return nil
	} else if length == 1 {
		return &chatMessage.ToolCalls[0]
	}

	for _, call := range chatMessage.ToolCalls {
		if call.ID == toolId {
			return &call
		}
	}
	return nil
}

// buildNoToolContent 构造没有工具的内容
// case1：tool 有数据同时content内容；如果此时在工具的输出中还没有输出完，则不输出content的相关内容
// case2：在tool输出前会输出规划内容，但是会重复输出相同的规划内容，所以当内容数字大于10时，同时出现重复数据，则不输出
// case3：正式输出
func buildNoToolContent(chatMessage *schema.Message, respContext *AgentChatRespContext) []string {
	notFinishList := filerToolByStep(respContext, ToolResultFinishStep, false)
	if len(notFinishList) > 0 { //在工具期间，不输出任何content内容
		return []string{}
	}
	//替换内容准备(工具未开始，但是输出了内容, 有的模型会重复输出一样的话)
	if len(respContext.ToolMap) == 0 {
		if utf8.RuneCountInString(chatMessage.Content) > 10 {
			var replaceContent = respContext.ReplaceContentStr
			if len(replaceContent) == 0 {
				replaceContent = respContext.ReplaceContent.String()
			}
			if replaceContent == chatMessage.Content {
				respContext.ReplaceContentDone = true
				respContext.ReplaceContentStr = replaceContent
				return []string{}
			}
		}
		if !respContext.ReplaceContentDone {
			respContext.ReplaceContent.WriteString(chatMessage.Content)
		}
	}
	return []string{chatMessage.Content}
}

//func buildContentWithTool(chatMessage *schema.Message, respContext *AgentChatRespContext) []string {
//	if toolStart(chatMessage) {
//		respContext.ToolStart = true
//		respContext.HasTool = true
//		var retList []string
//
//		for _, tool := range chatMessage.ToolCalls {
//			newTool := isNewTool(tool, respContext)
//			if !newTool && len(tool.Function.Arguments) > 0 { //只有此次不是新工具才add 参数，处理先输出方法名，后流式输出参数的情况
//				retList = append(retList, tool.Function.Arguments)
//				continue
//			}
//			if tool.Type == "function" {
//				if respContext.ToolIndex == -1 {
//					respContext.ToolIndex = *tool.Index
//				} else if *tool.Index != respContext.ToolIndex { //模型触发并发请求工具的bad case
//					respContext.ToolIndex = *tool.Index
//					if respContext.ToolParamsStartCount > respContext.ToolParamsEndCount {
//						respContext.ToolParamsEndCount = respContext.ToolParamsEndCount + 1
//						retList = append(retList, toolParamsEndFormat)
//					}
//				}
//				if len(tool.Function.Name) > 0 {
//					toolName := fmt.Sprintf(toolStartTitleFormat, tool.Function.Name)
//					retList = append(retList, toolName)
//				}
//
//				if newTool {
//					respContext.ToolParamsStartCount = respContext.ToolParamsStartCount + 1
//					retList = append(retList, toolStartTitle)
//					retList = append(retList, toolParamsStartFormat)
//					respContext.ToolCountMap[tool.ID] = 1
//					if len(tool.Function.Arguments) > 0 { //处理一次性同时返回 方法名和参数的情况
//						retList = append(retList, tool.Function.Arguments)
//					}
//					if toolParamsEnd(chatMessage) { //处理一次性同时返回 方法名和参数的情况 ,同时返回结束了
//						retList = append(retList, toolParamsEndFormat)
//					}
//				}
//			}
//		}
//
//		return retList
//	} else if toolParamsEnd(chatMessage) {
//		respContext.ToolParamsEndCount = respContext.ToolParamsEndCount + 1
//		return []string{toolParamsEndFormat}
//	} else if toolEnd(chatMessage) {
//		respContext.ToolEnd = true
//		toolResult := fmt.Sprintf(toolEndFormat, chatMessage.ToolName, chatMessage.Content)
//		respContext.ToolCountMap[chatMessage.ToolCallID] = 0
//		return []string{toolResult, toolEndTitle}
//	} else {
//		//在工具期间，不输出任何content内容
//		if respContext.ToolStart && !respContext.ToolEnd {
//			return []string{}
//		}
//		//替换内容准备(工具未开始，但是输出了内容)
//		if !respContext.ToolStart {
//			if utf8.RuneCountInString(chatMessage.Content) > 10 {
//				var replaceContent = respContext.ReplaceContentStr
//				if len(replaceContent) == 0 {
//					replaceContent = respContext.ReplaceContent.String()
//				}
//				if replaceContent == chatMessage.Content {
//					respContext.ReplaceContentDone = true
//					respContext.ReplaceContentStr = replaceContent
//					return []string{}
//				}
//			}
//			if !respContext.ReplaceContentDone {
//				respContext.ReplaceContent.WriteString(chatMessage.Content)
//			}
//		}
//		return []string{chatMessage.Content}
//	}
//}

func isNewTool(tool schema.ToolCall, respContext *AgentChatRespContext) bool {
	return len(tool.ID) > 0 && respContext.ToolMap[tool.ID] == nil
}

func toolStart(chatMessage *schema.Message) bool {
	return len(chatMessage.ToolCalls) > 0
}

func toolParamsEnd(chatMessage *schema.Message) bool {
	responseMeta := chatMessage.ResponseMeta
	if responseMeta == nil {
		return false
	}
	return responseMeta.FinishReason == "tool_calls"
}

func toolEnd(chatMessage *schema.Message) bool {
	return chatMessage.Role == schema.Tool
}

func buildFinish(chatMessage *schema.Message, notStop bool) int {
	if notStop {
		return notFinish
	}
	if chatMessage.ResponseMeta != nil && chatMessage.ResponseMeta.FinishReason == "stop" {
		return finish
	}
	if chatMessage.Role == schema.Tool && chatMessage.ToolName == "exit" {
		return finish
	}
	return notFinish
}

func buildUsage(chatMessage *schema.Message) *AgentChatUsage {
	if chatMessage.ResponseMeta != nil && chatMessage.ResponseMeta.Usage != nil {
		usage := chatMessage.ResponseMeta.Usage
		return &AgentChatUsage{
			PromptTokens:     usage.PromptTokens,
			CompletionTokens: usage.CompletionTokens,
			TotalTokens:      usage.TotalTokens,
		}
	}
	return &AgentChatUsage{}
}

func buildSubAgentSearchList(subAgentEventData *SubEventData, req *request.AgentChatContext) []interface{} {
	if subAgentEventData == nil || req == nil || len(req.SubAgentMap) == 0 {
		return nil
	}
	config := req.SubAgentMap[subAgentEventData.Name]
	if config == nil || config.AgentChatContext == nil {
		return nil
	}

	return buildSearchList(config.AgentChatContext)
}

func buildSearchList(req *request.AgentChatContext) []interface{} {
	if req.KnowledgeHitData == nil {
		return []interface{}{}
	}
	list := req.KnowledgeHitData.SearchList
	var retList = make([]interface{}, 0)
	if len(list) > 0 {
		for _, item := range list {
			retList = append(retList, item)
		}
	}
	return retList
}

func buildQaType(req *request.AgentChatContext) int {
	if req.KnowledgeHitData == nil {
		return 0
	}
	return 1
}

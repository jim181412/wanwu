package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	"github.com/gin-gonic/gin"
)

func ModelExperienceLLM(ctx *gin.Context, userId, orgId string, req *request.ModelExperienceLlmRequest) {
	// model info
	modelInfo, err := model.GetModel(ctx.Request.Context(), &model_service.GetModelReq{ModelId: req.ModelId})
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	if !modelInfo.IsActive {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFModelStatus, modelInfo.ModelId))
		return
	}

	// dialog records
	recordsResp, err := model.GetModelExperienceDialogRecords(ctx, &model_service.GetModelExperienceDialogRecordsReq{
		UserId: userId,
		OrgId:  orgId,
		// 常规模型对话记录（非模型对比时），modelExperienceId与sessionId非空
		// 模型对比时临时存储对话记录，modelExperienceId前端传空，sessionId非空
		ModelExperienceId: req.ModelExperienceId,
		SessionId:         req.SessionId,
	})
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	var messages []mp_common.OpenAIReqMsg
	for _, record := range recordsResp.Records {
		content := record.HandledContent
		if content == "" {
			content = record.OriginalContent
		}
		messages = append(messages, mp_common.OpenAIReqMsg{
			Role:    mp_common.MsgRole(record.Role),
			Content: content,
		})
	}
	// 添加当前用户消息
	messages = append(messages, mp_common.OpenAIReqMsg{
		Role:    mp_common.MsgRoleUser,
		Content: req.Content,
	})

	// 构造LLM请求
	stream := true
	llmReq := &mp_common.LLMReq{
		Model:    modelInfo.Model,
		Messages: messages,
		Stream:   &stream,
	}
	if req.TemperatureEnable {
		temp := float64(req.Temperature)
		llmReq.Temperature = &temp
	}
	if req.TopPEnable {
		topP := float64(req.TopP)
		llmReq.TopP = &topP
	}
	if req.PresencePenaltyEnable {
		presencePenalty := float64(req.PresencePenalty)
		llmReq.PresencePenalty = &presencePenalty
	}
	if req.FrequencyPenaltyEnable {
		frequencyPenalty := float64(req.FrequencyPenalty)
		llmReq.FrequencyPenalty = &frequencyPenalty
	}
	if req.MaxTokensEnable {
		maxTokens := int(req.MaxTokens)
		llmReq.MaxTokens = &maxTokens
	}

	// llm config
	llm, err := mp.ToModelConfig(modelInfo.Provider, modelInfo.ModelType, modelInfo.ProviderConfig)
	if err != nil {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, err.Error()))
		return
	}
	iLLM, ok := llm.(mp.ILLM)
	if !ok {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, err.Error()))
		return
	}

	// chat completions
	iLLMReq, err := iLLM.NewReq(llmReq)
	if err != nil {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, err.Error()))
		return
	}
	_, sseCh, err := iLLM.ChatCompletions(ctx.Request.Context(), iLLMReq)
	if err != nil {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, err.Error()))
		return
	}

	// save query
	if _, err := model.SaveModelExperienceDialogRecord(ctx.Request.Context(), &model_service.SaveModelExperienceDialogRecordReq{
		UserId:            userId,
		OrgId:             orgId,
		ModelExperienceId: req.ModelExperienceId,
		ModelId:           req.ModelId,
		SessionId:         req.SessionId,
		OriginalContent:   req.Content,
		Role:              string(mp_common.MsgRoleUser),
	}); err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}

	// stream
	var answer string
	var reasonContent string
	var (
		firstFlag = false // 思维链起始标识符，默认思维链未开始
		endFlag   = false // 思维链结束标识符，默认思维链未结束
	)
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Content-Type", "text/event-stream; charset=utf-8")
	var data *mp_common.LLMResp

	for sseResp := range sseCh {
		data, ok = sseResp.ConvertResp()
		dataStr := ""
		if ok && data != nil && len(data.Choices) > 0 {
			delta := data.Choices[0].Delta
			if delta != nil {
				answer = answer + delta.Content

				if delta.ReasoningContent != nil {
					reasonContent = reasonContent + *delta.ReasoningContent
				}
			}
			if firstFlag && !endFlag && delta.ReasoningContent != nil {
				delta.Content = delta.Content + *delta.ReasoningContent
			}
			if !endFlag && delta.Content != "" && ((delta.ReasoningContent != nil &&
				*delta.ReasoningContent == "") || delta.ReasoningContent == nil) && firstFlag {
				delta.Content = "\n</think>\n" + delta.Content
				endFlag = true
			}
			if !firstFlag && delta.ReasoningContent != nil && *delta.ReasoningContent != "" && delta.Content == "" {
				delta.Content = "<think>\n" +
					delta.Content + *delta.ReasoningContent
				firstFlag = true
			}
			dataByte, _ := json.Marshal(data)
			dataStr = fmt.Sprintf("data: %v\n", string(dataByte))
		} else {
			dataStr = fmt.Sprintf("%v\n", sseResp.String())
		}
		if _, err = ctx.Writer.Write([]byte(dataStr)); err != nil {
			log.Errorf("mdoel experience write sse err: %v", err)
		}
		ctx.Writer.Flush()
	}

	// save answer
	if _, err := model.SaveModelExperienceDialogRecord(ctx.Request.Context(), &model_service.SaveModelExperienceDialogRecordReq{
		UserId:            userId,
		OrgId:             orgId,
		ModelExperienceId: req.ModelExperienceId,
		ModelId:           req.ModelId,
		SessionId:         req.SessionId,
		OriginalContent:   answer,
		ReasoningContent:  reasonContent,
		Role:              string(mp_common.MsgRoleAssistant),
	}); err != nil {
		log.Errorf("mdoel experience save record err: %v", err)
		return
	}

	ctx.Set(gin_util.STATUS, http.StatusOK)
	ctx.Set(gin_util.RESULT, answer)
}

func SaveModelExperienceDialog(ctx *gin.Context, userId, orgId string, req *request.ModelExperienceDialogRequest) (*response.ModelExperienceDialog, error) {
	// 将interface{}类型的ModelSetting转换为 json string
	var modelSettingStr string
	if req.ModelSetting != nil {
		jsonBytes, err := json.Marshal(req.ModelSetting)
		if err != nil {
			return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("Model settings serialization error: err: %v", err))
		}
		modelSettingStr = string(jsonBytes)
	}
	dialog, err := model.SaveModelExperienceDialog(ctx.Request.Context(), &model_service.SaveModelExperienceDialogReq{
		UserId:       userId,
		OrgId:        orgId,
		ModelId:      req.ModelId,
		SessionId:    req.SessionId,
		ModelSetting: modelSettingStr,
		Title:        req.Title,
	})
	if err != nil {
		return nil, err
	}
	return toModelExperienceDialog(dialog), nil
}

func ListModelExperienceDialogs(ctx *gin.Context, userId, orgId string) (*response.ListResult, error) {
	resp, err := model.GetModelExperienceDialogs(ctx.Request.Context(), &model_service.ListModelExperienceDialogReq{
		UserId: userId,
		OrgId:  orgId,
	})
	if err != nil {
		return nil, err
	}

	// 收集所有唯一的模型ID
	modelIdMap := make(map[string]bool)
	for _, dialog := range resp.Dialogs {
		modelIdMap[dialog.ModelId] = true
	}

	// 提取唯一模型ID列表
	var uniqueModelIds []string
	for modelId := range modelIdMap {
		uniqueModelIds = append(uniqueModelIds, modelId)
	}

	// 批量检查模型权限
	authorizedModelIds, _ := CheckModelUserPermission(ctx, userId, orgId, uniqueModelIds)

	// 创建授权模型ID的映射，用于快速查找
	authorizedModelMap := make(map[string]bool)
	for _, modelId := range authorizedModelIds {
		authorizedModelMap[modelId] = true
	}

	// 过滤出用户有权限的对话
	var dialogs []*response.ModelExperienceDialog
	for _, dialog := range resp.Dialogs {
		if authorizedModelMap[dialog.ModelId] {
			dialogs = append(dialogs, toModelExperienceDialog(dialog))
		}
	}

	return &response.ListResult{
		List:  dialogs,
		Total: int64(len(resp.Dialogs)),
	}, nil
}

func DeleteModelExperienceDialog(ctx *gin.Context, userId, orgId, modelExperienceId string) error {
	_, err := model.DeleteModelExperienceDialog(ctx, &model_service.ModelExperienceDialogReq{
		ModelExperienceId: modelExperienceId,
		UserId:            userId,
		OrgId:             orgId,
	})
	return err
}

func ListModelExperienceDialogRecords(ctx *gin.Context, userId, orgId string, req *request.ModelExperienceDialogRecordRequest) (*response.ListResult, error) {
	recordsResp, err := model.GetModelExperienceDialogRecords(ctx, &model_service.GetModelExperienceDialogRecordsReq{
		UserId: userId,
		OrgId:  orgId,
		// 常规模型对话记录（非模型对比时），modelExperienceId非空，sessionId前端没传
		ModelExperienceId: req.ModelExperienceId,
		SessionId:         "",
	})
	if err != nil {
		return nil, err
	}
	var records []*response.ModelExperienceDialogRecord
	for _, record := range recordsResp.Records {
		records = append(records, &response.ModelExperienceDialogRecord{
			ModelExperienceId: record.ModelExperienceId,
			ModelId:           record.ModelId,
			SessionId:         record.SessionId,
			OriginalContent:   record.OriginalContent,
			ReasoningContent:  record.ReasoningContent,
			Role:              record.Role,
		})
	}
	return &response.ListResult{
		List:  records,
		Total: int64(len(records)),
	}, nil
}
func toModelExperienceDialog(dialog *model_service.ModelExperienceDialog) *response.ModelExperienceDialog {
	return &response.ModelExperienceDialog{
		ID:           dialog.ModelExperienceId,
		ModelId:      dialog.ModelId,
		SessionId:    dialog.SessionId,
		Title:        dialog.Title,
		ModelSetting: dialog.ModelSetting,
		CreatedAt:    dialog.CreatedAt,
	}
}

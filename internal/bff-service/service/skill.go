package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/minio"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	sse_util "github.com/UnicomAI/wanwu/pkg/sse-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	wga_sandbox_option "github.com/UnicomAI/wanwu/pkg/wga-sandbox/wga-sandbox-option"
	"github.com/gin-gonic/gin"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
)

const (
	skillConversationESIndexName = "skill_creation_conversation_detail_*"
)

func GetAgentSkillList(ctx *gin.Context, name string) (*response.ListResult, error) {
	var skillsTemplateList []*response.SkillDetail
	for _, skillsCfg := range config.Cfg().AgentSkills {
		if name != "" && !strings.Contains(skillsCfg.Name, name) {
			continue
		}
		skillsTemplateList = append(skillsTemplateList, buildSkillTempDetail(*skillsCfg, false))
	}
	return &response.ListResult{
		List:  skillsTemplateList,
		Total: int64(len(skillsTemplateList)),
	}, nil
}

func GetAgentSkillDetail(ctx *gin.Context, skillId string) (*response.SkillDetail, error) {
	skillsCfg, exist := config.Cfg().AgentSkill(skillId)
	if !exist {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "bff_agent_skill_detail", "get skill detail empty")
	}
	return buildSkillTempDetail(skillsCfg, true), nil
}

func DownloadAgentSkill(ctx *gin.Context, skillId string) ([]byte, error) {
	// 需要把SkConfigDir+templateId路径下的所有文件在内存打成一个压缩包
	sf, exist := config.Cfg().AgentSkill(skillId)
	if !exist {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "bff_agent_skill_download", "get skill detail empty")
	}
	return sf.AgentSkillZipToBytes(skillId)
}

// --- Skill Conversation ---

func CreateSkillConversation(ctx *gin.Context, userId, orgId string, req request.CreateSkillConversationReq) (*response.CreateSkillConversationResp, error) {
	rpcResp, err := assistant.CreateSkillConversation(ctx.Request.Context(), &assistant_service.CreateSkillConversationReq{
		Title: req.Title,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}
	return &response.CreateSkillConversationResp{
		ConversationId: rpcResp.ConversationId,
	}, nil
}

func DeleteSkillConversation(ctx *gin.Context, userId, orgId, conversationId string) error {
	// 异步删除 ES 中的历史记录
	go func() {
		// 索引格式为 skill_creation_conversation_detail_*
		_, _ = assistant.DeleteFromES(ctx.Request.Context(), &assistant_service.DeleteFromESReq{
			IndexName: "skill_creation_conversation_detail_*",
			Conditions: map[string]string{
				"conversationId": conversationId,
			},
		})
	}()

	_, err := assistant.DeleteSkillConversation(ctx.Request.Context(), &assistant_service.DeleteSkillConversationReq{
		ConversationId: conversationId,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func GetSkillConversationList(ctx *gin.Context, userId, orgId string, req request.GetSkillConversationListReq) (*response.PageResult, error) {
	rpcResp, err := assistant.GetSkillConversationList(ctx.Request.Context(), &assistant_service.GetSkillConversationListReq{
		PageNo:   int32(req.PageNo),
		PageSize: int32(req.PageSize),
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}

	list := make([]response.SkillConversationItem, 0, len(rpcResp.List))
	for _, item := range rpcResp.List {
		list = append(list, response.SkillConversationItem{
			ConversationId: item.ConversationId,
			Title:          item.Title,
			CreatedAt:      util.Time2Str(item.CreatedAt),
		})
	}

	return &response.PageResult{
		List:     list,
		Total:    rpcResp.Total,
		PageNo:   req.PageNo,
		PageSize: req.PageSize,
	}, nil
}

func GetSkillConversationDetail(ctx *gin.Context, userId, orgId string, req request.GetSkillConversationDetailReq) (*response.ListResult, error) {
	if req.ConversationId == "" {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "bff_skill_conversation_detail", "conversationId is empty")
	}

	// 从 ES 中读取
	searchResp, err := assistant.SearchFromES(ctx.Request.Context(), &assistant_service.SearchFromESReq{
		IndexName: skillConversationESIndexName,
		Conditions: map[string]string{
			"conversationId": req.ConversationId,
		},
		PageNo:    1,
		PageSize:  1000,
		SortField: "createdAt",
		SortOrder: "desc",
	})
	if err != nil {
		return nil, err
	}

	// 提取 SaveId
	saveIds := make([]string, 0, len(searchResp.DocJsonList))

	// 会话记录
	respList := make([]*response.SkillConversationDetailInfo, 0, len(searchResp.DocJsonList))
	for _, docJSON := range searchResp.DocJsonList {
		var item response.SkillConversationDetailInfo
		if err := json.Unmarshal([]byte(docJSON), &item); err != nil {
			continue
		}
		respList = append(respList, &item)
		// 提取 SaveId
		for _, rf := range item.ResponseFiles {
			if skillSaveId, ok := rf.MetaData["skillSaveId"].(string); ok {
				saveIds = append(saveIds, skillSaveId)
			}
		}
	}

	// 是否已发送
	mcpResp, err := mcp.CustomSkillGetBySaveIds(ctx.Request.Context(), &mcp_service.CustomSkillGetBySaveIdsReq{
		SaveIds: saveIds,
	})
	if err != nil {
		return nil, err
	}

	// 有效的 saveIds 集合
	validSaveIds := make(map[string]bool, len(mcpResp.SaveIds))
	for _, sid := range mcpResp.SaveIds {
		validSaveIds[sid] = true
	}

	// 标记是否已发送
	for _, item := range respList {
		for i := range item.ResponseFiles {
			if skillSaveId, ok := item.ResponseFiles[i].MetaData["skillSaveId"].(string); ok {
				item.ResponseFiles[i].MetaData["inResource"] = validSaveIds[skillSaveId]
			}
		}
	}

	return &response.ListResult{
		List:  respList,
		Total: searchResp.Total,
	}, nil
}

func SkillConversationChat(ctx *gin.Context, userId, orgId string, req request.SkillConversationChatReq) error {

	// 查询模型信息
	modelInfo, err := model.GetModel(ctx.Request.Context(), &model_service.GetModelReq{
		ModelId: req.ModelConfig.ModelId,
		UserId:  userId,
		OrgId:   orgId,
	})
	if err != nil {
		return err
	}

	// 从 ES 中读取
	indexName := "skill_creation_conversation_detail_*"
	searchResp, err := assistant.SearchFromES(ctx.Request.Context(), &assistant_service.SearchFromESReq{
		IndexName: indexName,
		Conditions: map[string]string{
			"conversationId": req.ConversationId,
		},
		PageNo:    1,
		PageSize:  1000,
		SortField: "createdAt",
		SortOrder: "desc",
	})
	if err != nil {
		return err
	}

	// 会话记录
	messages := make([]wga_sandbox_option.Message, 0, len(searchResp.DocJsonList))
	for _, docJSON := range searchResp.DocJsonList {
		var item response.SkillConversationDetailInfo
		if err := json.Unmarshal([]byte(docJSON), &item); err != nil {
			continue
		}
		messages = append(messages, wga_sandbox_option.Message{
			Role:    "user",
			Content: item.Prompt,
		}, wga_sandbox_option.Message{
			Role:    "assistant",
			Content: item.Response,
		})

	}

	// 存储路径 /tmp/skills/<uuid>
	fileName := util.GenUUID()
	outputDir := fmt.Sprintf("/tmp/skills/%v", fileName)

	// 模型
	modelConfig := wga_sandbox_option.ModelConfig{
		Provider:     modelInfo.Provider,
		ProviderName: modelInfo.Provider,
		Model:        modelInfo.Model,
		ModelName:    modelInfo.DisplayName,
	}
	endpoint := mp.ToModelEndpoint(modelInfo.ModelId, modelInfo.Model)
	for k, v := range endpoint {
		if k == "model_url" {
			modelConfig.BaseURL = v.(string)
			break
		}
	}

	// 流式问答
	streamCh, err := RunSkillCreator(ctx, modelConfig, "", outputDir, messages)
	if err != nil {
		return err
	}

	// 处理流式问答
	var responseStr string
	_ = sse_util.NewSSEWriter(ctx, fmt.Sprintf("[Skill] conversation %v user %v org %v recv", req.ConversationId, userId, orgId), sse_util.DONE_EMPTY).
		WriteStream(streamCh, nil, buildSkillChatRespLineProcessor(&responseStr), buildSkillChatDoneProcessor(ctx, userId, orgId, req, outputDir, &responseStr))
	return nil
}

func buildSkillChatDoneProcessor(ctx *gin.Context, userId, orgId string, req request.SkillConversationChatReq, outputDir string, responseStr *string) func(sse_util.SSEWriterClient[string], interface{}) error {
	return func(c sse_util.SSEWriterClient[string], params interface{}) error {

		mesageId := util.GenUUID()
		lastSSE := response.SkillConversationChatResp{
			Message: "success",
			Finish:  1,
		}

		defer func() {
			// save to es
			b, _ := json.Marshal(&response.SkillConversationDetailInfo{
				ConversationDetailInfo: response.ConversationDetailInfo{
					Id:             mesageId,
					ConversationId: req.ConversationId,
					Prompt:         req.Query,
					Response:       *responseStr,
				},
				ResponseFiles: lastSSE.ResponseFiles,
			})
			if _, err := assistant.SaveToES(ctx.Request.Context(), &assistant_service.SaveToESReq{
				IndexName: skillConversationESIndexName,
				DocJson:   string(b),
			}); err != nil {
				log.Errorf("[Skill] conversation %v user %v org %v save to es err: %v", req.ConversationId, userId, orgId, err)
			}

			// done sse
			marshal, _ := json.Marshal(lastSSE)
			data := "data: " + string(marshal) + "\n\n"
			_ = c.Write(data)
			_ = c.Write(sse_util.DONE_MSG)
			c.Flush()
		}()

		// 压缩文件夹
		zipBytes, err := util.ZipDir(outputDir)
		if err != nil {
			return err
		}
		// skillName, skillDesc
		_, skillName, skillDesc, err := extractSkillMarkdown(zipBytes)
		if err != nil {
			return err
		}
		// 上传到 minio
		fileName, _, err := minio.UploadFileCommon(ctx.Request.Context(), bytes.NewReader(zipBytes), ".zip", 0, false)
		if err != nil {
			return err
		}
		// 构建文件访问路径并添加到响应中
		lastSSE.ResponseFiles = append(lastSSE.ResponseFiles, &response.AssistantResponseFile{
			FileName: fileName,
			FileSize: int64(len(zipBytes)),
			FileUrl:  buildAccessFilePath(filepath.Join(minio.BucketFileUpload, minio.DirFileExpire, fileName)),
			MIMEType: "application/zip",
			MetaData: map[string]interface{}{
				"name":        skillName,
				"desc":        skillDesc,
				"author":      userId,
				"avatar":      cacheSkillAvatar(ctx, ""),
				"inResource":  false,
				"expiredAt":   util.Time2Str(time.Now().AddDate(0, 0, 7).UnixMilli()), // 7天后过期
				"skillSaveId": mesageId,
			},
		})
		// 删除临时文件
		if err := util.DeleteDirFile(outputDir); err != nil {
			return err
		}
		return nil
	}
}

func buildSkillChatRespLineProcessor(responeStr *string) func(sse_util.SSEWriterClient[string], string, interface{}) (string, bool, error) {
	return func(c sse_util.SSEWriterClient[string], lineText string, params interface{}) (string, bool, error) {

		// 累计流式输出
		*responeStr += lineText

		if strings.HasPrefix(lineText, "error:") {
			errorText := fmt.Sprintf("data: {\"code\": -1, \"message\": \"%s\"}\n\n", strings.TrimPrefix(lineText, "error:"))
			return errorText, false, nil
		}
		if strings.HasPrefix(lineText, "data:") {
			return lineText + "\n\n", false, nil
		}
		resp := response.SkillConversationChatResp{
			Response: lineText,
		}
		marshal, _ := json.Marshal(resp)
		return "data: " + string(marshal) + "\n\n", false, nil
	}
}

func SkillConversationSave(ctx *gin.Context, userId, orgId string, req request.SkillConversationSaveReq) (*response.CustomSkillIDResp, error) {

	// 从 ES 中读取
	searchResp, err := assistant.SearchFromES(ctx.Request.Context(), &assistant_service.SearchFromESReq{
		IndexName: skillConversationESIndexName,
		Conditions: map[string]string{
			"conversationId": req.ConversationId,
		},
		PageNo:    1,
		PageSize:  1000,
		SortField: "createdAt",
		SortOrder: "desc",
	})
	if err != nil {
		return nil, err
	}

	var zipUrl string
	for _, docJSON := range searchResp.DocJsonList {
		var item response.SkillConversationDetailInfo
		if err := json.Unmarshal([]byte(docJSON), &item); err != nil {
			continue
		}
		// 提取 SaveId
		for _, rf := range item.ResponseFiles {
			if skillSaveId, ok := rf.MetaData["skillSaveId"].(string); ok && skillSaveId == req.SkillSaveId {
				zipUrl = rf.FileUrl
			}
		}
	}

	skillId, err := CreateCustomSkill(ctx, userId, orgId, request.CreateCustomSkillReq{
		Author:     "wanwu",
		ZipUrl:     zipUrl,
		SaveId:     req.SkillSaveId,
		SourceType: "skill_conversation",
	})
	if err != nil {
		return nil, err
	}

	return skillId, nil
}

// --- internal ---
func buildSkillTempDetail(skillsCfg config.SkillsConfig, needMd bool) *response.SkillDetail {
	iconUrl := config.Cfg().DefaultIcon.SkillIcon
	if skillsCfg.Avatar != "" {
		iconUrl = skillsCfg.Avatar
	}
	ret := &response.SkillDetail{
		SkillId: skillsCfg.SkillId,
		Author:  skillsCfg.Author,
		Avatar:  request.Avatar{Path: iconUrl},
		Name:    skillsCfg.Name,
		Desc:    skillsCfg.Desc,
	}
	if needMd {
		ret.SkillMarkdown = string(skillsCfg.SkillMarkdown)
	}
	return ret
}

package params_process

import (
	"context"
	"encoding/json"

	"github.com/UnicomAI/wanwu/pkg/constant"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/log"
)

type SkillProcess struct {
}

func init() {
	AddServiceContainer(&SkillProcess{})
}

func (k *SkillProcess) ServiceType() ServiceType {
	return SkillType
}

func (k *SkillProcess) Prepare(agent *AgentInfo, prepareParams *AgentPrepareParams, clientInfo *ClientInfo, userQueryParams *UserQueryParams) error {
	skills := buildAssistantSkills(agent, clientInfo)
	if len(skills) == 0 {
		return nil
	}

	var builtinSkillIds []string
	var customSkillIds []string
	for _, skill := range skills {
		if !skill.Enable {
			continue
		}
		switch skill.SkillType {
		case constant.SkillTypeBuiltIn:
			builtinSkillIds = append(builtinSkillIds, skill.SkillId)
		case constant.SkillTypeCustom:
			customSkillIds = append(customSkillIds, skill.SkillId)
		}
	}

	//获取custom skill详情
	var skillInfos []*assistant_service.SkillInfo
	if len(customSkillIds) > 0 {
		customSkillResp, err := clientInfo.MCP.GetCustomSkillDetailByIdList(context.Background(), &mcp_service.CustomSkillDetailByIdListReq{
			SkillIds: customSkillIds,
		})
		if err != nil {
			log.Errorf("Assistant服务获取Custom Skill详情失败，assistantId: %d, error: %v", agent.Assistant.ID, err)
		} else if len(customSkillResp.SkillDetails) > 0 {
			for _, detail := range customSkillResp.SkillDetails {
				skillInfos = append(skillInfos, &assistant_service.SkillInfo{
					SkillId:     detail.SkillId,
					SkillType:   constant.SkillTypeCustom,
					SkillDetail: detail.Markdown,
					//minioPath:   buildAccessFilePath(detail.ObjectPath),
				})
			}
		}
	}

	// 获取builtin skill详情

	if len(builtinSkillIds) > 0 {
		for _, skillId := range builtinSkillIds {
			skillInfos = append(skillInfos, &assistant_service.SkillInfo{
				SkillId:     skillId,
				SkillType:   constant.SkillTypeBuiltIn,
				SkillDetail: "",
			})
		}
	}

	prepareParams.SkillList = skillInfos
	return nil
}

func buildAssistantSkills(agent *AgentInfo, clientInfo *ClientInfo) []*model.AssistantSkill {
	if agent.Draft {
		list, status := clientInfo.Cli.GetAssistantSkillList(context.Background(), agent.Assistant.ID)
		if status != nil {
			log.Errorf("GetAssistantSkillList error: %v", status)
			return nil
		}
		return list
	}
	var skillList []*model.AssistantSkill
	if agent.AssistantSnapshot.AssistantSkillConfig != "" {
		if err := json.Unmarshal([]byte(agent.AssistantSnapshot.AssistantSkillConfig), &skillList); err != nil {
			log.Errorf("GetAssistantSnapshotSkillList error: %v", err)
			return nil
		}
	}
	return skillList
}

func (k *SkillProcess) Build(assistant *AgentInfo, prepareParams *AgentPrepareParams, agentChatParams *assistant_service.AgentDetail) error {
	if len(prepareParams.SkillList) == 0 {
		return nil
	}
	if agentChatParams.SkillParams == nil {
		agentChatParams.SkillParams = &assistant_service.SkillParams{}
	}
	agentChatParams.SkillParams.SkillList = prepareParams.SkillList
	return nil
}

//func buildAccessFilePath(filePath string) string {
//	path := config.Cfg().Server.WebBaseUrl + "/minio/download/api/" + filePath
//	return path
//}

package service

import (
	"strings"

	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/gin-gonic/gin"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
)

func GetAgentSkillList(ctx *gin.Context, name string) (*response.ListResult, error) {
	var skillsTemplateList []*response.AgentSkillDetail
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

func GetAgentSkillDetail(ctx *gin.Context, skillId string) (*response.AgentSkillDetail, error) {
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

// --- internal ---
func buildSkillTempDetail(skillsCfg config.SkillsConfig, needMd bool) *response.AgentSkillDetail {
	iconUrl := config.Cfg().DefaultIcon.SkillIcon
	if skillsCfg.Avatar != "" {
		iconUrl = skillsCfg.Avatar
	}
	ret := &response.AgentSkillDetail{
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

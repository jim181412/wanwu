package response

import "github.com/UnicomAI/wanwu/internal/bff-service/model/request"

type SkillDetail struct {
	SkillId       string         `json:"skillId"`       // 模板ID
	Name          string         `json:"name"`          // 模板名称
	Avatar        request.Avatar `json:"avatar"`        // 模板头像
	Author        string         `json:"author"`        // 作者
	Desc          string         `json:"desc"`          // 模板描述
	SkillMarkdown string         `json:"skillMarkdown"` // 模板markdown预览
}

type SkillInfo struct {
	SkillId   string         `json:"skillId"`   // skillId
	SkillName string         `json:"skillName"` // 名称
	SkillType string         `json:"skillType"` // 类型
	Desc      string         `json:"desc"`      // 描述
	Author    string         `json:"author"`    // 作者
	Avatar    request.Avatar `json:"avatar"`    // 图标
}

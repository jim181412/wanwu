package orm

import (
	"context"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/orm/sqlopt"
)

func (c *Client) CreateAssistantSkill(ctx context.Context, assistantId uint32, skillId, skillType, userId, orgId string) *err_code.Status {
	var count int64
	if err := sqlopt.SQLOptions(
		sqlopt.WithAssistantID(assistantId),
		sqlopt.WithSkillId(skillId),
		sqlopt.WithSkillType(skillType),
	).Apply(c.db.WithContext(ctx)).Model(&model.AssistantSkill{}).
		Count(&count).Error; err != nil {
		return toErrStatus("assistant_skill_create", err.Error())
	}
	if count > 0 {
		return toErrStatus("assistant_skill_create", "skill already exists")
	}

	err := c.db.WithContext(ctx).Create(&model.AssistantSkill{
		AssistantId: assistantId,
		SkillId:     skillId,
		SkillType:   skillType,
		Enable:      true,
		UserId:      userId,
		OrgId:       orgId,
	}).Error

	if err != nil {
		return toErrStatus("assistant_skill_create", err.Error())
	}
	return nil
}

func (c *Client) DeleteAssistantSkill(ctx context.Context, assistantId uint32, skillId, skillType string) *err_code.Status {
	if err := sqlopt.SQLOptions(
		sqlopt.WithAssistantID(assistantId),
		sqlopt.WithSkillId(skillId),
		sqlopt.WithSkillType(skillType),
	).Apply(c.db.WithContext(ctx)).Delete(&model.AssistantSkill{}).Error; err != nil {
		return toErrStatus("assistant_skill_delete", err.Error())
	}
	return nil
}

func (c *Client) GetAssistantSkillById(ctx context.Context, assistantId uint32, skillId, skillType string) (*model.AssistantSkill, *err_code.Status) {
	var skill model.AssistantSkill
	if err := sqlopt.SQLOptions(
		sqlopt.WithAssistantID(assistantId),
		sqlopt.WithSkillId(skillId),
		sqlopt.WithSkillType(skillType),
	).Apply(c.db.WithContext(ctx)).First(&skill).Error; err != nil {
		return nil, toErrStatus("assistant_skill_get", err.Error())
	}
	return &skill, nil
}

func (c *Client) GetAssistantSkillList(ctx context.Context, assistantId uint32) ([]*model.AssistantSkill, *err_code.Status) {
	var skills []*model.AssistantSkill
	if err := sqlopt.SQLOptions(
		sqlopt.WithAssistantID(assistantId),
	).Apply(c.db.WithContext(ctx)).
		Find(&skills).Error; err != nil {
		return nil, toErrStatus("assistant_skill_get_list", err.Error())
	}
	return skills, nil
}

func (c *Client) UpdateAssistantSkillEnable(ctx context.Context, assistantId uint32, skillId, skillType string, enable bool) *err_code.Status {
	result := sqlopt.SQLOptions(
		sqlopt.WithAssistantID(assistantId),
		sqlopt.WithSkillId(skillId),
		sqlopt.WithSkillType(skillType),
	).Apply(c.db.WithContext(ctx)).
		Model(&model.AssistantSkill{}).
		Updates(map[string]interface{}{
			"enable": enable,
		})
	if result.Error != nil {
		return toErrStatus("assistant_skill_update", result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return toErrStatus("assistant_skill_update", "skill not exists")
	}

	return nil
}

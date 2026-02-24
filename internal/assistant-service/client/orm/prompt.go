package orm

import (
	"context"
	"errors"
	"strconv"
	"strings"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/pkg/util"
	"gorm.io/gorm"
)

func (c *Client) CreateCustomPrompt(ctx context.Context, avatarPath, name, desc, prompt, userId, orgID string) (string, *err_code.Status) {
	// 检查是否已存在相同名称的自定义提示
	var count int64
	if err := sqlopt.SQLOptions(
		sqlopt.WithCustomPromptName(name),
		sqlopt.WithUserId(userId),
		sqlopt.WithOrgID(orgID),
	).Apply(c.db.WithContext(ctx)).Model(&model.CustomPrompt{}).
		Count(&count).Error; err != nil {
		return "", toErrStatus("assistant_custom_prompt_create", err.Error())
	}
	if count > 0 {
		return "", toErrStatus("assistant_custom_prompt_exist")
	}

	// 创建记录
	customPrompt := model.CustomPrompt{
		AvatarPath: avatarPath,
		Name:       name,
		Desc:       desc,
		Prompt:     prompt,
		UserId:     userId,
		OrgId:      orgID,
	}

	err := c.db.WithContext(ctx).Create(&customPrompt).Error
	if err != nil {
		return "", toErrStatus("assistant_custom_prompt_create", err.Error())
	}

	return strconv.Itoa(int(customPrompt.ID)), nil
}

func (c *Client) DeleteCustomPrompt(ctx context.Context, customPromptID uint32) *err_code.Status {
	// 删除记录
	if err := sqlopt.WithID(customPromptID).Apply(c.db.WithContext(ctx)).Delete(&model.CustomPrompt{}).Error; err != nil {
		return toErrStatus("assistant_custom_prompt_delete", err.Error())
	}
	return nil
}

func (c *Client) UpdateCustomPrompt(ctx context.Context, info *assistant_service.CustomPromptUpdateReq) *err_code.Status {
	customPromptID := util.MustU32(info.CustomPromptId)
	// 检查记录是否存在
	var existingPrompt model.CustomPrompt
	err := sqlopt.WithID(customPromptID).Apply(c.db.WithContext(ctx)).First(&existingPrompt).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return toErrStatus("assistant_custom_prompt_get_err", "custom prompt does not exist")
		}
		return toErrStatus("assistant_custom_prompt_update_err", err.Error())
	}

	// 检查重名
	if info.Name != existingPrompt.Name {
		var count int64
		if err := sqlopt.SQLOptions(
			sqlopt.WithCustomPromptNotID(customPromptID),
			sqlopt.WithCustomPromptName(info.Name),
			sqlopt.WithUserId(existingPrompt.UserId),
			sqlopt.WithOrgID(existingPrompt.OrgId),
		).Apply(c.db.WithContext(ctx)).Model(&model.CustomPrompt{}).
			Count(&count).Error; err != nil {
			return toErrStatus("assistant_custom_prompt_update_err", err.Error())
		}
		if count > 0 {
			return toErrStatus("assistant_custom_prompt_update_err", "name already exists")
		}
	}

	// 执行更新操作
	updateData := map[string]interface{}{
		"avatar_path": info.AvatarPath,
		"name":        info.Name,
		"desc":        info.Desc,
		"prompt":      info.Prompt,
	}

	err = sqlopt.WithID(customPromptID).Apply(c.db.WithContext(ctx)).Model(&model.CustomPrompt{}).Updates(updateData).Error
	if err != nil {
		return toErrStatus("assistant_custom_prompt_update_err", err.Error())
	}

	return nil
}

func (c *Client) GetCustomPrompt(ctx context.Context, customPromptID uint32) (*model.CustomPrompt, *err_code.Status) {
	// 查询记录
	var customPrompt model.CustomPrompt
	if err := sqlopt.WithID(customPromptID).Apply(c.db.WithContext(ctx)).First(&customPrompt).Error; err != nil {
		return nil, toErrStatus("assistant_custom_prompt_get_err", err.Error())
	}
	return &customPrompt, nil
}

func (c *Client) GetCustomPromptList(ctx context.Context, userID, orgID string, name string) ([]*model.CustomPrompt, int64, *err_code.Status) {
	// 查询记录
	var customPrompts []*model.CustomPrompt
	if err := sqlopt.SQLOptions(
		sqlopt.WithUserId(userID),
		sqlopt.WithOrgID(orgID),
		sqlopt.WithCustomPromptLikeName(name),
	).Apply(c.db.WithContext(ctx)).Order("updated_at DESC").Find(&customPrompts).Error; err != nil {
		return nil, 0, toErrStatus("assistant_custom_prompt_get_list_err", err.Error())
	}
	return customPrompts, int64(len(customPrompts)), nil
}

func (c *Client) CopyCustomPrompt(ctx context.Context, customPromptID uint32, userId, orgID string) (string, *err_code.Status) {
	// 查询记录
	var customPrompt model.CustomPrompt
	if err := sqlopt.WithID(customPromptID).Apply(c.db.WithContext(ctx)).First(&customPrompt).Error; err != nil {
		return "", toErrStatus("assistant_custom_prompt_copy_err", err.Error())
	}

	// 名称前缀
	prefix := customPrompt.Name + "_"

	// 查询所有以“原名称_”为前缀的名称
	var existingNames []string
	err := c.db.WithContext(ctx).Model(&model.CustomPrompt{}).
		Where("name LIKE ?", prefix+"%").
		Pluck("name", &existingNames).Error
	if err != nil {
		return "", toErrStatus("assistant_custom_prompt_copy_err", err.Error())
	}

	// 解析名称
	maxNum := 0
	for _, name := range existingNames {
		numStr := strings.TrimPrefix(name, prefix)
		num, err := strconv.Atoi(numStr)
		if err != nil {
			continue
		}
		if num > maxNum {
			maxNum = num
		}
	}

	// 生成新名称
	newName := prefix + strconv.Itoa(maxNum+1)

	// 创建记录
	customPrompt.ID = 0
	customPrompt.Name = newName
	if err := c.db.WithContext(ctx).Create(&customPrompt).Error; err != nil {
		return "", toErrStatus("assistant_custom_prompt_copy_err", err.Error())
	}

	return util.Int2Str(customPrompt.ID), nil

}

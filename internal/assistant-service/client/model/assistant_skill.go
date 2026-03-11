package model

type AssistantSkill struct {
	ID          uint32 `gorm:"primarykey;column:id"`
	AssistantId uint32 `gorm:"column:assistant_id;index:idx_assistant_skill_assistant_id;comment:智能体id"`
	SkillId     string `gorm:"column:skill_id;index:idx_assistant_skill_skill_id;comment:skill id"`
	SkillType   string `gorm:"column:skill_type;comment:skill类型 builtin:内建 custom:自定义"`
	Enable      bool   `gorm:"column:enable;comment:是否启用"`
	UserId      string `gorm:"column:user_id;index:idx_assistant_skill_user_id;comment:用户id"`
	OrgId       string `gorm:"column:org_id;index:idx_assistant_skill_org_id;comment:组织id"`
	CreatedAt   int64  `gorm:"autoCreateTime:milli;comment:创建时间"`
	UpdatedAt   int64  `gorm:"autoUpdateTime:milli;comment:更新时间"`
}

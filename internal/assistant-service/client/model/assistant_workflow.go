package model

type AssistantWorkflow struct {
	ID          uint32 `gorm:"primarykey;column:id"`
	AssistantId uint32 `gorm:"column:assistant_id;index:idx_assistant_workflow_assistant_id;comment:智能体id"`
	WorkflowId  string `gorm:"column:workflow_id;index:idx_assistant_workflow_workflow_id;comment:工作流id"`
	Enable      bool   `gorm:"column:enable;comment:是否启用"`
	UserId      string `gorm:"column:user_id;index:idx_assistant_workflow_user_id;comment:用户id"`
	OrgId       string `gorm:"column:org_id;index:idx_assistant_workflow_org_id;comment:组织id"`
	CreatedAt   int64  `gorm:"autoCreateTime:milli;comment:创建时间"`
	UpdatedAt   int64  `gorm:"autoUpdateTime:milli;comment:更新时间"`
}

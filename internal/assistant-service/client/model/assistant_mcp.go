package model

type AssistantMCP struct {
	ID          uint32 `gorm:"primarykey;column:id"`
	AssistantId uint32 `gorm:"column:assistant_id;index:idx_assistant_mcp_assistant_id;comment:智能体id"`
	MCPId       string `gorm:"column:mcp_id;index:idx_assistant_mcp_mcp_id;comment:MCP id"`
	MCPType     string `gorm:"column:mcp_type;comment:MCP类型"`
	ActionName  string `gorm:"column:action_name;comment:操作名称"`
	Enable      bool   `gorm:"column:enable;comment:是否启用"`
	UserId      string `gorm:"column:user_id;index:idx_assistant_mcp_user_id;comment:用户id"`
	OrgId       string `gorm:"column:org_id;index:idx_assistant_mcp_org_id;comment:组织id"`
	CreatedAt   int64  `gorm:"autoCreateTime:milli;comment:创建时间"`
	UpdatedAt   int64  `gorm:"autoUpdateTime:milli;comment:更新时间"`
}

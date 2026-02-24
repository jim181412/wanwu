package request

type AgentChatBaseReq struct {
	Input          string   `json:"input" validate:"required"`
	UserId         string   `json:"userId"  validate:"required"`
	OrgId          string   `json:"orgId"  validate:"required"`
	UploadFile     []string `json:"uploadFile"`
	Stream         bool     `json:"stream"`
	Draft          bool     `json:"draft"`
	ConversationId string   `json:"conversationId"` //会话ID
}

type AgentConfig struct {
	AgentId          string //智能体的id
	AgentName        string //智能体名称
	AgentAvatar      string //智能体头像
	AgentChatContext *AgentChatContext
}

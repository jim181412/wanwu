package response

type CreateSkillConversationResp struct {
	ConversationId string `json:"conversationId"` // 会话ID
}

type SkillConversationItem struct {
	ConversationId string `json:"conversationId"` // 会话ID
	Title          string `json:"title"`          // 会话标题
	CreatedAt      string `json:"createdAt"`      // 创建时间
}

type SkillConversationDetailInfo struct {
	ConversationDetailInfo
	ResponseFiles []*AssistantResponseFile `json:"responseFiles"`
}

type SkillConversationSSEData struct {
	ConversationSSEData
	ResponseFiles []*AssistantResponseFile `json:"responseFiles"`
}

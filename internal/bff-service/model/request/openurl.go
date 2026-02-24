package request

type UrlConversationCreateRequest struct {
	Prompt string `json:"prompt"  validate:"required"`
}

func (c *UrlConversationCreateRequest) Check() error { return nil }

type UrlConversationIdRequest struct {
	ConversationId string `json:"conversationId" form:"conversationId"  validate:"required"`
}

func (c *UrlConversationIdRequest) Check() error { return nil }

type UrlConversionStreamRequest struct {
	ConversationId string `json:"conversationId" form:"conversionId"`
	Prompt         string `json:"prompt" form:"prompt"  validate:"required"`
}

func (c *UrlConversionStreamRequest) Check() error {
	return nil
}

type UrlQuestionRecommendRequest struct {
	ConversationId string `json:"conversationId" form:"conversionId"`
	Query          string `json:"query" form:"query"  validate:"required"`
	CommonCheck
}

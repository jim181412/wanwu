package constant

// openapi type
const (
	OpenAPITypeChatflow  = "chatflow"  // 对话问答
	OpenAPITypeWorkflow  = "workflow"  // 工作流
	OpenAPITypeAgent     = "agent"     // 智能体
	OpenAPITypeRag       = "rag"       // 文本问答
	OpenAPITypeKnowledge = "knowledge" // 知识库
)

// app type
const (
	AppTypeAgent     = "agent"     // 智能体
	AppTypeRag       = "rag"       // 文本问答
	AppTypeWorkflow  = "workflow"  // 工作流
	AppTypeChatflow  = "chatflow"  // 对话流
	AppTypeMCPServer = "mcpserver" // mcp server
)

// app publish type
const (
	AppPublishPublic       = "public"       // 系统公开发布
	AppPublishOrganization = "organization" // 组织公开发布
	AppPublishPrivate      = "private"      // 私密发布
)

// tool type
const (
	ToolTypeBuiltIn = "builtin" // 内置工具
	ToolTypeCustom  = "custom"  // 自定义工具
)

// mcp type
const (
	MCPTypeMCP       = "mcp"       // mcp
	MCPTypeMCPServer = "mcpserver" // mcp server
)

// mcp server tool type
const (
	MCPServerToolTypeCustomTool  = "custom"  // 自定义工具
	MCPServerToolTypeBuiltInTool = "builtin" // 内置工具
	MCPServerToolTypeOpenAPI     = "openapi" // 用户导入的openapi
)

// agent category
const (
	AgentCategorySingle = 1
	AgentCategoryMulti  = 2
)

// conversation type
const (
	ConversationTypeWebURL    = "openurl"   // openurl
	ConversationTypePublished = "published" // 已发布
	ConversationTypeDraft     = "draft"     // 草稿
	ConversationTypeOpenAPI   = "openapi"   // openapi
)

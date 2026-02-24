package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/UnicomAI/wanwu/internal/agent-service/model/request"
	mcp_client "github.com/UnicomAI/wanwu/internal/agent-service/service/mcp-client"
	"github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/components/tool"
	"github.com/mark3labs/mcp-go/client"
	mcpTypes "github.com/mark3labs/mcp-go/mcp"
)

type MCPServerInfo struct {
	Transport    string   `json:"transport"`
	URL          string   `json:"url"`
	ToolNameList []string `json:"toolNameList"`
}

func createMCPClient(ctx context.Context, url string) (client.MCPClient, error) {
	// 使用 NewSSEMCPClient 创建客户端
	mcpClient, err := client.NewSSEMCPClient(url)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSE MCP client: %w", err)
	}
	retryMcpClient := mcp_client.NewDefaultRetryMcpClient(mcpClient)

	// 启动客户端
	err = retryMcpClient.Start(ctx)
	if err != nil {
		_ = retryMcpClient.Close()
		return nil, fmt.Errorf("failed to start SSE MCP client: %w", err)
	}

	// 初始化 MCP 客户端
	initCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	initRequest := mcpTypes.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcpTypes.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcpTypes.Implementation{
		Name:    "eino-mcp-client",
		Version: "0.1.0",
	}
	initRequest.Params.Capabilities = mcpTypes.ClientCapabilities{}

	_, err = retryMcpClient.Initialize(initCtx, initRequest)
	if err != nil {
		_ = retryMcpClient.Close()
		return nil, fmt.Errorf("failed to initialize MCP client: %w", err)
	}

	log.Println("SSE MCP client initialized successfully")
	return retryMcpClient, nil
}

func GetToolsFromMCPServers(ctx context.Context, toolParamsList []*request.MCPToolInfo) ([]tool.BaseTool, error) {
	if len(toolParamsList) == 0 {
		return nil, nil
	}

	var allTools []tool.BaseTool

	for _, serverInfo := range toolParamsList {
		log.Printf("Connecting to MCP server: %v", serverInfo)

		mcpClient, err := createMCPClient(ctx, serverInfo.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to create MCP client for %v: %v", serverInfo, err)
		}
		// 注意:不要在这里关闭客户端,因为工具在后续使用时还需要这个连接
		// defer mcpClient.Close()

		tools, err := mcp.GetTools(ctx, &mcp.Config{
			Cli:          mcpClient,
			ToolNameList: serverInfo.ToolNameList,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get tools from %v: %v", serverInfo, err)
		}

		log.Printf("Loaded %d tools from %v", len(tools), serverInfo)
		allTools = append(allTools, tools...)
	}

	return allTools, nil
}

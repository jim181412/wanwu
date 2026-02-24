package response

type AgentEventType int
type SubEventStatus int

const (
	MainAgentEventType = 0 //单智能体事件/多智能体主智能体
	SubAgentEventType  = 1 //子智能体事件

	EventStartStatus   SubEventStatus = 1 //开始事件
	EventProcessStatus SubEventStatus = 2 //输出中
	EventEndStatus     SubEventStatus = 3 //结束事件
	EventFailStatus    SubEventStatus = 4 //子智能体失败
)

type SubEventData struct {
	Status   SubEventStatus `json:"status"`
	Id       string         `json:"id"`
	Name     string         `json:"name"`
	Profile  string         `json:"profile"`
	TimeCost string         `json:"timeCost"`
	ParentId string         `json:"parentId"`
}

func BuildStartSubAgent(respContext *AgentChatRespContext) *SubEventData {
	return StartSubAgent(respContext.CurrentAgentId, respContext.CurrentAgent, respContext.CurrentAgentAvatar)
}

func BuildProcessSubAgent(respContext *AgentChatRespContext) *SubEventData {
	return ProcessSubAgent(respContext.CurrentAgentId, respContext.CurrentAgent, respContext.CurrentAgentAvatar)
}

func BuildEndSubAgent(respContext *AgentChatRespContext, timeCost string) *SubEventData {
	return EndSubAgent(respContext.CurrentAgentId, respContext.CurrentAgent, respContext.CurrentAgentAvatar, timeCost)
}

func StartSubAgent(agentId, agentName, agentAvatar string) *SubEventData {
	return &SubEventData{
		Status:  EventStartStatus,
		Id:      agentId,
		Name:    agentName,
		Profile: agentAvatar,
	}
}

func ProcessSubAgent(agentId, agentName, agentAvatar string) *SubEventData {
	if len(agentId) == 0 || len(agentName) == 0 {
		return nil
	}
	return &SubEventData{
		Status:  EventProcessStatus,
		Id:      agentId,
		Name:    agentName,
		Profile: agentAvatar,
	}
}

func EndSubAgent(agentId, agentName, agentAvatar, timeCost string) *SubEventData {
	return &SubEventData{
		Status:   EventEndStatus,
		Id:       agentId,
		Name:     agentName,
		Profile:  agentAvatar,
		TimeCost: timeCost,
	}
}

package message_builder

import (
	"context"
	"encoding/json"

	rag_manage_service "github.com/UnicomAI/wanwu/internal/rag-service/service/rag-manage-service"
	"github.com/UnicomAI/wanwu/pkg/log"
)

type QAFinishBuilder struct {
}

func (QAFinishBuilder) MessageType() RagMessageType {
	return QAFinish
}

func (QAFinishBuilder) Build(ctx context.Context, ragContext *RagContext) *RagEvent {
	if len(ragContext.QAIds) == 0 {
		return &RagEvent{Skip: true}
	}
	params, err := rag_manage_service.BuildQaHitParams(ragContext.Req, ragContext.Rag, ragContext.KnowledgeIDToName, ragContext.QAIds)
	if err != nil {
		log.Errorf("BuildQaHitParams error %s", err)
		return &RagEvent{
			Skip: true,
		}
	}
	search, err := rag_manage_service.RagQASearch(ctx, params)
	if err != nil {
		log.Errorf("RagQASearch error %s", err)
		return &RagEvent{
			Skip: true,
		}
	}

	if len(search.Data.SearchList) > 0 {
		list := search.Data.SearchList
		searchData := list[0]
		var resultSearchList = make([]interface{}, 0)
		list = list[1:]
		for _, item := range list {
			resultSearchList = append(resultSearchList, item)
		}

		ragMessage := &RagMessage{
			Code:    0,
			Message: "success",
			MsgId:   ragContext.MessageId,
			MsgType: QAFinish,
			History: make([]*RagHistory, 0),
			Data: &RagData{
				Output:     searchData.Answer,
				SearchList: resultSearchList,
			},
			Finish: 1,
		}
		data, err := json.Marshal(ragMessage)
		if err != nil {
			return &RagEvent{
				Error: err,
			}
		}
		return &RagEvent{
			Stop:    true,
			Message: []string{LineData(data)},
		}
	}

	if len(ragContext.KnowledgeIds) == 0 {
		ragMessage := &RagMessage{
			Code:    0,
			MsgId:   ragContext.MessageId,
			Message: "success",
			MsgType: QAFinish,
			History: make([]*RagHistory, 0),
			Data: &RagData{
				Output:     "根据已知信息，无法回答您的问题。",
				SearchList: make([]interface{}, 0),
			},
			Finish: 1,
		}
		data, err := json.Marshal(ragMessage)
		if err != nil {
			return &RagEvent{
				Error: err,
			}
		}
		return &RagEvent{
			Stop:    true,
			Message: []string{LineData(data)},
		}
	}
	return &RagEvent{
		Skip: true,
	}
}

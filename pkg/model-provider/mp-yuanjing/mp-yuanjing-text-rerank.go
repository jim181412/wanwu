package mp_yuanjing

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/UnicomAI/wanwu/pkg/log"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	"github.com/UnicomAI/wanwu/pkg/util"
)

type Rerank struct {
	ApiKey      string `json:"apiKey"`      // ApiKey
	EndpointUrl string `json:"endpointUrl"` // 推理url
	ContextSize *int   `json:"contextSize"` // 上下文长度
}

func (cfg *Rerank) Tags() []mp_common.Tag {
	tags := []mp_common.Tag{
		{
			Text: mp_common.TagTextRerank,
		},
	}
	tags = append(tags, mp_common.GetTagsByContentSize(cfg.ContextSize)...)
	return tags
}

func (cfg *Rerank) NewReq(req *mp_common.TextRerankReq) (mp_common.ITextRerankReq, error) {
	m := map[string]interface{}{}
	instruction := "Given a web search query, retrieve relevant passages that answer the query"
	if req.Instruction == nil {
		m["instruction"] = instruction
	}
	m["query"] = req.Query
	m["documents"] = req.Documents
	return mp_common.NewRerankReq(m), nil
}

func (cfg *Rerank) Rerank(ctx context.Context, req mp_common.ITextRerankReq, headers ...mp_common.Header) (mp_common.ITextRerankResp, error) {
	b, err := mp_common.Rerank(ctx, "yuanjing", cfg.ApiKey, cfg.rerankUrl(), req.Data(), headers...)
	if err != nil {
		return nil, err
	}
	return &textRerankResp{raw: string(b)}, nil
}

func (cfg *Rerank) rerankUrl() string {
	ret, _ := url.JoinPath(cfg.EndpointUrl, "/rerank")
	return ret
}

// --- textRerankResp ---

type textRerankResp struct {
	raw     string
	Results []mp_common.Result `json:"results"`
}

func (resp *textRerankResp) String() string {
	return resp.raw
}

func (resp *textRerankResp) Data() (interface{}, bool) {
	ret := []map[string]interface{}{}
	if err := json.Unmarshal([]byte(resp.raw), &ret); err != nil {
		log.Errorf("yuanjing rerank resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}
	return ret, true
}

func (resp *textRerankResp) ConvertResp() (*mp_common.RerankResp, bool) {
	if err := json.Unmarshal([]byte(resp.raw), resp); err != nil {
		log.Errorf("yuanjing rerank resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}
	if err := util.Validate(resp); err != nil {
		log.Errorf("yuanjing rerank resp validate err: %v", err)
		return nil, false
	}
	return &mp_common.RerankResp{
		Results: resp.Results,
	}, true
}

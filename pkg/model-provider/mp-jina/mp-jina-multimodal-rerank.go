package mp_jina

import (
	"context"
	"fmt"
	"net/url"

	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
)

type MultiModalRerank struct {
	ApiKey              string   `json:"apiKey"`                     // ApiKey
	EndpointUrl         string   `json:"endpointUrl"`                // 推理url
	ContextSize         *int     `json:"contextSize"`                // 上下文长度
	MaxTextLength       *int64   `json:"maxTextLength"`              // 最大文本长度
	MaxImageSize        *int64   `json:"maxImageSize,omitempty"`     // 最大图片大小限制
	MaxVideoClipSize    *int64   `json:"maxVideoClipSize,omitempty"` // 最大视频片大小限制
	SupportFileTypes    []string `json:"supportFileTypes"`           // 支持的文件类型列表
	SupportImageInQuery bool     `json:"supportImageInQuery"`        // 是否支持query中传图片格式
}

func (cfg *MultiModalRerank) Tags() []mp_common.Tag {
	tags := []mp_common.Tag{
		{
			Text: mp_common.TagMultiModalRerank,
		},
	}
	tags = append(tags, mp_common.GetTagsByContentSize(cfg.ContextSize)...)
	return tags
}

func (cfg *MultiModalRerank) NewReq(req *mp_common.MultiModalRerankReq) (mp_common.IMultiModalRerankReq, error) {
	m, err := req.Data()
	if err != nil {
		return nil, err
	}
	queryVal := m["query"]
	switch q := queryVal.(type) {
	case string:
		m["query"] = q
	case map[string]interface{}:
		if textVal, ok := q["text"].(string); ok && textVal != "" {
			m["query"] = textVal
		} else {
			return nil, fmt.Errorf("query对象格式无效，必须包含非空text字符串字段")
		}
	default:
		return nil, fmt.Errorf("不支持的query类型: %T，仅支持字符串或{text:string}格式对象", q)
	}
	return mp_common.NewRerankReq(m), nil
}

func (cfg *MultiModalRerank) MultiModalRerank(ctx context.Context, req mp_common.IMultiModalRerankReq, headers ...mp_common.Header) (mp_common.IMultiModalRerankResp, error) {
	b, err := mp_common.MultiModalRerank(ctx, "jina", cfg.ApiKey, cfg.rerankUrl(), req.Data(), headers...)
	if err != nil {
		return nil, err
	}
	return mp_common.NewMultiModalRerankResp(string(b)), nil
}

func (cfg *MultiModalRerank) rerankUrl() string {
	ret, _ := url.JoinPath(cfg.EndpointUrl, "/rerank")
	return ret
}

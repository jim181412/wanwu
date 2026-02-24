package sse_util

import (
	"bufio"
	"context"
	"fmt"
	"net/http"

	"github.com/UnicomAI/wanwu/pkg/log"
	safe_go_util "github.com/UnicomAI/wanwu/pkg/safe-go-util"
	"google.golang.org/grpc"
)

type SSEReader[Res any] struct {
	BusinessKey    string
	Params         string
	StreamReceiver StreamReceiver[Res]
}

type StreamReceiver[Res any] interface {
	Receive() (*Res, error)
	Close() error
}

type HttpStreamReceiver[Res string] struct {
	httpResponse *http.Response
	reader       *bufio.Reader
}

func NewHttpStreamReceiver(httpResponse *http.Response) *HttpStreamReceiver[string] {
	return &HttpStreamReceiver[string]{
		httpResponse: httpResponse,
		reader:       bufio.NewReader(httpResponse.Body),
	}
}

func (htr *HttpStreamReceiver[string]) Receive() (*string, error) {
	line, err := htr.reader.ReadBytes('\n')
	if err != nil { //异常結束
		return nil, err
	}
	s := string(line)
	return &s, nil
}

func (htr *HttpStreamReceiver[string]) Close() error {
	return htr.httpResponse.Body.Close()
}

type GrpcStreamReceiver[Res any] struct {
	reader grpc.ServerStreamingClient[Res]
}

func NewGrpcStreamReceiver[Res any](stream grpc.ServerStreamingClient[Res]) *GrpcStreamReceiver[Res] {
	return &GrpcStreamReceiver[Res]{
		reader: stream,
	}
}

func (gr *GrpcStreamReceiver[Res]) Receive() (*Res, error) {
	return gr.reader.Recv()
}

func (gr *GrpcStreamReceiver[Res]) Close() error {
	return nil
}

// ReadStream 目专为HttpStreamReceiver使用
func (sr *SSEReader[Res]) ReadStream(ctx context.Context) (<-chan string, error) {
	return sr.ReadStreamWithBuilder(ctx, func(r *Res) string {
		return buildString(r)
	})
}

func (sr *SSEReader[Res]) ReadStreamWithBuilder(ctx context.Context, lineBuilder func(*Res) string) (<-chan string, error) {
	closer := func(ctx context.Context) {
		err1 := sr.StreamReceiver.Close()
		if err1 != nil {
			log.Errorf("%s close err: %v", sr.BusinessKey, err1)
		}
	}
	rawCh := safe_go_util.SafeChannelReceive[string](ctx, func(ctx context.Context, rawCh chan string) safe_go_util.ChannelReceiveResult[string] {
		content, err := sr.StreamReceiver.Receive()
		if err != nil {
			return safe_go_util.ChannelResult[string]("", err, sr.BusinessKey, sr.Params)
		}
		return safe_go_util.ChannelResult[string](lineBuilder(content), nil, sr.BusinessKey, sr.Params)
	}, closer)
	return rawCh, nil
}

func buildString[Res any](data *Res) string {
	if data == nil {
		return ""
	}
	// 使用类型开关处理不同类型
	switch v := any(*data).(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", *data)
	}
}

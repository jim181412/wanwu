package safe_go_util

import (
	"context"
	"io"

	"github.com/UnicomAI/wanwu/pkg/log"
)

type ChannelReceiveResult[T any] struct {
	ResultList []T
	Err        error
	Skip       bool
	Stop       bool
}

func SafeChannelReceive[T any](ctx context.Context, lineProcessor func(context.Context, chan T) ChannelReceiveResult[T], receiverCloser func(ctx context.Context)) <-chan T {
	rawCh := make(chan T, 128)
	var closer = func(ctx context.Context) {
		//执行receiver的close
		if receiverCloser != nil {
			receiverCloser(ctx)
		}
		//执行通道close
		close(rawCh)
	}
	// 起一个协程安全执行数据接收的方法
	SafeGo(safeCycleReceiveFunc(ctx, lineProcessor, rawCh, closer))
	return rawCh
}

func SafeCycleReceive[T any](ctx context.Context, lineProcessor func(context.Context, chan T) ChannelReceiveResult[T], rawCh chan T, closer func(context.Context)) error {
	defer func() {
		if closer != nil {
			closer(ctx)
		}
	}()
	for {
		resp := safeReceive(ctx, lineProcessor, rawCh)
		if len(resp.ResultList) > 0 {
			for _, result := range resp.ResultList {
				rawCh <- result
			}
		}
		if resp.Skip { //需要跳过
			continue
		}
		if stop(resp) { //错误或者正常结束
			return resp.Err
		}
	}
}

// ChannelResult 构造结果
func ChannelResult[T any](content T, err error, businessKey, params string) ChannelReceiveResult[T] {
	if err == io.EOF {
		log.Infof("%s stop, params: %s", businessKey, params)
		return ChannelStop[T]()
	}
	if err != nil {
		log.Errorf("%s recv err: %v", businessKey, err)
		return ChannelErr[T](nil, err)
	}
	return ChannelReceiveResult[T]{ResultList: []T{content}}
}

func ChannelSkip[T any]() ChannelReceiveResult[T] {
	return ChannelReceiveResult[T]{Skip: true}
}

func ChannelStop[T any]() ChannelReceiveResult[T] {
	return ChannelReceiveResult[T]{Stop: true}
}

func ChannelErr[T any](result []T, err error) ChannelReceiveResult[T] {
	return ChannelReceiveResult[T]{Err: err, ResultList: result}
}

func safeCycleReceiveFunc[T any](ctx context.Context, lineProcessor func(context.Context, chan T) ChannelReceiveResult[T], rawCh chan T, closer func(context.Context)) func() {
	return func() {
		_ = SafeCycleReceive(ctx, lineProcessor, rawCh, closer)
	}
}

// safeReceive 安全的channel读取，支持context取消
func safeReceive[T any](ctx context.Context, lineProcessor func(context.Context, chan T) ChannelReceiveResult[T], rawCh chan T) ChannelReceiveResult[T] {
	select {
	case <-ctx.Done():
		return ChannelReceiveResult[T]{Err: ctx.Err(), Stop: true}
	default:
		return lineProcessor(ctx, rawCh)
	}
}

// stopProcess 停止处理
func stop[T any](resp ChannelReceiveResult[T]) bool {
	if resp.Err != nil {
		log.Errorf("receive data error %s", resp.Err)
		return true
	}
	if resp.Stop {
		return true
	}
	return false
}

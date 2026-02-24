package safe_go_util

import (
	"sync"

	"github.com/UnicomAI/wanwu/pkg/util"
)

func SafeGo(f func()) {
	go func() {
		defer util.PrintPanicStack()
		f()
	}()
}

func SageGoWaitGroup(fnList ...func()) {
	if len(fnList) == 0 {
		return
	}
	wg := &sync.WaitGroup{}
	wg.Add(len(fnList))
	for _, f := range fnList {
		go func() {
			defer util.PrintPanicStack()
			defer wg.Done()
			f()
		}()
	}
	wg.Wait()
}

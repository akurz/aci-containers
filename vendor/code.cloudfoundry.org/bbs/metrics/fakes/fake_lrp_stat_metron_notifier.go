// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"os"
	"sync"
	"time"

	"code.cloudfoundry.org/bbs/metrics"
)

type FakeLRPStatMetronNotifier struct {
	RecordCellCountsStub        func(int, int)
	recordCellCountsMutex       sync.RWMutex
	recordCellCountsArgsForCall []struct {
		arg1 int
		arg2 int
	}
	RecordConvergenceDurationStub        func(time.Duration)
	recordConvergenceDurationMutex       sync.RWMutex
	recordConvergenceDurationArgsForCall []struct {
		arg1 time.Duration
	}
	RecordFreshDomainsStub        func([]string)
	recordFreshDomainsMutex       sync.RWMutex
	recordFreshDomainsArgsForCall []struct {
		arg1 []string
	}
	RecordLRPCountsStub        func(int, int, int, int, int, int, int, int, int, int)
	recordLRPCountsMutex       sync.RWMutex
	recordLRPCountsArgsForCall []struct {
		arg1  int
		arg2  int
		arg3  int
		arg4  int
		arg5  int
		arg6  int
		arg7  int
		arg8  int
		arg9  int
		arg10 int
	}
	RunStub        func(<-chan os.Signal, chan<- struct{}) error
	runMutex       sync.RWMutex
	runArgsForCall []struct {
		arg1 <-chan os.Signal
		arg2 chan<- struct{}
	}
	runReturns struct {
		result1 error
	}
	runReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeLRPStatMetronNotifier) RecordCellCounts(arg1 int, arg2 int) {
	fake.recordCellCountsMutex.Lock()
	fake.recordCellCountsArgsForCall = append(fake.recordCellCountsArgsForCall, struct {
		arg1 int
		arg2 int
	}{arg1, arg2})
	fake.recordInvocation("RecordCellCounts", []interface{}{arg1, arg2})
	fake.recordCellCountsMutex.Unlock()
	if fake.RecordCellCountsStub != nil {
		fake.RecordCellCountsStub(arg1, arg2)
	}
}

func (fake *FakeLRPStatMetronNotifier) RecordCellCountsCallCount() int {
	fake.recordCellCountsMutex.RLock()
	defer fake.recordCellCountsMutex.RUnlock()
	return len(fake.recordCellCountsArgsForCall)
}

func (fake *FakeLRPStatMetronNotifier) RecordCellCountsCalls(stub func(int, int)) {
	fake.recordCellCountsMutex.Lock()
	defer fake.recordCellCountsMutex.Unlock()
	fake.RecordCellCountsStub = stub
}

func (fake *FakeLRPStatMetronNotifier) RecordCellCountsArgsForCall(i int) (int, int) {
	fake.recordCellCountsMutex.RLock()
	defer fake.recordCellCountsMutex.RUnlock()
	argsForCall := fake.recordCellCountsArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeLRPStatMetronNotifier) RecordConvergenceDuration(arg1 time.Duration) {
	fake.recordConvergenceDurationMutex.Lock()
	fake.recordConvergenceDurationArgsForCall = append(fake.recordConvergenceDurationArgsForCall, struct {
		arg1 time.Duration
	}{arg1})
	fake.recordInvocation("RecordConvergenceDuration", []interface{}{arg1})
	fake.recordConvergenceDurationMutex.Unlock()
	if fake.RecordConvergenceDurationStub != nil {
		fake.RecordConvergenceDurationStub(arg1)
	}
}

func (fake *FakeLRPStatMetronNotifier) RecordConvergenceDurationCallCount() int {
	fake.recordConvergenceDurationMutex.RLock()
	defer fake.recordConvergenceDurationMutex.RUnlock()
	return len(fake.recordConvergenceDurationArgsForCall)
}

func (fake *FakeLRPStatMetronNotifier) RecordConvergenceDurationCalls(stub func(time.Duration)) {
	fake.recordConvergenceDurationMutex.Lock()
	defer fake.recordConvergenceDurationMutex.Unlock()
	fake.RecordConvergenceDurationStub = stub
}

func (fake *FakeLRPStatMetronNotifier) RecordConvergenceDurationArgsForCall(i int) time.Duration {
	fake.recordConvergenceDurationMutex.RLock()
	defer fake.recordConvergenceDurationMutex.RUnlock()
	argsForCall := fake.recordConvergenceDurationArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeLRPStatMetronNotifier) RecordFreshDomains(arg1 []string) {
	var arg1Copy []string
	if arg1 != nil {
		arg1Copy = make([]string, len(arg1))
		copy(arg1Copy, arg1)
	}
	fake.recordFreshDomainsMutex.Lock()
	fake.recordFreshDomainsArgsForCall = append(fake.recordFreshDomainsArgsForCall, struct {
		arg1 []string
	}{arg1Copy})
	fake.recordInvocation("RecordFreshDomains", []interface{}{arg1Copy})
	fake.recordFreshDomainsMutex.Unlock()
	if fake.RecordFreshDomainsStub != nil {
		fake.RecordFreshDomainsStub(arg1)
	}
}

func (fake *FakeLRPStatMetronNotifier) RecordFreshDomainsCallCount() int {
	fake.recordFreshDomainsMutex.RLock()
	defer fake.recordFreshDomainsMutex.RUnlock()
	return len(fake.recordFreshDomainsArgsForCall)
}

func (fake *FakeLRPStatMetronNotifier) RecordFreshDomainsCalls(stub func([]string)) {
	fake.recordFreshDomainsMutex.Lock()
	defer fake.recordFreshDomainsMutex.Unlock()
	fake.RecordFreshDomainsStub = stub
}

func (fake *FakeLRPStatMetronNotifier) RecordFreshDomainsArgsForCall(i int) []string {
	fake.recordFreshDomainsMutex.RLock()
	defer fake.recordFreshDomainsMutex.RUnlock()
	argsForCall := fake.recordFreshDomainsArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeLRPStatMetronNotifier) RecordLRPCounts(arg1 int, arg2 int, arg3 int, arg4 int, arg5 int, arg6 int, arg7 int, arg8 int, arg9 int, arg10 int) {
	fake.recordLRPCountsMutex.Lock()
	fake.recordLRPCountsArgsForCall = append(fake.recordLRPCountsArgsForCall, struct {
		arg1  int
		arg2  int
		arg3  int
		arg4  int
		arg5  int
		arg6  int
		arg7  int
		arg8  int
		arg9  int
		arg10 int
	}{arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10})
	fake.recordInvocation("RecordLRPCounts", []interface{}{arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10})
	fake.recordLRPCountsMutex.Unlock()
	if fake.RecordLRPCountsStub != nil {
		fake.RecordLRPCountsStub(arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10)
	}
}

func (fake *FakeLRPStatMetronNotifier) RecordLRPCountsCallCount() int {
	fake.recordLRPCountsMutex.RLock()
	defer fake.recordLRPCountsMutex.RUnlock()
	return len(fake.recordLRPCountsArgsForCall)
}

func (fake *FakeLRPStatMetronNotifier) RecordLRPCountsCalls(stub func(int, int, int, int, int, int, int, int, int, int)) {
	fake.recordLRPCountsMutex.Lock()
	defer fake.recordLRPCountsMutex.Unlock()
	fake.RecordLRPCountsStub = stub
}

func (fake *FakeLRPStatMetronNotifier) RecordLRPCountsArgsForCall(i int) (int, int, int, int, int, int, int, int, int, int) {
	fake.recordLRPCountsMutex.RLock()
	defer fake.recordLRPCountsMutex.RUnlock()
	argsForCall := fake.recordLRPCountsArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4, argsForCall.arg5, argsForCall.arg6, argsForCall.arg7, argsForCall.arg8, argsForCall.arg9, argsForCall.arg10
}

func (fake *FakeLRPStatMetronNotifier) Run(arg1 <-chan os.Signal, arg2 chan<- struct{}) error {
	fake.runMutex.Lock()
	ret, specificReturn := fake.runReturnsOnCall[len(fake.runArgsForCall)]
	fake.runArgsForCall = append(fake.runArgsForCall, struct {
		arg1 <-chan os.Signal
		arg2 chan<- struct{}
	}{arg1, arg2})
	fake.recordInvocation("Run", []interface{}{arg1, arg2})
	fake.runMutex.Unlock()
	if fake.RunStub != nil {
		return fake.RunStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.runReturns
	return fakeReturns.result1
}

func (fake *FakeLRPStatMetronNotifier) RunCallCount() int {
	fake.runMutex.RLock()
	defer fake.runMutex.RUnlock()
	return len(fake.runArgsForCall)
}

func (fake *FakeLRPStatMetronNotifier) RunCalls(stub func(<-chan os.Signal, chan<- struct{}) error) {
	fake.runMutex.Lock()
	defer fake.runMutex.Unlock()
	fake.RunStub = stub
}

func (fake *FakeLRPStatMetronNotifier) RunArgsForCall(i int) (<-chan os.Signal, chan<- struct{}) {
	fake.runMutex.RLock()
	defer fake.runMutex.RUnlock()
	argsForCall := fake.runArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeLRPStatMetronNotifier) RunReturns(result1 error) {
	fake.runMutex.Lock()
	defer fake.runMutex.Unlock()
	fake.RunStub = nil
	fake.runReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeLRPStatMetronNotifier) RunReturnsOnCall(i int, result1 error) {
	fake.runMutex.Lock()
	defer fake.runMutex.Unlock()
	fake.RunStub = nil
	if fake.runReturnsOnCall == nil {
		fake.runReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.runReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeLRPStatMetronNotifier) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.recordCellCountsMutex.RLock()
	defer fake.recordCellCountsMutex.RUnlock()
	fake.recordConvergenceDurationMutex.RLock()
	defer fake.recordConvergenceDurationMutex.RUnlock()
	fake.recordFreshDomainsMutex.RLock()
	defer fake.recordFreshDomainsMutex.RUnlock()
	fake.recordLRPCountsMutex.RLock()
	defer fake.recordLRPCountsMutex.RUnlock()
	fake.runMutex.RLock()
	defer fake.runMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeLRPStatMetronNotifier) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ metrics.LRPStatMetronNotifier = new(FakeLRPStatMetronNotifier)

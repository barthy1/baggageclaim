// This file was generated by counterfeiter
package fake_namespacer

import (
	"sync"

	"github.com/concourse/baggageclaim/uidjunk"
)

type FakeNamespacer struct {
	CacheKeyStub        func() string
	cacheKeyMutex       sync.RWMutex
	cacheKeyArgsForCall []struct{}
	cacheKeyReturns     struct {
		result1 string
	}
	NamespaceStub        func(rootfsPath string) error
	namespaceMutex       sync.RWMutex
	namespaceArgsForCall []struct {
		rootfsPath string
	}
	namespaceReturns struct {
		result1 error
	}
}

func (fake *FakeNamespacer) CacheKey() string {
	fake.cacheKeyMutex.Lock()
	fake.cacheKeyArgsForCall = append(fake.cacheKeyArgsForCall, struct{}{})
	fake.cacheKeyMutex.Unlock()
	if fake.CacheKeyStub != nil {
		return fake.CacheKeyStub()
	} else {
		return fake.cacheKeyReturns.result1
	}
}

func (fake *FakeNamespacer) CacheKeyCallCount() int {
	fake.cacheKeyMutex.RLock()
	defer fake.cacheKeyMutex.RUnlock()
	return len(fake.cacheKeyArgsForCall)
}

func (fake *FakeNamespacer) CacheKeyReturns(result1 string) {
	fake.CacheKeyStub = nil
	fake.cacheKeyReturns = struct {
		result1 string
	}{result1}
}

func (fake *FakeNamespacer) Namespace(rootfsPath string) error {
	fake.namespaceMutex.Lock()
	fake.namespaceArgsForCall = append(fake.namespaceArgsForCall, struct {
		rootfsPath string
	}{rootfsPath})
	fake.namespaceMutex.Unlock()
	if fake.NamespaceStub != nil {
		return fake.NamespaceStub(rootfsPath)
	} else {
		return fake.namespaceReturns.result1
	}
}

func (fake *FakeNamespacer) NamespaceCallCount() int {
	fake.namespaceMutex.RLock()
	defer fake.namespaceMutex.RUnlock()
	return len(fake.namespaceArgsForCall)
}

func (fake *FakeNamespacer) NamespaceArgsForCall(i int) string {
	fake.namespaceMutex.RLock()
	defer fake.namespaceMutex.RUnlock()
	return fake.namespaceArgsForCall[i].rootfsPath
}

func (fake *FakeNamespacer) NamespaceReturns(result1 error) {
	fake.NamespaceStub = nil
	fake.namespaceReturns = struct {
		result1 error
	}{result1}
}

var _ uidjunk.Namespacer = new(FakeNamespacer)

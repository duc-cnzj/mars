package services

import (
	"context"
	"sync"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type registryFunc = func(s grpc.ServiceRegistrar, app contracts.ApplicationInterface)

var (
	registryFuncs  = make([]registryFunc, 0)
	registryFuncMu = sync.RWMutex{}
)

func AddServerFunc(fn registryFunc) {
	registryFuncMu.Lock()
	defer registryFuncMu.Unlock()
	registryFuncs = append(registryFuncs, fn)
}

func ServerFuncs() []registryFunc {
	registryFuncMu.RLock()
	defer registryFuncMu.RUnlock()

	return registryFuncs
}

type endpointFunc func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error)

var (
	endpointFuncs  = make([]endpointFunc, 0)
	endpointFuncMu = sync.RWMutex{}
)

func AddEndpointFunc(fn endpointFunc) {
	endpointFuncMu.Lock()
	defer endpointFuncMu.Unlock()
	endpointFuncs = append(endpointFuncs, fn)
}

func EndpointFuncs() []endpointFunc {
	endpointFuncMu.RLock()
	defer endpointFuncMu.RUnlock()

	return endpointFuncs
}

package bridge

import (
	"context"
	"sync"

	clientapp "github.com/albe194e/albz/app/core-go/app"
	coreruntime "github.com/albe194e/albz/app/core-go/runtime"
)

type Bridge struct {
	service *coreruntime.Service

	sinkMu sync.RWMutex
	sink   EventSink
}

func New(options Options) (*Bridge, error) {
	service, err := coreruntime.New(context.Background(), coreruntime.Options{
		ProfileName: options.ProfileName,
		DataDir:     options.DataDir,
		ServerURL:   options.ServerURL,
	})
	if err != nil {
		return nil, err
	}

	b := &Bridge{
		service: service,
	}

	if service.Controller != nil {
		service.Controller.OnStateChanged = func() {
			b.emitStateChanged()
		}
	}

	return b, nil
}

func (b *Bridge) Close() error {
	if b == nil || b.service == nil {
		return nil
	}

	return b.service.Close()
}

func (b *Bridge) controller() *clientapp.Controller {
	if b == nil || b.service == nil || b.service.Controller == nil {
		panic("bridge controller is not initialized")
	}

	return b.service.Controller
}

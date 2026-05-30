package hotloader

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"sync"
	"time"
)

// ----------------------------------------------------------------------------
// Structures
// ----------------------------------------------------------------------------

type Impl struct {
	filename   string
	config     Loader
	ch         chan os.Signal
	end        chan bool
	done       chan bool
	mu         *sync.RWMutex
	beforeFunc func(l Loader)
	afterFunc  func(l Loader)
	errorFunc  func(e error)
}

// ----------------------------------------------------------------------------
// Constructor
// ----------------------------------------------------------------------------

func New(filename string, loader Loader, signals ...os.Signal) HotLoader {
	thiz := &Impl{
		filename:   filename,
		config:     loader,
		ch:         make(chan os.Signal, 1),
		end:        make(chan bool, 1),
		done:       make(chan bool),
		mu:         &sync.RWMutex{},
		beforeFunc: nil,
		afterFunc:  nil,
		errorFunc:  nil,
	}
	thiz.start(signals...)
	runtime.SetFinalizer(thiz, func(o *Impl) {
		o.stop()
	})
	return thiz
}

// ----------------------------------------------------------------------------
// Exported functions
// ----------------------------------------------------------------------------

func (impl *Impl) Load() error {
	temp, err := impl.load()
	if err == nil {
		impl.mu.Lock()
		impl.config = temp
		impl.mu.Unlock()
	}
	return err
}

func (impl *Impl) Get() Loader {
	impl.mu.RLock()
	defer impl.mu.RUnlock()
	return impl.getConfig()
}

func (impl *Impl) SetBeforeFunc(f func(l Loader)) {
	impl.beforeFunc = f
}

func (impl *Impl) SetAfterFunc(f func(l Loader)) {
	impl.afterFunc = f
}

func (impl *Impl) SetErrorFunc(f func(e error)) {
	impl.errorFunc = f
}

// ----------------------------------------------------------------------------
// Unexported functions
// ----------------------------------------------------------------------------

func (impl *Impl) load() (Loader, error) {
	if impl.config == nil {
		return nil, errors.New("config is nil")
	}
	val := reflect.ValueOf(impl.config)
	if val.Kind() == reflect.Pointer {
		val = reflect.Indirect(val)
	}
	newLoader := reflect.New(val.Type()).Interface().(Loader) //nolint:forcetypeassert
	newLoader.InitDefault()
	err := newLoader.Load(impl.filename)
	if err != nil {
		return newLoader, fmt.Errorf("failed to load file '%s': %w", impl.filename, err)
	}
	return newLoader, nil
}

func (impl *Impl) getConfig() Loader {
	return impl.config
}

func (impl *Impl) start(signals ...os.Signal) {
	if len(signals) == 0 {
		return
	}
	go func() {
		signal.Notify(impl.ch, signals...)
		defer func() {
			signal.Stop(impl.ch)
			signal.Reset(signals...)
			close(impl.done)
		}()
		for {
			select {
			case <-impl.ch:
				if impl.beforeFunc != nil {
					impl.beforeFunc(impl.config)
				}
				err := impl.Load()
				if err != nil {
					if impl.errorFunc != nil {
						impl.errorFunc(err)
					}
				} else {
					if impl.afterFunc != nil {
						impl.afterFunc(impl.config)
					}
				}
			case <-impl.end:
				return
			}
		}
	}()
}

func (impl *Impl) stop() {
	if impl.end == nil {
		return
	}
	select {
	case impl.end <- true:
		<-impl.done
	case <-time.After(time.Second):
		return
	}
}

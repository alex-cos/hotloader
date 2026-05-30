package hotloader

import (
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

func (thiz *Impl) Load() error {
	temp, err := thiz.load()
	if err == nil {
		thiz.mu.Lock()
		thiz.config = temp
		thiz.mu.Unlock()
	}
	return err
}

func (thiz *Impl) Get() Loader {
	thiz.mu.RLock()
	defer thiz.mu.RUnlock()
	return thiz.getConfig()
}

func (thiz *Impl) SetBeforeFunc(f func(l Loader)) {
	thiz.beforeFunc = f
}

func (thiz *Impl) SetAfterFunc(f func(l Loader)) {
	thiz.afterFunc = f
}

func (thiz *Impl) SetErrorFunc(f func(e error)) {
	thiz.errorFunc = f
}

// ----------------------------------------------------------------------------
// Unexported functions
// ----------------------------------------------------------------------------

func (thiz *Impl) load() (Loader, error) {
	val := reflect.ValueOf(thiz.config)
	if val.Kind() == reflect.Pointer {
		val = reflect.Indirect(val)
	}
	newLoader := reflect.New(val.Type()).Interface().(Loader) //nolint:forcetypeassert
	newLoader.InitDefault()
	err := newLoader.Load(thiz.filename)
	if err != nil {
		return newLoader, fmt.Errorf("failed to load file '%s': %w", thiz.filename, err)
	}
	return newLoader, nil
}

func (thiz *Impl) getConfig() Loader {
	return thiz.config
}

func (thiz *Impl) start(signals ...os.Signal) {
	if len(signals) == 0 {
		return
	}
	go func() {
		signal.Notify(thiz.ch, signals...)
		defer func() {
			signal.Stop(thiz.ch)
			signal.Reset(signals...)
			close(thiz.done)
		}()
		for {
			select {
			case <-thiz.ch:
				if thiz.beforeFunc != nil {
					thiz.beforeFunc(thiz.config)
				}
				err := thiz.Load()
				if err != nil {
					if thiz.errorFunc != nil {
						thiz.errorFunc(err)
					}
				} else {
					if thiz.afterFunc != nil {
						thiz.afterFunc(thiz.config)
					}
				}
			case <-thiz.end:
				return
			}
		}
	}()
}

func (thiz *Impl) stop() {
	if thiz.end == nil {
		return
	}
	select {
	case thiz.end <- true:
		<-thiz.done
	case <-time.After(time.Second):
		return
	}
}

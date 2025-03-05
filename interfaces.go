package hotloader

type Loader interface {
	Load(filename string) error
	InitDefault()
}

type HotLoader interface {
	Load() error
	Get() Loader
	SetBeforeFunc(f func(l Loader))
	SetAfterFunc(f func(l Loader))
	SetErrorFunc(f func(e error))
}

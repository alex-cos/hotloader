package hotloader

type MockLoader struct {
	NbLoad        int
	NbInitDefault int
}

func NewMockLoader() Loader {
	return &MockLoader{
		NbLoad:        0,
		NbInitDefault: 0,
	}
}

func (thiz *MockLoader) Load(filename string) error {
	thiz.NbLoad++
	return nil
}

func (thiz *MockLoader) InitDefault() {
	thiz.NbInitDefault++
}

type MockHotLoader struct {
}

func (thiz *MockHotLoader) Load() error {
	return nil
}

func (thiz *MockHotLoader) Get() Loader {
	return NewMockLoader()
}

func (thiz *MockHotLoader) SetBeforeFunc(f func(l Loader)) {
	// Nothing to do
}

func (thiz *MockHotLoader) SetAfterFunc(f func(l Loader)) {
	// Nothing to do
}

func (thiz *MockHotLoader) SetErrorFunc(f func(e error)) {
	// Nothing to do
}

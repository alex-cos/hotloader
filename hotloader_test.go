package hotloader_test

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/alex-cos/hotloader"
	"github.com/stretchr/testify/assert"
)

func TestNominal(t *testing.T) {
	t.Parallel()

	loader := hotloader.NewMockLoader()

	hl := hotloader.New("testfile", loader, syscall.Signal(10))
	err := hl.Load()
	assert.NoError(t, err)

	hl.SetBeforeFunc(func(l hotloader.Loader) {
		fmt.Println("Received signal to reload configuration")
	})
	hl.SetAfterFunc(func(l hotloader.Loader) {
		fmt.Println("Successfully reloaded configuration")
	})
	hl.SetErrorFunc(func(e error) {
		fmt.Printf("Failed to reload configuration: %v\n", e)
	})

	l := hl.Get().(*hotloader.MockLoader)
	assert.Equal(t, 1, l.NbLoad)
	assert.Equal(t, 1, l.NbInitDefault)

	if runtime.GOOS != "windows" {
		p, err := os.FindProcess(os.Getpid())
		assert.NoError(t, err)
		err = p.Signal(syscall.Signal(10))
		assert.NoError(t, err)

		time.Sleep(250 * time.Millisecond)

		l := hl.Get().(*hotloader.MockLoader)
		assert.Equal(t, 1, l.NbLoad)
		assert.Equal(t, 1, l.NbInitDefault)
	}
}

package Week03

import (
	"syscall"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// HTTPSever will run infinitely, the only way to exit normally is receive signal.
func TestNoError_SignalListenerExits(t *testing.T) {
	errChan := make(chan error)
	go func() {
		errChan <- Main()
	}()

	// SignalListenter must listen before we send signal to it.
	time.Sleep(time.Second)
	syscall.Kill(syscall.Getpid(), syscall.SIGUSR2)
	select {
	case err := <-errChan:
		assert.Nil(t, err)
	case <-time.After(time.Second):
		assert.Fail(t, "Main exits timeout")
	}
}

func TestError_HTTPServerExits(t *testing.T) {
	testErr = errors.New("mock error")
	testPanic = false
	errChan := make(chan error)
	go func() {
		errChan <- Main()
	}()

	select {
	case err := <-errChan:
		assert.True(t, errors.Is(err, testErr))
	case <-time.After(time.Second):
		assert.Fail(t, "Main exits timeout")
	}
}

func TestError_HTTPServerPanic(t *testing.T) {
	testErr = nil
	testPanic = true
	errChan := make(chan error)
	go func() {
		errChan <- Main()
	}()

	select {
	case err := <-errChan:
		assert.Contains(t, err.Error(), "errgroup: panic recovered")
	case <-time.After(time.Second):
		assert.Fail(t, "Main exits timeout")
	}
}

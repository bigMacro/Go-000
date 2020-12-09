package Week03

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

/*
基于 errgroup 实现一个 http server 的启动和关闭 ，以及 linux signal 信号的注册和处理，要保证能够一个退出，全部注销退出。
*/

type HTTPServer struct {
	server *http.Server
	ctx    context.Context
}

// for unit test
var testErr error
var testPanic bool

func NewHTTPServer(ctx context.Context) *HTTPServer {
	return &HTTPServer{
		server: &http.Server{
			Addr:    ":80",
			Handler: http.DefaultServeMux,
		},
		ctx: ctx,
	}

}

func (s *HTTPServer) start() error {
	errChan := make(chan error)
	go func() {
		// start server and recover if server panic
		defer func() {
			if r := recover(); r != nil {
				buf := make([]byte, 64<<10)
				buf = buf[:runtime.Stack(buf, false)]
				errChan <- fmt.Errorf("errgroup: panic recovered: %s\n%s", r, buf)
			}
		}()
		// for unit test
		{
			if testErr != nil {
				errChan <- testErr
				return
			}
			if testPanic {
				panic("testPanic")
			}
		}
		errChan <- s.server.ListenAndServe()
	}()

	select {
	case err := <-errChan:
		// HTTPServer exits with error, we should wrap it. And SignalListener will exit by cancel.
		return errors.Wrapf(err, "listen and serve failed")
	case <-s.ctx.Done():
		// HTTPServer exits by cancel context, we should shutdown server.
		ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
		return s.server.Shutdown(ctx)
	}
}

type SignalListener struct {
	ch  chan os.Signal
	ctx context.Context
}

func NewSignalListener(ctx context.Context) *SignalListener {
	return &SignalListener{
		ch:  make(chan os.Signal),
		ctx: ctx,
	}
}

func (s *SignalListener) notify(sg ...os.Signal) {
	signal.Notify(s.ch, sg...)
}

var errSignalExit = errors.New("SignalListener exits")

func (s *SignalListener) listen() error {
	select {
	case sg := <-s.ch:
		// HTTPServer exits by cancel context, so we must return error if SignalListener receives signal.
		return errors.WithMessagef(errSignalExit, "signal is %v", sg)
	case <-s.ctx.Done():
		// SignalListener exits by cancel context, so we needn't return error.
		fmt.Printf("SignalListener closed by others\n")
		return nil
	}
}

func Main() error {
	// all will be exited if http server is closed or receiving signals, so we need context to cancel the other.
	g, ctx := errgroup.WithContext(context.Background())

	{
		httpServer := NewHTTPServer(ctx)
		g.Go(func() error {
			return httpServer.start()
		})
	}

	{
		signalListner := NewSignalListener(ctx)
		signalListner.notify(syscall.SIGUSR2) // use SIGUSR2 for test
		g.Go(func() error {
			return signalListner.listen()
		})
	}

	err := g.Wait()
	// err will never be nil, if it is errSignlaExit then exits normally, otherwise not.
	if errors.Is(err, errSignalExit) {
		return nil
	}
	fmt.Printf("%v\n", err)
	return err
}

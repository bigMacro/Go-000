package main

import (
	v1 "Go-000/Week04/api/helloworld/v1"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

/*
按照自己的构想，写一个项目满足基本的目录结构和工程，代码需要包含对数据层、业务层、API 注册，以及 main 函数对于服务的注册和启动，
信号处理，使用 Wire 构建依赖。可以使用自己熟悉的框架。
*/

func main() {
	g, ctx := errgroup.WithContext(context.Background())
	g.Go(func() error {
		lis, err := net.Listen("tcp", ":666")
		if err != nil {
			return fmt.Errorf("failed to listen: %w", err)
		}
		s := grpc.NewServer()
		v1.RegisterGreeterServer(s, InitializeGreeterService())
		errChan := make(chan error)
		go func() {
			if err := s.Serve(lis); err != nil {
				errChan <- fmt.Errorf("failed to serve: %w", err)
			}
		}()

		select {
		case <-ctx.Done():
			// graceful stop
			stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			go func() {
				s.GracefulStop()
				cancel()
			}()

			select {
			case <-stopCtx.Done():
				if err := stopCtx.Err(); errors.Is(err, context.Canceled) {
					return nil
				}
				log.Printf("graceful stop server error: %v", stopCtx.Err())
				return nil
			}
		case <-errChan:
			return err
		}
	})

	errExit := errors.New("exited by signal")
	g.Go(func() error {
		sigChan := make(chan os.Signal)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
		select {
		case <-sigChan:
			return errExit
		case <-ctx.Done():
			return nil
		}
		return nil
	})

	if err := g.Wait(); !errors.Is(err, errExit) {
		log.Fatal(err)
	}
}

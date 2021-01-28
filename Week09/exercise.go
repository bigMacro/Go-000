package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

/*
用 Go 实现一个 tcp server ，用两个 goroutine 读写 conn， 两个 goroutine 通过 chan 可以传递 message，能够正确退出
*/

type ConnectionHandler struct {
	msgChan chan []byte
	conn    net.Conn
	closed  chan struct{}
	wg      sync.WaitGroup
}

func NewConnectionHandler(conn net.Conn) *ConnectionHandler {
	return &ConnectionHandler{
		msgChan: make(chan []byte),
		conn:    conn,
		closed:  make(chan struct{}),
	}
}

func (c *ConnectionHandler) Start() {
	c.wg.Add(2)
	go c.Read()
	go c.Write()
}

func (c *ConnectionHandler) Stop() {
	close(c.closed)
	c.conn.Close()
	c.wg.Wait()
}

func (c *ConnectionHandler) Read() {
	r := bufio.NewReader(c.conn)
	for {
		select {
		case <-c.closed:
			log.Printf("Read %v closes", c.conn.RemoteAddr())
			close(c.msgChan)
			c.wg.Done()
			return
		default:
			msgBytes := []byte{}
			readAll := false
			for {
				lineMsg, isPrefix, err := r.ReadLine()
				if err == io.EOF {
					log.Printf("read EOF from %v, Read exits", c.conn.RemoteAddr())
					close(c.msgChan)
					c.wg.Done()
					return
				}
				if err != nil {
					log.Printf("read from %v error: %v", c.conn.RemoteAddr(), err)
					break
				}
				msgBytes = append(msgBytes, lineMsg...)
				if !isPrefix {
					readAll = true
					break
				}
			}
			if readAll {
				c.msgChan <- msgBytes
			}
		}
	}
}

func (c *ConnectionHandler) Write() {
	for {
		select {
		case msg, ok := <-c.msgChan:
			if !ok {
				log.Printf("Write %v exits", c.conn.RemoteAddr())
				c.wg.Done()
				return
			}
			_, err := c.conn.Write([]byte(fmt.Sprintf("Received %s", msg)))
			if err != nil {
				log.Printf("write to %v error: %v", c.conn.RemoteAddr(), err)
			}
		}
	}
}

type ConnectionHandlerManger struct {
	handlers map[*ConnectionHandler]struct{}
	wg       sync.WaitGroup
}

func NewConnectionHandlerManger() *ConnectionHandlerManger {
	return &ConnectionHandlerManger{
		handlers: make(map[*ConnectionHandler]struct{}),
	}
}

func (c *ConnectionHandlerManger) Add(handler *ConnectionHandler) {
	c.wg.Add(1)
	c.handlers[handler] = struct{}{}
	handler.Start()
}

func (c *ConnectionHandlerManger) ClearAll() {
	for handler, _ := range c.handlers {
		go func() {
			handler.Stop()
			c.wg.Done()
		}()
	}
	c.wg.Wait()
}

func main() {
	connManager := NewConnectionHandlerManger()
	ctx, cancel := context.WithCancel(context.Background())
	g, _ := errgroup.WithContext(ctx)
	g.Go(func() error {
		listener, err := net.Listen("tcp", ":666")
		if err != nil {
			return errors.Wrapf(err, "listen on 666 failed")
		}
		go func() {
			for {
				conn, err := listener.Accept()
				if err != nil {
					log.Printf("accept error: %v", err)
				} else {
					handler := NewConnectionHandler(conn)
					connManager.Add(handler)
				}

				select {
				case <-ctx.Done():
					return
				default:
				}
			}
		}()

		select {
		case <-ctx.Done():
			listener.Close()
			connManager.ClearAll()
		}
		return nil
	})

	g.Go(func() error {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
		select {
		case s := <-c:
			log.Printf("receive signal %v", s)
			cancel()
		}
		return nil
	})

	g.Wait()
}

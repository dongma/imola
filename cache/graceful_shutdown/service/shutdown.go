package service

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

type Option func(*App)

// ShutdownCallback 采用context.Context来控制超时，而不是用time.After是因为
// - 超时本质上是使用这个回调的人控制的
// - 还希望用户知道，回调必须在一定时间内处理完毕，而且他必须显示处理超时错误
type ShutdownCallback func(ctx context.Context)

func WithShutdownCallbacks(callbacks ...ShutdownCallback) Option {
	return func(app *App) {
		app.cbs = callbacks
	}
}

type App struct {
	servers []*Server

	// 优雅退出整个超时时间，默认30秒
	shutdownTimeout time.Duration

	// 优雅退出等待处理已有请求时间，默认10秒
	waitTime time.Duration

	// 自定义回调超时时间，默认3秒
	cbTimeout time.Duration
	cbs       []ShutdownCallback
}

func NewApp(servers []*Server, opts ...Option) *App {
	res := &App{
		waitTime:        10 * time.Second,
		cbTimeout:       3 * time.Second,
		shutdownTimeout: 30 * time.Second,
		servers:         servers,
	}

	for _, opt := range opts {
		opt(res)
	}
	return res
}

func (app *App) StartAndServe() {
	for _, server := range app.servers {
		srv := server
		go func() {
			if err := srv.Start(); err != nil {
				if err == http.ErrServerClosed {
					log.Printf("服务器%s已关闭", srv.name)
				} else {
					log.Printf("服务器%s异常退出", srv.name)
				}
			}
		}()
	}

	// 从这里开始启动监听系统信号，ch := make(...) 首先创建一个接收系统信号的channel ch
	// 定义要监听的目标新型号 signals []os.Signal
	// 调用signal
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, signals...)
	<-ch
	go func() {
		select {
		case <-ch:
			log.Printf("强制退出")
			os.Exit(1)
		case <-time.After(app.shutdownTimeout):
			log.Printf("超时强制退出")
			os.Exit(1)
		}
	}()
	app.shutdown()
}

// shutdown 设计里面的执行步骤，具体每一步可以用time.Sleep来模拟
func (app *App) shutdown() {
	log.Println("开始关闭应用，停止接收新请求")
	for _, s := range app.servers {
		// note: 为什么这里不用并发控制（不用锁），也不用原子操作
		s.rejectReq()
	}
	log.Println("等待正在执行的请求完结")
	// 这里可以改造为实时统计正在处理的请求数量，为0 则下一步
	time.Sleep(app.waitTime)

	log.Println("开始关闭服务器")
	var wg sync.WaitGroup
	wg.Add(len(app.servers))
	for _, srv := range app.servers {
		srvCp := srv
		go func() {
			if err := srvCp.stop(); err != nil {
				log.Printf("关闭服务器失败%s \n", srvCp.name)
			}
			wg.Done()
		}()
	}

	wg.Wait()
	log.Println("开始执行自定义回调")
	wg.Add(len(app.cbs))
	for _, cb := range app.cbs {
		callbackFn := cb
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), app.cbTimeout)
			callbackFn(ctx)
			cancel()
			wg.Done()
		}()
	}
	wg.Wait()
	// 释放资源
	log.Println("开始释放资源")
	app.close()
}

// close 在这里释放掉一些可能的资源
func (app *App) close() {
	time.Sleep(time.Second)
	log.Println("应用关闭")
}

type Server struct {
	srv  *http.Server
	name string
	mux  *serverMux
}

// serverMux 既可以看做是装饰器模式，可以看做是委托模式
type serverMux struct {
	reject bool
	*http.ServeMux
}

func (s *serverMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.reject {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("服务已关闭"))
		return
	}
	s.ServeMux.ServeHTTP(w, r)
}

func NewServer(name string, addr string) *Server {
	mux := &serverMux{ServeMux: http.NewServeMux()}
	return &Server{
		name: name,
		mux:  mux,
		srv: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
	}
}

func (s *Server) Handle(pattern string, handler http.Handler) {
	s.mux.Handle(pattern, handler)
}

func (s *Server) Start() error {
	return s.srv.ListenAndServe()
}

func (s *Server) rejectReq() {
	s.mux.reject = true
}

func (s *Server) stop1(ctx context.Context) error {
	log.Printf("服务器%s关闭中", s.name)
	return s.srv.Shutdown(ctx)
}

func (s *Server) stop() error {
	log.Printf("服务器%s关闭中", s.name)
	return s.srv.Shutdown(context.Background())
}

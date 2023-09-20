// Package server 服务端，包括网络通信 名字服务 监控统计 链路跟踪等各个组件基础接口

package server

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Server 是一個tinyRpc的server，包括很多个service
type Server struct {
	MaxCloseWaitTime time.Duration      //服務最大等待時間
	services         map[string]Service //服务的映射，k是服务名称，v是服务实体
	mux              sync.Mutex         //互斥锁（Mutex）结构，避免多个协程同时修改或访问同一个资源导致的竞态条件
	failedServices   sync.Map           // 失败的服务的map
	closeCh          chan struct{}      //空结构体用于在通道上进行信号传递或同步操作
	closeOnce        sync.Once          //用于一些只需要执行一次的操作，比如初始化一个全局变量、注册信号处理程序等。
}

// AddService 添加Service到map中
// 参数 serviceName 是用于命名服务的名称，并且在配置文件（通常为tinyRpc_go.yaml）
// 中进行配置。当 tinyRpc.NewServer() 被调用时，它将遍历配置文件中的服务配置，并调用
// AddService 方法将服务实现添加到服务器的 map[string]Service（serviceName 作为键）中。
func (s *Server) AddService(serviceName string, service Service) {
	if s.services == nil {
		s.services = make(map[string]Service)
	}
	s.services[serviceName] = service
}

func (s *Server) Service(serviceName string) Service {
	if s.services == nil {
		return nil
	}
	return s.services[serviceName]
}

// Register 会注册map里所有的服务
func (s *Server) Register(serviceDesc interface{}, serviceImpl interface{}) error {
	desc, ok := serviceDesc.(*ServiceDesc)
	if !ok {
		return errors.New("service desc type invalid")
	}

	for _, srv := range s.services {
		if err := srv.Register(desc, serviceImpl); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) Close(ch chan struct{}) error {
	if s.closeCh != nil {
		close(s.closeCh)
	}

	s.tryClose()

	if ch != nil {
		ch <- struct{}{}
	}
	return nil
}

func (s *Server) tryClose() {
	// 定义只执行一次的闭包函数
	// 函数字面量（Function Literal）或匿名函数
	// 匿名函数常用于需要定义临时函数的场景，比如作为函数参数进行传递、在协程中进行并发执行等。
	fn := func() {

		// 在关闭服务之前执行关闭钩子函数(不实现)
		//s.mux.Lock()
		//for _, f := range s.onShutdownHooks {
		//	f()
		//}
		//s.mux.Unlock()

		// 关闭所有服务
		closeWaitTime := s.MaxCloseWaitTime
		if closeWaitTime < MaxCloseWaitTime {
			closeWaitTime = MaxCloseWaitTime
		}

		// 创建一个带有超时的上下文，用于等待服务关闭
		// 将 context.Background() 作为父上下文，然后使用 WithTimeout 派生一个具有超时功能的子上下文。
		// 这样，在执行一些可能耗时的操作时，可以在超时时间到达时自动取消并终止操作。
		ctx, cancel := context.WithTimeout(context.Background(), closeWaitTime)
		defer cancel()

		// 使用 WaitGroup 跟踪所有服务的关闭操作
		var wg sync.WaitGroup
		for name, service := range s.services {
			// 跳过已失败的服务
			if _, ok := s.failedServices.Load(name); ok {
				continue
			}

			wg.Add(1)
			go func(srv Service) {
				defer wg.Done()

				// 创建一个用于通知服务关闭的通道
				c := make(chan struct{}, 1)
				// 对该Service执行关闭
				go func() {
					err := srv.Close(c)
					if err != nil {
						//关闭失败
					}
				}()

				// 等待服务关闭或上下文超时
				select {
				case <-c:
					// 服务成功关闭
				case <-ctx.Done():
					// 服务关闭超时
				}
			}(service)
		}
		wg.Wait()
	}

	// 只执行一次闭包函数
	s.closeOnce.Do(fn)
}

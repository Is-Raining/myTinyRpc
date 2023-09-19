// Package server 服务端，包括网络通信 名字服务 监控统计 链路跟踪等各个组件基础接口
package server

import (
	"errors"
	"time"
)

// Server 是一個tinyRpc的server
type Server struct {
	MaxCloseWaitTime time.Duration //服務最大等待時間

	services map[string]Service //服务的映射，k是服务名称，v是服务实体
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

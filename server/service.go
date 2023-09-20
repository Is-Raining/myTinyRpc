package server

// Service is the interface that provides services.
type Service interface {
	// Register 注册服务.
	Register(serviceDesc interface{}, serviceImpl interface{}) error
	// Serve 开始服务.
	Serve() error
	// Close 关闭服务.
	Close(chan struct{}) error
}

type ServiceDesc struct {
	ServiceName string
}

package iface

import "github.com/ajdwfnhaps/easy-logrus/logger"

//IServer 定义服务器接口
type IServer interface {
	//启动服务器方法
	Start()
	//停止服务器方法
	Stop()
	//开启业务服务方法
	Serve()
	//路由功能：给当前服务注册一个路由业务方法，供客户端链接处理使用
	AddRouter(cmd string, router IRouter)
	//得到链接管理
	GetConnMgr() IConnManager
	//设置该Server的连接创建时Hook函数
	SetOnConnStart(func(IConnection))
	//设置该Server的连接断开时的Hook函数
	SetOnConnStop(func(IConnection))
	//调用连接OnConnStart Hook函数
	CallOnConnStart(conn IConnection)
	//调用连接OnConnStop Hook函数
	CallOnConnStop(conn IConnection)
	//设置使用日志框架
	SetLogger(logger logger.ILogger)
	//获取日志框架
	GetLogger() logger.ILogger
	//广播
	Broadcast(data []byte)

	//设置该Server成功启动后Hook函数
	SetOnServerStarted(func(s IServer))
	//调用OnServerStarted Hook函数
	CallOnServerStarted(s IServer)
}

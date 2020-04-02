package iface

import (
	"github.com/ajdwfnhaps/easy-tcp-server/dto"
)

//IRequest 接口：	实际上是把客户端请求的链接信息 和 请求的数据 包装到了 Request里
type IRequest interface {
	GetConnection() IConnection //获取请求连接信息
	GetMsg() IMessage           //接收到的消息
	GetRet() dto.Result         //接收到的消息反序列化的结果
	GetRouterCmd() string       //获取路由路径
}

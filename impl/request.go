package impl

import (
	"github.com/ajdwfnhaps/easy-tcp-server/dto"
	"github.com/ajdwfnhaps/easy-tcp-server/iface"
)

type Request struct {
	conn iface.IConnection //已经和客户端建立好的 链接
	msg  iface.IMessage    //客户端请求的数据
	ret  dto.Result        //反序列化的结果
}

//获取请求连接信息
func (r *Request) GetConnection() iface.IConnection {
	return r.conn
}

//获取请求的消息的ID
func (r *Request) GetMsg() iface.IMessage {
	return r.msg
}

func (r *Request) GetRet() dto.Result {
	return r.ret
}

func (r *Request) GetRouterCmd() string {
	return r.ret.Cmd
}

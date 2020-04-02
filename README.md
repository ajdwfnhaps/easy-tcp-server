# easy-tcp-server
easy-tcp-server


### 服务端启动代码：
```go

var (
	configPath string
)

func main() {

	configPath = "config.toml"
	//初始化tcp-server
	go initTCPServer(configPath)
	//等待退出信号
	handleSignal()
}

func initTCPServer(configPath string) {
	tcpUtils.SetConfigPath(configPath)
	//创建一个server句柄
	s := impl.NewServer()

	//设置使用的日志框架
	s.SetLogger(logger.CreateLogger())
	//注册路由
	registerTCPRouters(s)
	//开启服务
	s.Serve()
}

func registerTCPRouters(s iface.IServer) {
	//心跳检测
	s.AddRouter("request_heartbeat", &router.HeartbeatHandler{})
}

func handleSignal() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case <-c:
		//断开mqtt连接
		//mqtt.GetClient().Disconnect()

		if tcpUtils.GlobalObject.TcpServer != nil {
			tcpUtils.GlobalObject.TcpServer.Stop()
		}
	}
}

```

### 心跳路由处理器

```go

package router

import (
	"encoding/json"

	"github.com/ajdwfnhaps/easy-tcp-server/dto"
	"github.com/ajdwfnhaps/easy-tcp-server/iface"
	"github.com/ajdwfnhaps/easy-tcp-server/impl"
	"github.com/ajdwfnhaps/easy-logrus/logger"
)

//HeartbeatHandler 心跳处理器
type HeartbeatHandler struct {
	impl.BaseRouter
}

//Handle 处理方法
func (h *HeartbeatHandler) Handle(req iface.IRequest) {
	log := logger.CreateLogger()
	reqRet := req.GetRet()

	ret := dto.Result{
		Seqno: reqRet.Seqno,
		Cmd:   reqRet.Cmd,
	}

	data, err := json.Marshal(ret)
	if err != nil {
		log.Errorf("HeartbeatHandler 响应结果序列化出错，%s", err.Error())
	}

	//向客户端发送消息
	if err = req.GetConnection().SendMsg(data); err != nil {
		log.Errorf("HeartbeatHandler SendMsg出错，%s", err.Error())
	}

	log.Infof("处理来自[%s]的请求，cmd:%s，TCP响应成功！", req.GetConnection().RemoteAddr(), reqRet.Cmd)
}

```

### 配置文件解释 config.toml
```
[tcp]
# 服务端ip
host="0.0.0.0"
# 服务端监听端口
port=8091
# 当前服务器主机允许的最大链接个数
max_conn=100
# 业务工作Worker池的数量
worker_pool_size=5
# 业务工作Worker对应负责的任务队列最大任务存储数量
max_worker_task_len=128 
# SendBuffMsg发送消息的缓冲最大长度
max_msg_chan_len=128
```

# 客户端测试
请参照示例代码:[server_test.go](impl/server_test.go)

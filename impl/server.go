package impl

import (
	"fmt"
	"net"

	"github.com/ajdwfnhaps/easy-tcp-server/iface"
	"github.com/ajdwfnhaps/easy-tcp-server/utils"
	"github.com/ajdwfnhaps/easy-logrus/logger"
)

//Server 接口实现，定义一个Server服务类
type Server struct {
	//服务器的名称
	Name string
	//tcp4 or other
	IPVersion string
	//服务绑定的IP地址
	IP string
	//服务绑定的端口
	Port int
	//当前Server的消息管理模块，用来绑定MsgId和对应的处理方法
	msgHandler iface.IMsgHandle
	//当前Server的链接管理器
	ConnMgr iface.IConnManager
	//该Server的连接创建时Hook函数
	OnConnStart func(conn iface.IConnection)
	//该Server的连接断开时的Hook函数
	OnConnStop func(conn iface.IConnection)
	//该Server成功启动后的Hook函数
	OnServerStarted func(s iface.IServer)
	//日志
	Logger logger.ILogger
	//广播消息channel
	bcChan chan []byte
}

// NewServer 创建一个服务器句柄
func NewServer() iface.IServer {

	s := &Server{
		Name:       utils.GlobalObject.Name,
		IPVersion:  "tcp4",
		IP:         utils.GlobalObject.Host,
		Port:       utils.GlobalObject.TcpPort,
		msgHandler: NewMsgHandle(),
		ConnMgr:    NewConnManager(),
		bcChan:     make(chan []byte),
	}
	return s
}

//============== 实现 iface.IServer 里的全部接口方法 ========

//Start 开启网络服务
func (s *Server) Start() {
	s.Logger.Printf("IOT Tcp Server : %s , listen at IP: %s, Port %d is starting\n", s.Name, s.IP, s.Port)

	//开启一个go去做服务端Linster业务
	go func() {
		//0 启动worker工作池机制
		s.msgHandler.StartWorkerPool()

		//1 获取一个TCP的Addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			s.Logger.Errorf("resolve tcp addr err: ", err)
			return
		}

		//2 监听服务器地址
		listenner, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			s.Logger.Errorf("listen", s.IPVersion, "err", err)
			return
		}

		//已经监听成功
		s.Logger.Info("start iot tcp server  ", s.Name, " succ, now listenning...")

		//开启广播处理器
		go s.handleBroadcast()
		//TODO server.go 应该有一个自动生成ID的方法
		var cid uint32
		cid = 0

		//3 启动server网络连接业务
		for {
			//3.1 阻塞等待客户端建立连接请求
			conn, err := listenner.AcceptTCP()
			if err != nil {
				s.Logger.Errorf("Accept err ", err)
				continue
			}
			s.Logger.Info("新的tcp客户端连接已创建, conn remote addr = ", conn.RemoteAddr().String())

			//3.2 设置服务器最大连接控制,如果超过最大连接，那么则关闭此新的连接
			if s.ConnMgr.Len() >= utils.GlobalObject.MaxConn {
				s.Logger.Warnf("tcp连接数已超出配置上限：%d,将会关闭连接", utils.GlobalObject.MaxConn)
				conn.Close()
				continue
			}

			//3.3 处理该新连接请求的 业务 方法， 此时应该有 handler 和 conn是绑定的
			dealConn := NewConntion(s, conn, cid, s.msgHandler)

			// var cidLock sync.RWMutex
			// cidLock.Lock()
			cid++
			// cidLock.Unlock()

			//3.4 启动当前链接的处理业务
			go dealConn.Start()
		}
	}()
}

//Stop 停止服务
func (s *Server) Stop() {
	//将其他需要清理的连接信息或者其他信息 也要一并停止或者清理
	s.ConnMgr.ClearConn()
	close(s.bcChan)
	s.Logger.Info("iot tcp server has been stoped")
}

//Serve 运行服务
func (s *Server) Serve() {

	if s.Logger == nil {
		s.Logger = &logger.Logger{}
	}

	utils.GlobalObject.Logger = s.Logger

	s.Start()
	utils.GlobalObject.TcpServer = s
	s.CallOnServerStarted(s)
	//TODO Server.Serve() 是否在启动服务的时候 还要处理其他的事情呢 可以在这里添加

	//阻塞,否则主Go退出， listenner的go将会退出
	select {}
}

//AddRouter 路由功能：给当前服务注册一个路由业务方法，供客户端链接处理使用
func (s *Server) AddRouter(cmd string, router iface.IRouter) {
	s.msgHandler.AddRouter(cmd, router)
}

//GetConnMgr 得到链接管理
func (s *Server) GetConnMgr() iface.IConnManager {
	return s.ConnMgr
}

//SetOnConnStart 设置该Server的连接创建时Hook函数
func (s *Server) SetOnConnStart(hookFunc func(iface.IConnection)) {
	s.OnConnStart = hookFunc
}

//SetOnConnStop 设置该Server的连接断开时的Hook函数
func (s *Server) SetOnConnStop(hookFunc func(iface.IConnection)) {
	s.OnConnStop = hookFunc
}

//CallOnConnStart 调用连接OnConnStart Hook函数
func (s *Server) CallOnConnStart(conn iface.IConnection) {
	if s.OnConnStart != nil {
		s.Logger.Info("---> CallOnConnStart....")
		s.OnConnStart(conn)
	}
}

//CallOnConnStop 调用连接OnConnStop Hook函数
func (s *Server) CallOnConnStop(conn iface.IConnection) {
	if s.OnConnStop != nil {
		s.Logger.Info("---> CallOnConnStop....")
		s.OnConnStop(conn)
	}
}

//SetOnServerStarted 设置该Server成功启动后Hook函数
func (s *Server) SetOnServerStarted(hookFunc func(s iface.IServer)) {
	s.OnServerStarted = hookFunc
}

//CallOnServerStarted 调用连接OnServerStarted Hook函数
func (s *Server) CallOnServerStarted(svr iface.IServer) {
	if s.OnServerStarted != nil {
		s.Logger.Info("CallOnServerStarted....")
		s.OnServerStarted(svr)
	}
}

// SetLogger 设置日志框架
func (s *Server) SetLogger(logger logger.ILogger) {
	utils.GlobalObject.Logger = logger
	s.Logger = logger
}

//GetLogger 获取日志框架
func (s *Server) GetLogger() logger.ILogger {
	return s.Logger
}

//Broadcast 广播
func (s *Server) Broadcast(data []byte) {
	s.bcChan <- data
}

//handleBroadcast 广播处理器
func (s *Server) handleBroadcast() {
	s.Logger.Debug("广播处理器已启动...")
	for {
		select {
		case data := <-s.bcChan:
			if len(data) == 0 {
				break
			}

			connCount := s.ConnMgr.Len()
			if connCount < 1 {
				s.Logger.Info("当前客户端连接数为0，退出广播")
				break
			}

			var i int
			for _, conn := range s.ConnMgr.GetAll() {
				err := conn.SendBuffMsg(data)
				if err != nil {
					s.Logger.Errorf("广播到客户端[%s]报错:%s", conn.RemoteAddr(), err.Error())
				} else {
					i++
				}
			}
			s.Logger.Infof("广播成功数量：%d, 总客户端连接数：%d", i, connCount)
		}
	}
}

// func init() {
// 	fmt.Printf("[Iot Tcp Server Init] Version: %s, MaxConn: %d, MaxPacketSize: %d\n",
// 		utils.GlobalObject.Version,
// 		utils.GlobalObject.MaxConn,
// 		utils.GlobalObject.MaxPacketSize)
// }

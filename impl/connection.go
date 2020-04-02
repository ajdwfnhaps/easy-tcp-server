package impl

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"net"
	"sync"

	"github.com/ajdwfnhaps/easy-logrus/logger"
	"github.com/ajdwfnhaps/easy-tcp-server/dto"
	"github.com/ajdwfnhaps/easy-tcp-server/iface"
	"github.com/ajdwfnhaps/easy-tcp-server/utils"
)

//Connection tcp连接
type Connection struct {
	//当前Conn属于哪个Server
	TcpServer iface.IServer
	//当前连接的socket TCP套接字
	Conn *net.TCPConn
	//当前连接的ID 也可以称作为SessionID，ID全局唯一
	ConnID uint32
	//当前连接的关闭状态
	isClosed bool
	//消息管理MsgId和对应处理方法的消息管理模块
	MsgHandler iface.IMsgHandle
	//告知该链接已经退出/停止的channel
	ExitBuffChan chan bool
	//无缓冲管道，用于读、写两个goroutine之间的消息通信
	msgChan chan []byte
	//有关冲管道，用于读、写两个goroutine之间的消息通信
	msgBuffChan chan []byte

	//链接属性
	property map[string]interface{}
	//保护链接属性修改的锁
	propertyLock sync.RWMutex
	//日志
	logger logger.ILogger
}

//NewConntion 创建连接的方法
func NewConntion(server iface.IServer, conn *net.TCPConn, connID uint32, msgHandler iface.IMsgHandle) *Connection {
	//初始化Conn属性
	c := &Connection{
		TcpServer:    server,
		Conn:         conn,
		ConnID:       connID,
		isClosed:     false,
		MsgHandler:   msgHandler,
		ExitBuffChan: make(chan bool, 1),
		msgChan:      make(chan []byte),
		msgBuffChan:  make(chan []byte, utils.GlobalObject.MaxMsgChanLen),
		property:     make(map[string]interface{}),
	}

	c.logger = c.TcpServer.GetLogger()
	//将新创建的Conn添加到链接管理中
	c.TcpServer.GetConnMgr().Add(c)
	return c
}

//StartWriter 写消息Goroutine， 用户将数据发送给客户端
func (c *Connection) StartWriter() {
	c.logger.Info("[Tcp Writer Goroutine is running]")
	defer c.logger.Info(c.RemoteAddr().String(), "[Tcp conn Writer exit!]")

	for {
		select {
		case data := <-c.msgChan:
			//有数据要写给客户端
			if _, err := c.Conn.Write(data); err != nil {
				c.logger.Error("Send Data error:, ", err, " Conn Writer exit")
				return
			}
			//fmt.Printf("Send data succ! data = %+v\n", data)
		case data, ok := <-c.msgBuffChan:
			if ok {
				//有数据要写给客户端
				if _, err := c.Conn.Write(data); err != nil {
					c.logger.Error("Send Buff Data error:, ", err, " Conn Writer exit")
					return
				}
			} else {
				c.logger.Info("msgBuffChan is Closed")
				break
			}
		case <-c.ExitBuffChan:
			return
		}
	}
}

//StartReader 读消息Goroutine，用于从客户端中读取数据
func (c *Connection) StartReader() {
	c.logger.Info("[Tcp Reader Goroutine is running]")
	defer c.logger.Info(c.RemoteAddr().String(), "[Tcp conn Reader exit!]")
	defer c.Stop()

	for {

		reader := bufio.NewReader(c.Conn)
		bodySizeBytes, err := reader.Peek(4)
		if err != nil {
			c.logger.Error("reader.Peek error.", err)
			break
		}

		c.logger.Info("读取到客户端[", c.RemoteAddr(), "]发过来的新消息...")

		bodySizeBuff := bytes.NewBuffer(bodySizeBytes)
		var bodySize uint32
		err = binary.Read(bodySizeBuff, binary.LittleEndian, &bodySize)
		if err != nil {
			if err == io.EOF {
				c.logger.Info("binary.Read error.跳出循环", err)
				break // 跳出循环
			} else {
				c.logger.Error("binary.Read error.", err)
				continue
			}
		}

		dataPacker := NewDataPack()
		pkgLength := bodySize + dataPacker.GetHeadLen() //包头长度24

		// Buffered返回缓冲中现有的可读取的字节数。
		if uint32(reader.Buffered()) < pkgLength {
			c.logger.Warn("binary.Read 读取完整包失败,丢弃包.\r\n读取包的字节数：", reader.Buffered(), ",按自定义协议计算的包长度：", pkgLength)
			continue
		}

		//c.Conn.SetReadDeadline(time.Now().Add(time.Duration(10) * time.Second))

		data := make([]byte, pkgLength)
		if _, err := io.ReadFull(reader, data); err != nil {
			c.logger.Error("read msg data error ", err)
			break
		}

		msg, err := dataPacker.Unpack(data)
		if err != nil {
			c.logger.Error("unpack error ", err)
			break
		}

		var jsonObj dto.Result
		json.Unmarshal(msg.GetBody(), &jsonObj)

		//得到当前客户端请求的Request数据
		req := Request{
			conn: c,
			msg:  msg,
			ret:  jsonObj,
		}

		if utils.GlobalObject.WorkerPoolSize > 0 {
			//已经启动工作池机制，将消息交给Worker处理
			c.MsgHandler.SendMsgToTaskQueue(&req)
		} else {
			//从绑定好的消息和对应的处理方法中执行对应的Handle方法
			go c.MsgHandler.DoMsgHandler(&req)
		}
	}
}

//Start 启动连接，让当前连接开始工作
func (c *Connection) Start() {
	//1 开启用户从客户端读取数据流程的Goroutine
	go c.StartReader()
	//2 开启用于写回客户端数据流程的Goroutine
	go c.StartWriter()
	//按照用户传递进来的创建连接时需要处理的业务，执行钩子方法
	c.TcpServer.CallOnConnStart(c)
}

//Stop 停止连接，结束当前连接状态M
func (c *Connection) Stop() {
	c.logger.Info("tcp客户端断开连接...ConnID = ", c.ConnID, ", ClientAddr:", c.RemoteAddr())
	//如果当前链接已经关闭
	if c.isClosed == true {
		return
	}
	c.isClosed = true

	//如果用户注册了该链接的关闭回调业务，那么在此刻应该显示调用
	c.TcpServer.CallOnConnStop(c)

	// 关闭socket链接
	c.Conn.Close()
	//关闭Writer
	c.ExitBuffChan <- true

	//将链接从连接管理器中删除
	c.TcpServer.GetConnMgr().Remove(c)

	//关闭该链接全部管道
	close(c.ExitBuffChan)
	close(c.msgBuffChan)
}

//GetTCPConnection 从当前连接获取原始的socket TCPConn
func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

//GetConnID 获取当前连接ID
func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

//RemoteAddr 获取远程客户端地址信息
func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

//SendMsg 直接将Message数据发送数据给远程的TCP客户端
func (c *Connection) SendMsg(data []byte) error {
	if c.isClosed == true {
		return errors.New("Connection closed when send msg")
	}
	//将data封包，并且发送
	msg := NewMsgPackage(data)
	dataPacker := NewDataPack()
	data, err := dataPacker.Pack(msg)
	if err != nil {
		return err
	}
	//写回客户端
	c.msgChan <- data

	return nil
}

//SendBuffMsg SendBuffMsg
func (c *Connection) SendBuffMsg(data []byte) error {
	if c.isClosed == true {
		return errors.New("Connection closed when send buff msg")
	}
	//将data封包，并且发送
	msg := NewMsgPackage(data)
	dataPacker := NewDataPack()
	data, err := dataPacker.Pack(msg)
	if err != nil {
		return err
	}

	//写回客户端
	c.msgBuffChan <- data

	return nil
}

//SetProperty 设置链接属性
func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.property[key] = value
}

//GetProperty 获取链接属性
func (c *Connection) GetProperty(key string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	if value, ok := c.property[key]; ok {
		return value, nil
	}
	return nil, errors.New("no property found")
}

//RemoveProperty 移除链接属性
func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.property, key)
}

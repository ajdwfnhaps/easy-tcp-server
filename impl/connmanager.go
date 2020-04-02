package impl

import (
	"errors"
	"sync"

	"github.com/ajdwfnhaps/easy-tcp-server/iface"
	"github.com/ajdwfnhaps/easy-tcp-server/utils"
)

/*
	连接管理模块
*/
type ConnManager struct {
	connections map[uint32]iface.IConnection //管理的连接信息
	connLock    sync.RWMutex                 //读写连接的读写锁
}

/*
	创建一个链接管理
*/
func NewConnManager() *ConnManager {
	return &ConnManager{
		connections: make(map[uint32]iface.IConnection),
	}
}

//添加链接
func (connMgr *ConnManager) Add(conn iface.IConnection) {
	//保护共享资源Map 加写锁
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()

	//将conn连接添加到ConnMananger中
	connMgr.connections[conn.GetConnID()] = conn

	utils.GlobalObject.Logger.Info("connection 添加到tcp连接管理池成功: conn count = ", connMgr.Len())
}

//删除连接
func (connMgr *ConnManager) Remove(conn iface.IConnection) {
	//保护共享资源Map 加写锁
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()

	//删除连接信息
	delete(connMgr.connections, conn.GetConnID())

	utils.GlobalObject.Logger.Info("connection 从tcp连接管理池移除成功 ConnID=", conn.GetConnID(), ": conn count = ", connMgr.Len())
}

//利用ConnID获取链接
func (connMgr *ConnManager) Get(connID uint32) (iface.IConnection, error) {
	//保护共享资源Map 加读锁
	connMgr.connLock.RLock()
	defer connMgr.connLock.RUnlock()

	if conn, ok := connMgr.connections[connID]; ok {
		return conn, nil
	} else {
		return nil, errors.New("connection not found")
	}
}

//获取当前连接
func (connMgr *ConnManager) Len() int {
	return len(connMgr.connections)
}

//清除并停止所有连接
func (connMgr *ConnManager) ClearConn() {
	// 此处加写锁会导致死锁
	// connMgr.connLock.Lock()
	// defer connMgr.connLock.Unlock()

	//停止并删除全部的连接信息
	for connID, conn := range connMgr.connections {
		//停止
		conn.Stop()
		//删除
		delete(connMgr.connections, connID)
	}

	utils.GlobalObject.Logger.Info("Clear All Connections successfully: conn count = ", connMgr.Len())
}

//GetAll 获取当前所有连接
func (connMgr *ConnManager) GetAll() map[uint32]iface.IConnection {
	return connMgr.connections
}

//GetConnByProp 根据属性获取所有连接
func (connMgr *ConnManager) GetConnByProp(key string, val interface{}) []iface.IConnection {
	var conns []iface.IConnection
	for _, conn := range connMgr.GetAll() {
		if v, err := conn.GetProperty(key); err == nil {
			if val == nil && v != nil {
				conns = append(conns, conn)
			} else if val != nil && v != nil && val == v {
				conns = append(conns, conn)
			}
		}
	}

	return conns
}

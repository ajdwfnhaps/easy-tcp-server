package utils

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/ajdwfnhaps/easy-tcp-server/iface"
	"github.com/ajdwfnhaps/easy-logrus/logger"
)

//GlobalObj 全局配置
type GlobalObj struct {
	/*
		Server
	*/
	TcpServer iface.IServer //当前Zinx的全局Server对象
	Host      string        `toml:"host"`            //当前服务器主机IP
	TcpPort   int           `toml:"port"`            //当前服务器主机监听端口号
	Name      string        `toml:"tcp_server_name"` //当前服务器名称

	/*
		Zinx
	*/
	Version          string `toml:"version"`             //当前Zinx版本号
	MaxPacketSize    uint32 `toml:"max_packet_size"`     //都需数据包的最大值
	MaxConn          int    `toml:"max_conn"`            //当前服务器主机允许的最大链接个数
	WorkerPoolSize   uint32 `toml:"worker_pool_size"`    //业务工作Worker池的数量
	MaxWorkerTaskLen uint32 `toml:"max_worker_task_len"` //业务工作Worker对应负责的任务队列最大任务存储数量
	MaxMsgChanLen    uint32 `toml:"max_msg_chan_len"`    //SendBuffMsg发送消息的缓冲最大长度

	/*
		config file path
	*/
	ConfFilePath string

	//日志
	Logger logger.ILogger
}

//TCPConfig 统一配置类
type TCPConfig struct {
	Opt *GlobalObj `toml:"tcp"`
}

//GlobalObject 定义一个全局的对象
var GlobalObject *GlobalObj

//PathExists 判断一个文件是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

//Reload 读取用户的配置文件
func Reload(c *TCPConfig) {
	g := c.Opt
	if confFileExists, _ := PathExists(g.ConfFilePath); confFileExists != true {
		//fmt.Println("Config File ", g.ConfFilePath, " is not exist!!")
		return
	}

	_, err := toml.DecodeFile(g.ConfFilePath, c)
	if err != nil {
		panic(err)
	}

}

/*
	提供init方法，默认加载
*/
func init() {
	//初始化GlobalObject变量，设置一些默认值
	GlobalObject = &GlobalObj{
		Name:             "AamIotTcpServer",
		Version:          "V1.0.0",
		TcpPort:          8090,
		Host:             "0.0.0.0",
		MaxConn:          12000,
		MaxPacketSize:    4096,
		ConfFilePath:     "../configs/tcp.toml",
		WorkerPoolSize:   10,
		MaxWorkerTaskLen: 1024,
		MaxMsgChanLen:    1024,
	}

	//从配置文件中加载一些用户配置的参数
	Reload(&TCPConfig{Opt: GlobalObject})
}

//SetConfigPath 设置配置文件路径
func SetConfigPath(fpath string) {
	GlobalObject.ConfFilePath = fpath
	Reload(&TCPConfig{Opt: GlobalObject})
}

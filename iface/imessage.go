package iface

//IMessage 将请求的一个消息封装到message中，定义抽象层接口
type IMessage interface {
	GetBodySize() int32
	GetVersion() [4]byte
	GetCompress() byte
	GetClientType() byte
	GetBsdCode() [14]byte
	GetBody() []byte
}

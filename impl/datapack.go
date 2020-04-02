package impl

import (
	"bytes"
	"encoding/binary"

	"github.com/ajdwfnhaps/easy-tcp-server/iface"
)

//DataPack 封包拆包类实例，暂时不需要成员
type DataPack struct{}

//NewDataPack 封包拆包实例初始化方法
func NewDataPack() *DataPack {
	return &DataPack{}
}

//GetHeadLen 获取包头长度方法
func (dp *DataPack) GetHeadLen() uint32 {
	return 24
}

//Pack 封包方法(压缩数据)
func (dp *DataPack) Pack(msg iface.IMessage) ([]byte, error) {
	//创建一个存放bytes字节的缓冲
	dataBuff := bytes.NewBuffer([]byte{})

	var err error
	err = binary.Write(dataBuff, binary.LittleEndian, msg.GetBodySize())
	err = binary.Write(dataBuff, binary.BigEndian, msg.GetVersion())
	err = binary.Write(dataBuff, binary.BigEndian, msg.GetCompress())
	err = binary.Write(dataBuff, binary.BigEndian, msg.GetClientType())
	err = binary.Write(dataBuff, binary.BigEndian, msg.GetBsdCode())
	err = binary.Write(dataBuff, binary.BigEndian, msg.GetBody())

	return dataBuff.Bytes(), err
}

//Unpack 拆包方法(解压数据)
func (dp *DataPack) Unpack(binaryData []byte) (iface.IMessage, error) {
	//创建一个从输入二进制数据的ioReader
	dataBuff := bytes.NewReader(binaryData)

	//只解压head的信息，得到dataLen和msgID
	msg := &Message{}

	var err error
	err = binary.Read(dataBuff, binary.LittleEndian, &msg.BodySize)
	err = binary.Read(dataBuff, binary.BigEndian, &msg.Version)
	err = binary.Read(dataBuff, binary.BigEndian, &msg.Compress)
	err = binary.Read(dataBuff, binary.BigEndian, &msg.ClientType)

	err = binary.Read(dataBuff, binary.BigEndian, &msg.BsdCode)

	msg.Body = make([]byte, msg.BodySize)
	err = binary.Read(dataBuff, binary.BigEndian, &msg.Body)

	return msg, err
}

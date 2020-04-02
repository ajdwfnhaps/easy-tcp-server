package impl

//Message 消息
type Message struct {
	BodySize   int32
	Version    [4]byte
	Compress   byte
	ClientType byte
	BsdCode    [14]byte
	Body       []byte
}

//NewMsgPackage 创建一个Message消息包
func NewMsgPackage(bodyData []byte) *Message {
	var msg = &Message{
		Version:    [4]byte{'2', '0', '0', '1'},
		Compress:   0,
		ClientType: 0,
	}
	msg.BsdCode[0] = 'i'
	msg.BsdCode[1] = 'o'
	msg.BsdCode[2] = 't'
	msg.BodySize = int32(len(bodyData))
	msg.Body = bodyData
	return msg
}

//GetBodySize Get BodySize
func (p *Message) GetBodySize() int32 {
	return p.BodySize
}

//GetVersion Get Version
func (p *Message) GetVersion() [4]byte {
	return p.Version
}

//GetCompress Get Compress
func (p *Message) GetCompress() byte {
	return p.Compress
}

//GetClientType Get ClientType
func (p *Message) GetClientType() byte {
	return p.ClientType
}

//GetBsdCode Get BsdCode
func (p *Message) GetBsdCode() [14]byte {
	return p.BsdCode
}

//GetBody Get Body
func (p *Message) GetBody() []byte {
	return p.Body
}

package impl

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/ajdwfnhaps/easy-tcp-server/dto"
	"github.com/ajdwfnhaps/easy-tcp-server/iface"
)

func TestTagHandler(t *testing.T) {
	send(`{
		"cmd": "report_tag",
		"seqno":"111",
		"data":"tag-log-client"
}`)
}

func TestHeartbeatHandler(t *testing.T) {
	send(`{
		"cmd": "request_heartbeat",
		"seqno":"你好吗",
		"data":{
			"mid":"123",
			"product_type":"169",
			"pid":"",
			"mac_status":1,
			"product_key":""
	}
}`)
}

func TestGatewayHandler(t *testing.T) {
	send(`{
		"cmd": "request_gateway",
		"seqno":""
}`)
}

func TestMidHandler(t *testing.T) {
	send(`{
		"cmd": "request_mid",
		"seqno":"",
		"data":null
}`)
}

func TestTerminateHandler(t *testing.T) {
	send(`{
		"cmd": "request_terminate",
		"seqno":"777"
}`)
}

func TestBarcodeHandle(t *testing.T) {
	send(`{
		"cmd": "request_barcode",
		"seqno":"1001"
}`)
}

func TestTokenHandler(t *testing.T) {
	send(`{
		"cmd": "request_token",
		"seqno":"001"
}`)
}

func send(msg string) {

	fmt.Println("Client Test  ... start")
	//3秒之后发起测试请求，给服务端开启服务的机会
	// time.Sleep(3 * time.Second)

	conn, err := net.Dial("tcp", "127.0.0.1:8091") //10.0.3.173
	if err != nil {
		fmt.Println("client start err, exit!")
		return
	}

	for {
		//request_mid request_heartbeat
		msg := NewMsgPackage([]byte(msg))

		dataPacker := NewDataPack()
		data, err := dataPacker.Pack(msg)

		if err != nil {
			fmt.Println("Pack error ", err)
			return
		}

		_, err = conn.Write(data)
		if err != nil {
			fmt.Println("write error err ", err)
			return
		}

		reader := bufio.NewReader(conn)
		bodySizeBytes, _ := reader.Peek(4)

		bodySizeBuff := bytes.NewBuffer(bodySizeBytes)
		var bodySize uint32
		err = binary.Read(bodySizeBuff, binary.LittleEndian, &bodySize)
		if err != nil {
			if err == io.EOF {
				fmt.Println("binary.Read error.跳出循环", err)
				break // 跳出循环
			} else {
				fmt.Println("binary.Read error.", err)
				continue
			}
		}
		pkgLength := bodySize + dataPacker.GetHeadLen() //包头长度24

		// Buffered返回缓冲中现有的可读取的字节数。
		if uint32(reader.Buffered()) < pkgLength {
			fmt.Println("binary.Read 读取完整包失败", reader.Buffered(), pkgLength)
			continue
		}

		conn.SetReadDeadline(time.Now().Add(time.Duration(10) * time.Second))

		recdata := make([]byte, pkgLength)
		if _, err := io.ReadFull(reader, recdata); err != nil {
			fmt.Println("read msg data error ", err)
			break
		}

		//拆包
		recMsg, err := dataPacker.Unpack(recdata)
		if err != nil {
			fmt.Println("unpack error ", err)
			break
		}

		body := recMsg.GetBody()
		fmt.Printf("server call back :\r\n%s \r\nbody-length = %d\n", string(body), len(body))
		//fmt.Println(msg)

		time.Sleep(3000 * time.Millisecond)
	}
}

func BenchmarkClientRecBroadcast(b *testing.B) {

	b.ResetTimer()
	for i := 1; i <= b.N; i++ {
		go clientRec(fmt.Sprintf("clien-%d", i))
	}

	b.StopTimer()
	//select {}
}

func TestClientRecBroadcast(t *testing.T) {
	//for i := 1; i <= 3; i++ {
	go clientRec(fmt.Sprintf("clien-%d", 1))
	//}
	select {}
}

func clientRec(name string) {
	conn, err := net.Dial("tcp", "127.0.0.1:8091") //10.0.3.173
	if err != nil {
		fmt.Println("client start err, exit!" + err.Error())
		return
	}

	fmt.Printf("client %s started\r\n", name)

	for {
		reader := bufio.NewReader(conn)
		bodySizeBytes, _ := reader.Peek(4)

		bodySizeBuff := bytes.NewBuffer(bodySizeBytes)
		var bodySize uint32
		err = binary.Read(bodySizeBuff, binary.LittleEndian, &bodySize)
		if err != nil {
			if err == io.EOF {
				fmt.Println("binary.Read error.跳出循环", err)
				break // 跳出循环
			} else {
				fmt.Println("binary.Read error.", err)
				continue
			}
		}

		dataPacker := NewDataPack()
		pkgLength := bodySize + dataPacker.GetHeadLen() //包头长度24

		// Buffered返回缓冲中现有的可读取的字节数。
		if uint32(reader.Buffered()) < pkgLength {
			fmt.Println("binary.Read 读取完整包失败", reader.Buffered(), pkgLength)
			continue
		}

		//conn.SetReadDeadline(time.Now().Add(time.Duration(10) * time.Second))

		recdata := make([]byte, pkgLength)
		if _, err := io.ReadFull(reader, recdata); err != nil {
			fmt.Println("read msg data error ", err)
			break
		}

		//拆包
		recMsg, err := dataPacker.Unpack(recdata)
		if err != nil {
			fmt.Println("unpack error ", err)
			break
		}

		body := recMsg.GetBody()

		fmt.Printf("server call back for %s :\r\n%s \r\nbody-length = %d\n", name, string(body), len(body))

		ret := dto.Result{
			Seqno: "",
			Cmd:   "response_software_upgrade",
		}

		data, err := json.Marshal(ret)
		if err != nil {
			fmt.Printf("响应结果序列化出错，" + err.Error())
		}

		msg := NewMsgPackage(data)

		wdata, err := dataPacker.Pack(msg)

		if err != nil {
			fmt.Println("Pack error ", err)
			return
		}

		_, err = conn.Write(wdata)
		if err != nil {
			fmt.Println("write error err ", err)
			return
		}
	}
}

func TestServerV0_3(t *testing.T) {
	//创建一个server句柄
	s := NewServer()

	s.AddRouter("ping", &PingRouter{})

	//2 开启服务
	go s.Serve()

	//s.Stop()

	for {
		time.Sleep(5 * time.Second)
		fmt.Println("...")
	}
}

//ping test 自定义路由
type PingRouter struct {
	BaseRouter
}

//Test PreHandle
func (this *PingRouter) PreHandle(req iface.IRequest) {
	body := req.GetMsg().GetBody()
	fmt.Println(string(body))
}

//Test Handle
func (this *PingRouter) Handle(request iface.IRequest) error {
	fmt.Println("response pong msg succ")
	err := request.GetConnection().SendMsg([]byte(`{
    "status": 0,
    "cmd":"pong",
    "seqno":"",
    "msg": ""
}`))
	if err != nil {
		fmt.Println("call back ping ping ping error")
	}
	return errors.New("aaabbb-eer")
}

func TestTimer(t *testing.T) {
	t1 := time.NewTimer(time.Second * 2)
	// ticker := time.NewTicker(3 * time.Second)

	// ticker.Reset(5 * time.Second)
	go func() {
		for {
			<-t1.C
			fmt.Println("abcd")
			t1.Stop()
			t1.Reset(1 * time.Second)

		}
	}()

	var a string
	fmt.Scan(&a)
}

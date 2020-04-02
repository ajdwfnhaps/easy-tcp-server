package dto

// Result Json 返回类型
type Result struct {
	Status int         `json:"status"`
	Cmd    string      `json:"cmd"`
	Msg    string      `json:"msg"`
	Seqno  string      `json:"seqno"`
	Data   interface{} `json:"data,omitempty"`
}

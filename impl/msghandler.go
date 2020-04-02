package impl

import (
	"github.com/ajdwfnhaps/easy-tcp-server/iface"
	"github.com/ajdwfnhaps/easy-tcp-server/utils"
)

type MsgHandle struct {
	Apis           map[string]iface.IRouter //存放每个MsgId 所对应的处理方法的map属性
	WorkerPoolSize uint32                   //业务工作Worker池的数量
	TaskQueue      []chan iface.IRequest    //Worker负责取任务的消息队列

}

func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		Apis:           make(map[string]iface.IRouter),
		WorkerPoolSize: utils.GlobalObject.WorkerPoolSize,
		//一个worker对应一个queue
		TaskQueue: make([]chan iface.IRequest, utils.GlobalObject.WorkerPoolSize),
	}
}

//SendMsgToTaskQueue 将消息交给TaskQueue,由worker进行处理
func (mh *MsgHandle) SendMsgToTaskQueue(request iface.IRequest) {
	//根据ConnID来分配当前的连接应该由哪个worker负责处理
	//轮询的平均分配法则

	//得到需要处理此条连接的workerID
	workerID := request.GetConnection().GetConnID() % mh.WorkerPoolSize
	utils.GlobalObject.Logger.Info("ConnID=", request.GetConnection().GetConnID(), " request-cmd=", request.GetRouterCmd(), "分配给 workerID=", workerID, " 来处理")
	//将请求消息发送给任务队列
	mh.TaskQueue[workerID] <- request
}

//DoMsgHandler 马上以非阻塞方式处理消息
func (mh *MsgHandle) DoMsgHandler(request iface.IRequest) {
	cmd := request.GetRouterCmd()
	//utils.GlobalObject.Logger.Info("msg cmd=", cmd)
	handler, ok := mh.Apis[cmd]
	if !ok {
		errMsg := "handler cmd = " + cmd + " is not FOUND!"
		utils.GlobalObject.Logger.Error(errMsg)
		request.GetConnection().SendMsg([]byte(`{
			"status": -1,
			"cmd":"unkown-action",
			"seqno":"",
			"msg": "` + errMsg + `"
	}`))
		return
	}

	//执行对应处理方法
	handler.PreHandle(request)

	if err := handler.Handle(request); err != nil {
		ret := request.GetRet()
		errMsg := err.Error()
		request.GetConnection().SendMsg([]byte(`{
			"status": -1,
			"cmd":"` + ret.Cmd + `",
			"seqno":"` + ret.Seqno + `",
			"msg": "` + errMsg + `"
	}`))

		utils.GlobalObject.Logger.Errorf("DoMsgHandler Err: %s", errMsg)
		return
	}

	handler.PostHandle(request)
}

//AddRouter 为消息添加具体的处理逻辑
func (mh *MsgHandle) AddRouter(cmd string, router iface.IRouter) {
	//1 判断当前msg绑定的API处理方法是否已经存在
	if _, ok := mh.Apis[cmd]; ok {
		panic("repeated api , cmd = " + cmd)
	}
	//2 添加msg与api的绑定关系
	mh.Apis[cmd] = router
	if utils.GlobalObject.Logger != nil {
		utils.GlobalObject.Logger.Info("Add tcp handler cmd = ", cmd)
	}
}

//StartOneWorker 启动一个Worker工作流程
func (mh *MsgHandle) StartOneWorker(workerID int, taskQueue chan iface.IRequest) {
	utils.GlobalObject.Logger.Info("Tcp-Worker ID = ", workerID, " is started.")
	//不断的等待队列中的消息
	for {
		select {
		//有消息则取出队列的Request，并执行绑定的业务方法
		case request := <-taskQueue:
			mh.DoMsgHandler(request)
		}
	}
}

//StartWorkerPool 启动worker工作池
func (mh *MsgHandle) StartWorkerPool() {
	//遍历需要启动worker的数量，依此启动
	for i := 0; i < int(mh.WorkerPoolSize); i++ {
		//一个worker被启动
		//给当前worker对应的任务队列开辟空间
		mh.TaskQueue[i] = make(chan iface.IRequest, utils.GlobalObject.MaxWorkerTaskLen)
		//启动当前Worker，阻塞的等待对应的任务队列是否有消息传递进来
		go mh.StartOneWorker(i, mh.TaskQueue[i])
	}
}

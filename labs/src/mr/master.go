package mr

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

const (
	IDLE        = 0
	IN_PROGRESS = 1
	COMPLETED   = 2
	MAP         = 0
	REDUCE      = 1
)

type Task struct {
	lock      sync.Mutex
	filename  string
	state     int
	timestamp time.Time
}

type Master struct {
	mu     sync.Mutex
	remain int
	phase  int
	mtasks []*Task
	rtasks []*Task
}

// Your code here -- RPC handlers for the worker to call.

// give the asking worker a task if possible
// otherwise tell the worker there's no work for him/her to do
func (m *Master) HandleAsk(args *AskArgs, reply *AskReply) error {
	// TODO
	return nil
}

// receive response from a worker, ignore it if the worker's performing time exceeds 10s
func (m *Master) HandleResponse(args *ResponseArgs, reply *ResponseReply) error {
	// TODO
	return nil
}

//
// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
//
func (m *Master) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

//
// start a thread that listens for RPCs from worker.go
//
func (m *Master) server() {
	rpc.Register(m)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := masterSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrmaster.go calls Done() periodically to find out
// if the entire job has finished.
//
func (m *Master) Done() bool {
	ret := false

	m.mu.Lock()
	if m.remain == 0 {
		ret = true
	}
	m.mu.Unlock()

	return ret
}

//
// create a Master.
// main/mrmaster.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeMaster(files []string, nReduce int) *Master {
	master := Master{}
	master.mu = sync.Mutex{}
	master.remain = nReduce
	master.phase = MAP

	// initialize master data structure
	for _, file := range files {
		mtask := new(Task)
		mtask.lock = sync.Mutex{}
		mtask.filename = file
		mtask.state = IDLE
		master.mtasks = append(master.mtasks, mtask)
	}

	for i := 0; i < nReduce; i++ {
		master.rtasks[i] = new(Task)
		master.rtasks[i].lock = sync.Mutex{}
	}

	fmt.Printf("Master initialization completed\n")

	master.server()
	return &master
}

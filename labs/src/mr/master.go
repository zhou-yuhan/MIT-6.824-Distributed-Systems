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
	mu            sync.Mutex
	map_remain    int
	reduce_remain int
	mtasks        []*Task
	rtasks        []*Task
}

// Your code here -- RPC handlers for the worker to call.

// for each allocated task, the master waits for 10s
// after 10s, the master checks if the task has been completed
// if the task has not been finished by the worker in 10s, the master gives up
func wait(task *Task) {
	time.Sleep(10 * time.Second)

	task.lock.Lock()
	if task.state == COMPLETED {
		fmt.Printf("Master: task %s completed\n", task.filename)
	} else {
		task.state = IDLE
		fmt.Printf("Master: task %s failed, re-allocate to othter workers\n", task.filename)
	}
	task.lock.Unlock()
}

// give the asking worker a task if possible
// otherwise tell the worker there's no work for him/her to do
func (m *Master) HandleAsk(args *AskArgs, reply *AskReply) error {
	reply.kind = "none"
	if m.map_remain != 0 {
		// look for a map task
		for i, task := range m.mtasks {
			task.lock.Lock()
			defer task.lock.Unlock()
			if task.state == IDLE {
				task.state = IN_PROGRESS
				reply.kind = "map"
				reply.file = task.filename
				reply.nReduce = len(m.rtasks)
				reply.index = i
				task.timestamp = time.Now()
				go wait(task) // start timer
				break
			}
		}
	} else {
		// look for a reduce task
		for i, task := range m.rtasks {
			task.lock.Lock()
			defer task.lock.Unlock()
			if task.state == IDLE {
				task.state = IN_PROGRESS
				reply.kind = "reduce"
				reply.splite = len(m.mtasks)
				reply.index = i
				task.timestamp = time.Now()
				go wait(task) // start timer
				break
			}
		}
	}
	return nil
}

// receive response from a worker, ignore it if the worker's performing time exceeds 10s
func (m *Master) HandleResponse(args *ResponseArgs, reply *ResponseReply) error {
	now := time.Now()
	var task *Task
	if args.kind == "map" {
		task = m.mtasks[args.index]
	} else {
		task = m.rtasks[args.index]
	}

	if now.Before(task.timestamp.Add(10 * time.Second)) {
		task.lock.Lock()
		task.state = COMPLETED
		task.lock.Unlock()
		// a task is completed, decrease remain count
		m.mu.Lock()
		if args.kind == "map" {
			m.map_remain--
		} else {
			m.reduce_remain--
		}
		m.mu.Unlock()
	}
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
	if m.reduce_remain == 0 {
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
	master.mtasks = make([]*Task, len(files))
	master.rtasks = make([]*Task, nReduce)
	master.mu = sync.Mutex{}
	master.map_remain = len(files)
	master.reduce_remain = nReduce

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

	fmt.Printf("Master: initialization completed\n")

	master.server()
	return &master
}

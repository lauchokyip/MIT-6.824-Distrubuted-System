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

const timeout = time.Second * 10

type Master struct {
	sync.Mutex
	// Your definitions here.
	MapTasksPool    *TaskPool
	ReduceTasksPool *TaskPool
}

func NewMaster(mtaskpool, rtaskpool *TaskPool) *Master {
	return &Master{
		MapTasksPool:    mtaskpool,
		ReduceTasksPool: rtaskpool,
	}
}

// Your code here -- RPC handlers for the worker to call.
func (m *Master) GetTask(args *GetTaskArgs, reply *GetTaskReply) error {
	// if there is no more idle tasks but still has
	// incompleted tasks
	if m.MapTasksPool.HasIncompletedOrIdleTasks() {
		mapTask, err := m.MapTasksPool.GetTaskFromTaskPool(args.WorkerID)
		if err != nil {
			return err
		}
		if mapTask != nil {
			reply.Action = MAP
			reply.T = *mapTask

			go m.MapTasksPool.MarkTaskAsIdleAfterTimeout(
				mapTask.TaskID,
				timeout,
			)
		} else {
			reply.Action = WAIT
		}

		// only proceed to reduce tasks if map tasks have been completed
	} else if m.ReduceTasksPool.HasIncompletedOrIdleTasks() {
		reduceTask, err := m.ReduceTasksPool.GetTaskFromTaskPool(args.WorkerID)
		if err != nil {
			return err
		}
		if reduceTask != nil {
			reply.Action = REDUCE
			reply.T = *reduceTask

			go m.ReduceTasksPool.MarkTaskAsIdleAfterTimeout(
				reduceTask.TaskID,
				timeout,
			)
		} else {
			reply.Action = WAIT
		}
		// Reduce and Map tasks are completed
	} else {
		reply.Action = END
	}
	return nil
}

func (m *Master) ReportTaskDone(args *ReportTaskDoneArgs, reply *ReportTaskDoneReply) error {
	var err error

	switch args.Action {
	case MAP:
		m.MapTasksPool.ReportTaskDone(args.TaskID, args.WorkerID)
	case REDUCE:
		m.ReduceTasksPool.ReportTaskDone(args.TaskID, args.WorkerID)
	default:
		err = fmt.Errorf("invalid action %d", args.Action)
	}

	return err
}

// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
func (m *Master) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

// start a thread that listens for RPCs from worker.go
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

// main/mrmaster.go calls Done() periodically to find out
// if the entire job has finished.
func (m *Master) Done() bool {
	ret := false

	// Your code here.

	return ret
}

// create a Master.
// main/mrmaster.go calls this function.
// nReduce is the number of reduce tasks to use, which is also
// equivalent of the number of intermediate files
func MakeMaster(files []string, nReduce int) *Master {
	m := NewMaster(NewTaskPool(), NewTaskPool())

	m.MapTasksPool.InitMapTasks(files)
	m.ReduceTasksPool.InitReduceTasks(nReduce)

	m.server()
	return m
}

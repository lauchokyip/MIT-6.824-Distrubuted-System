package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	IDLE States = iota
	IN_PROGRESS
	COMPLETED
)

type States int

type Task struct {
	Filename string
	State    States
	WorkerID int
	TaskID   int
}

type TaskType int

type Tasks []*Task

type TaskPool struct {
	sync.Mutex
	Tasks Tasks

	workerIDtoTaskID            map[int]string
	numOfIncompletedOrIdleTasks uint
}

func NewTaskPool() *TaskPool {
	return &TaskPool{
		Tasks:            Tasks{},
		workerIDtoTaskID: map[int]string{},
	}
}

func (t *TaskPool) InitMapTasks(files []string) {
	for i, f := range files {
		t.Tasks = append(t.Tasks, &Task{
			Filename: f,
			State:    IDLE,
			WorkerID: -1,
			TaskID:   i,
		})

		t.numOfIncompletedOrIdleTasks++
	}
}

func (t *TaskPool) InitReduceTasks(nReduce int) {
	for i := 0; i < nReduce; i++ {
		t.Tasks = append(t.Tasks, &Task{
			State:    IDLE,
			WorkerID: -1,
			TaskID:   i,
		})

		t.numOfIncompletedOrIdleTasks++
	}
}

// GetTaskFromTaskPool returns a task from the Task Pool.
// It will return nil if the Tasks have not been set up or
// there is no idle taks
func (t *TaskPool) GetTaskFromTaskPool(workerID int) (*Task, error) {
	t.Lock()
	defer t.Unlock()

	if t.Tasks == nil {
		return nil, errors.New("taskpool has not been initialized")
	}

	for _, task := range t.Tasks {
		if task.State == IDLE {
			if taskID := t.workerIDtoTaskID[workerID]; taskID != "" {
				return nil, fmt.Errorf("the worker %d has been assigned a task", workerID)
			}

			task.WorkerID = workerID
			task.State = IN_PROGRESS
			t.workerIDtoTaskID[workerID] = fmt.Sprint(task.TaskID)

			return task, nil
		}
	}

	return nil, nil
}

func (t *TaskPool) HasIncompletedOrIdleTasks() bool {
	t.Lock()
	defer t.Unlock()

	return t.numOfIncompletedOrIdleTasks > 0
}

func (t *TaskPool) MarkTaskAsIdleAfterTimeout(taskID int, duration time.Duration) {
	timer := time.NewTimer(duration)
	<-timer.C

	t.Lock()
	defer t.Unlock()

	t.Tasks[taskID].State = IDLE
	tempWorkerID := t.Tasks[taskID].WorkerID // remove the task from the worker mapping
	t.workerIDtoTaskID[tempWorkerID] = ""
	t.Tasks[taskID].WorkerID = -1

}

func (t *TaskPool) ReportTaskDone(taskID, workerID int) {
	t.Lock()
	defer t.Unlock()

	if taskID < len(t.Tasks) {
		if t.Tasks[taskID].WorkerID == workerID {
			if t.numOfIncompletedOrIdleTasks > 0 {
				t.workerIDtoTaskID[workerID] = ""
				t.Tasks[taskID].State = COMPLETED
				t.numOfIncompletedOrIdleTasks--
			} else {
				fmt.Printf("WARNING: less than 0 Incompleted or IDLE tasks reported by %d", workerID)
			}
		} else {
			fmt.Printf("WARNING: wrong worker %d reported task done\n", workerID)
		}
	} else {
		fmt.Println("WARNING: given taskID is out of bound.")
	}
}

////////////////////////////////////////
//
// Worker action
//
////////////////////////////////////////

type Actions int

const (
	NONE Actions = iota
	MAP
	REDUCE
	WAIT
	END
)

////////////////////////////////////////
//
// RPC struct
//
////////////////////////////////////////

// ReplyTask is a struct
// that the worker will receive
// to identify a Task
type GetTaskReply struct {
	Action Actions
	T      Task
}

type GetTaskArgs struct {
	WorkerID int
}

// ReportTaskDone
type ReportTaskDoneReply struct {
}

type ReportTaskDoneArgs struct {
	WorkerID int
	TaskID   int

	Action Actions // not the best design but it's fine
}

// example to show how to declare the arguments
// and reply for an RPC.

type ExampleArgs struct {
	X int
}
type ExampleReply struct {
	Y int
}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the master.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func masterSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}

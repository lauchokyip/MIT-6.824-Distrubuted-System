package mr

import (
	"fmt"
	"hash/fnv"
	"log"
	"net/rpc"
	"os"
	"time"
)

const waitime = 1

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

// main/mrworker.go calls this function.
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	// Your worker implementation here.
	task, action, rpcErr := GetTaskFromMaster()
	if rpcErr {
		return
	}

	switch action {
	case MAP:
		doMap(task, mapf)
		ReportTaskDoneToMaster(task, MAP)
	case REDUCE:
		doReduce()
		ReportTaskDoneToMaster(task, REDUCE)
	case WAIT:
		time.Sleep(waitime * time.Second)
	case END, NONE:
		fallthrough
	default:
		return
	}

	// uncomment to send the Example RPC to the master.
	// CallExample()
}

func doMap(task Task, mapf func(string, string) []KeyValue) {
	var intermediate []KeyValue
	content, err := os.ReadFile(task.Filename)
	if err != nil {
		log.Fatalf("cannot read content for file %v", task.Filename)
	}
	intermediate = mapf(task.Filename, string(content))

	fmt.Println(intermediate)

}

func doReduce() {

}

func ReportTaskDoneToMaster(task Task, action Actions) bool {
	args := ReportTaskDoneArgs{
		WorkerID: os.Getpid(),
		TaskID:   task.TaskID,
		Action:   action,
	}
	reply := ReportTaskDoneReply{}
	rpcErr := call("Master.ReportTaskDone", &args, &reply)

	return rpcErr
}

func GetTaskFromMaster() (task Task, action Actions, rpcErr bool) {
	args := GetTaskArgs{WorkerID: os.Getpid()}
	reply := GetTaskReply{}
	rpcErr = call("Master.GetTask", &args, &reply)
	return reply.T, reply.Action, rpcErr
}

// example function to show how to make an RPC call to the master.
//
// the RPC argument and reply types are defined in rpc.go.
func CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	call("Master.Example", &args, &reply)

	// reply.Y should be 100.
	fmt.Printf("reply.Y %v\n", reply.Y)
}

// send an RPC request to the master, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := masterSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}

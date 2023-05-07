# MIT 6.824 Lab 1: Map Reduce



A lot of time people don't document how they came out with the solutions but only provide the final outcome. I hope this would clearly illustrate all the questions I came across before I started my implementation

Problems/Questions I faced:
1) How to name RPC function 
I think it's hard to treat/definite a rpc function like a normal function, for example if you name it as `send_task` and when the worker calls the function `send_task` when somebody reading the worker code, it would think the worker is sending task to somewhere but it's actually receiving the task.  
**It's probably best to think of it as a REST endpoint, for in that case, I would name the function `get_task`**

2) How can the worker knows when to stop ?
It's important to know that Reduce tasks cannot be started before Map tasks, so I am thinking of workers keep on asking for the master and if there is no task it will then `time.Sleep` for a few seconds, and ask the master for the task again. 
Worker will only quit when it cannot contact master anymore

3) How does the worker knows the reduce task that's assigned since each intermediate files would have the same key. For example `intermediate_1` has A 10 and `intermediate_2` has A 20?
This kinda answer question number 5 too, I have to think of it as an 2d array. That's also where the `ihash` function comes in. It would be easier to understand it with a given example. 

Imagine if Map_0 has A, and `ihash` output is 123 and we take the mod 10 (123 % 10), Map_1 also has A, and `ihash` output is the same as 123 and we take the mod 10 (123 % 10), Map_0 will write to mr-0-3 and Map_1 will write to mr-1-3, if the worker is given reduce task 3, then it can combine everything together


4) What if the worker crashes after finish writing the files but master doesn't know?
It doesn't matter because if the Map Tasks did not finish and the master has not received it, it will assign the Map Tasks to another worker and only if they are completed, it will then be saved in the Reduce Tasks 


5) The paper mentions about needing the identity of the worker machines for non-idle tasks why?
Imagine if there is a case where the master assigns the task to a slow worker with slow processing power and slow network, if we don't tie the worker ID with a task, when the master assigns the task again to a new worker, it has to know if it's the newly-assigned worker that finish it or the previous-assigned worker 

6) How does the worker know if it has to wait?
Master will send a struct 
```go
type Actions int

const (
	NONE Actions = iota
	MAP
	REDUCE
	WAIT
	END
)
```
Worker will do the action that the master instructs

7) How does the intermediate files look like for MapReduce?



Next Step:
[12/4] Modifying GetTask to handle idle tasks and incompleted task
[12/21] Write unit test for GetTask
[12/28] Work on worker logic
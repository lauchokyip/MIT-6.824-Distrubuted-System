package mr

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMasterAssignTaskLogic(t *testing.T) {
	m := NewMaster(NewTaskPool(), NewTaskPool())

	// Initialize two map taks and 2 reduce tasks
	m.MapTasksPool.InitMapTasks([]string{"foo", "bar"})
	m.ReduceTasksPool.InitReduceTasks(2)

	args := GetTaskArgs{WorkerID: 0}
	reply := GetTaskReply{}
	err := m.GetTask(&args, &reply)
	require.NoError(t, err)
	require.Equal(t, MAP, reply.Action)
	require.Equal(t, 0, reply.T.TaskID)
	require.Equal(t, 0, reply.T.WorkerID)

	args = GetTaskArgs{WorkerID: 1}
	reply = GetTaskReply{}
	err = m.GetTask(&args, &reply)
	require.NoError(t, err)
	require.Equal(t, MAP, reply.Action)
	require.Equal(t, 1, reply.T.TaskID)
	require.Equal(t, 1, reply.T.WorkerID)

	// Third worker asking for task
	args = GetTaskArgs{WorkerID: 2}
	reply = GetTaskReply{}
	err = m.GetTask(&args, &reply)
	require.NoError(t, err)
	require.Equal(t, WAIT, reply.Action)
	require.Equal(t, 0, reply.T.TaskID)
	require.Equal(t, 0, reply.T.WorkerID)

	// wait for 10 seconds and see if the task will be newly assigned
	time.Sleep(10 * time.Second)

	args = GetTaskArgs{WorkerID: 2}
	reply = GetTaskReply{}
	err = m.GetTask(&args, &reply)
	require.NoError(t, err)
	require.Equal(t, MAP, reply.Action)
	require.Equal(t, 0, reply.T.TaskID)
	require.Equal(t, 2, reply.T.WorkerID)

	// wait for 1 more second, bar should be available,
	// but what if it's the same worker asking, can't be greedy
	time.Sleep(1 * time.Second)

	args = GetTaskArgs{WorkerID: 2}
	reply = GetTaskReply{}
	err = m.GetTask(&args, &reply)
	require.EqualError(t, err, "the worker 2 has been assigned a task")
	require.Equal(t, NONE, reply.Action)
	require.Equal(t, 0, reply.T.TaskID)
	require.Equal(t, 0, reply.T.WorkerID)

	// Worker 0 should be able to take the task again
	args = GetTaskArgs{WorkerID: 0}
	reply = GetTaskReply{}
	err = m.GetTask(&args, &reply)
	require.NoError(t, err)
	require.Equal(t, MAP, reply.Action)
	require.Equal(t, 1, reply.T.TaskID)
	require.Equal(t, 0, reply.T.WorkerID)

	// simulate tasks are finished
	// wrong worker reports the wrong tasks
	m.ReportTaskDone(&ReportTaskDoneArgs{WorkerID: 2, TaskID: 1, Action: MAP}, &ReportTaskDoneReply{})
	m.ReportTaskDone(&ReportTaskDoneArgs{WorkerID: 1, TaskID: 0, Action: MAP}, &ReportTaskDoneReply{})
	m.ReportTaskDone(&ReportTaskDoneArgs{WorkerID: 1, TaskID: 1, Action: MAP}, &ReportTaskDoneReply{})
	m.ReportTaskDone(&ReportTaskDoneArgs{WorkerID: 1, TaskID: 5, Action: MAP}, &ReportTaskDoneReply{})

	// properly report tasks done
	m.ReportTaskDone(&ReportTaskDoneArgs{WorkerID: 2, TaskID: 0, Action: MAP}, &ReportTaskDoneReply{})
	m.ReportTaskDone(&ReportTaskDoneArgs{WorkerID: 0, TaskID: 1, Action: MAP}, &ReportTaskDoneReply{})

	// Should return reduce tasks
	args = GetTaskArgs{WorkerID: 0}
	reply = GetTaskReply{}
	err = m.GetTask(&args, &reply)
	require.NoError(t, err)
	require.Equal(t, REDUCE, reply.Action)
	require.Equal(t, 0, reply.T.TaskID)
	require.Equal(t, 0, reply.T.WorkerID)

	args = GetTaskArgs{WorkerID: 2}
	reply = GetTaskReply{}
	err = m.GetTask(&args, &reply)
	require.NoError(t, err)
	require.Equal(t, REDUCE, reply.Action)
	require.Equal(t, 1, reply.T.TaskID)
	require.Equal(t, 2, reply.T.WorkerID)

	args = GetTaskArgs{WorkerID: 1}
	reply = GetTaskReply{}
	err = m.GetTask(&args, &reply)
	require.NoError(t, err)
	require.Equal(t, WAIT, reply.Action)
	require.Equal(t, 0, reply.T.TaskID)
	require.Equal(t, 0, reply.T.WorkerID)

	// properly report tasks done
	m.ReportTaskDone(&ReportTaskDoneArgs{WorkerID: 0, TaskID: 0, Action: REDUCE}, &ReportTaskDoneReply{})
	m.ReportTaskDone(&ReportTaskDoneArgs{WorkerID: 2, TaskID: 1, Action: REDUCE}, &ReportTaskDoneReply{})

	args = GetTaskArgs{WorkerID: 0}
	reply = GetTaskReply{}
	err = m.GetTask(&args, &reply)
	require.NoError(t, err)
	require.Equal(t, END, reply.Action)
	require.Equal(t, 0, reply.T.TaskID)
	require.Equal(t, 0, reply.T.WorkerID)

	args = GetTaskArgs{WorkerID: 2}
	reply = GetTaskReply{}
	err = m.GetTask(&args, &reply)
	require.NoError(t, err)
	require.Equal(t, END, reply.Action)
	require.Equal(t, 0, reply.T.TaskID)
	require.Equal(t, 0, reply.T.WorkerID)

	args = GetTaskArgs{WorkerID: 1}
	reply = GetTaskReply{}
	err = m.GetTask(&args, &reply)
	require.NoError(t, err)
	require.Equal(t, END, reply.Action)
	require.Equal(t, 0, reply.T.TaskID)
	require.Equal(t, 0, reply.T.WorkerID)
}

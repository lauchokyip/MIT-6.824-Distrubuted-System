package mr

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTaskPool(t *testing.T) {
	dummyTaskPool := NewTaskPool()
	dummyTaskPool.InitReduceTasks(2)

	var wg sync.WaitGroup

	wg.Add(3)
	go func() {
		defer wg.Done()

		fmt.Printf("GoRoutine 1: HasIncompletedOrIDLETasks: %v\n", dummyTaskPool.HasIncompletedOrIdleTasks())
		task, err := dummyTaskPool.GetTaskFromTaskPool(0)
		require.NoError(t, err)
		fmt.Printf("GoRoutine 1: Got task %#v\n", task)
		fmt.Printf("GoRoutine 1: HasIncompletedOrIDLETasks: %v\n", dummyTaskPool.HasIncompletedOrIdleTasks())
	}()
	go func() {
		defer wg.Done()

		fmt.Printf("GoRoutine 2: HasIncompletedOrIDLETasks: %v\n", dummyTaskPool.HasIncompletedOrIdleTasks())
		task, err := dummyTaskPool.GetTaskFromTaskPool(1)
		require.NoError(t, err)
		fmt.Printf("GoRoutine 2: Got task %#v\n", task)
		fmt.Printf("GoRoutine 2: HasIncompletedOrIDLETasks: %v\n", dummyTaskPool.HasIncompletedOrIdleTasks())
	}()

	go func() {
		defer wg.Done()

		fmt.Printf("GoRoutine 3: HasIncompletedOrIDLETasks: %v\n", dummyTaskPool.HasIncompletedOrIdleTasks())
		task, err := dummyTaskPool.GetTaskFromTaskPool(2)
		require.NoError(t, err)
		fmt.Printf("GoRoutine 3: Got task %#v\n", task)
		fmt.Printf("GoRoutine 3: HasIncompletedOrIDLETasks: %v\n", dummyTaskPool.HasIncompletedOrIdleTasks())
	}()

	wg.Wait()
}

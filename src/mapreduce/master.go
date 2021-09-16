package mapreduce

import (
	"container/list"
	"fmt"
	"sync"
)

type WorkerInfo struct {
	address string
	// You can add definitions here.
}

// Clean up all workers by sending a Shutdown RPC to each one of them Collect
// the number of jobs each work has performed.
// 清楚所有的worker
func (mr *MapReduce) KillWorkers() *list.List {
	l := list.New()
	for _, w := range mr.Workers {
		DPrintf("DoWork: shutdown %s\n", w.address)
		args := &ShutdownArgs{}
		var reply ShutdownReply
		ok := call(w.address, "Worker.Shutdown", args, &reply)
		if ok == false {
			fmt.Printf("DoWork: RPC %s shutdown error\n", w.address)
		} else {
			l.PushBack(reply.Njobs)
		}
	}
	return l
}

// 主要是要实现这个函数
// 模拟运行master进程，
func (mr *MapReduce) RunMaster() *list.List {
	// Your code here
	// master的工作：
	// 第一：创建workers
	var wg sync.WaitGroup
	wg.Add(mr.nMap)
	for i := 0; i != mr.nMap; i++ {
		doJobArgs := DoJobArgs{
			File:          mr.file,
			Operation:     Map,
			JobNumber:     i,
			NumOtherPhase: mr.nReduce,
		}
		go func(doJobArgs DoJobArgs, registerChan chan string) {
			// 获取一个可用worker的address
			address := <-registerChan
			call(address, "Worker.DoJob", doJobArgs, nil)
			// 释放一个可用worker
			registerChan <- address
			wg.Done()
		}(doJobArgs, mr.registerChannel)
	}
	wg.Wait()
	wg.Add(mr.nReduce)
	for i := 0; i != mr.nMap; i++ {
		doJobArgs := DoJobArgs{
			File:          mr.file,
			Operation:     Reduce,
			JobNumber:     i,
			NumOtherPhase: mr.nMap,
		}
		go func(doJobArgs DoJobArgs, registerChan chan string) {
			address := <-registerChan
			call(address, "Worker.DoJob", doJobArgs, nil)
			registerChan <- address
			wg.Done()
		}(doJobArgs, mr.registerChannel)
	}
	wg.Wait()
	return mr.KillWorkers()
}

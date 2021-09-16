# lab1实验说明

## part1

首先我们要让自己的程序能够跑起来。我把项目放在$GOPATH/src/6.824，然后需要修改wc文件的import为``import 6.824/src/mapreduce``。
然后运行：``go run wc.go master kjv12.txt sequential``，获得如下结果说明程序已经能过跑起来了：
```
# command-line-arguments
.\wc.go:15:1: missing return at end of function
.\wc.go:21:1: missing return at end of function
```
接下来我们看看任务：
建立一个reduce程序和map程序，能统计单词的次数，并且按照字母排序报告结果。
结果希望如下：
```
$ go run wc.go master kjv12.txt sequential
Split kjv12.txt
Split read 4834757
DoMap: read split mrtmp.kjv12.txt-0 966954
DoMap: read split mrtmp.kjv12.txt-1 966953
DoMap: read split mrtmp.kjv12.txt-2 966951
DoMap: read split mrtmp.kjv12.txt-3 966955
DoMap: read split mrtmp.kjv12.txt-4 966944
DoReduce: read mrtmp.kjv12.txt-0-0
DoReduce: read mrtmp.kjv12.txt-1-0
DoReduce: read mrtmp.kjv12.txt-2-0
DoReduce: read mrtmp.kjv12.txt-3-0
DoReduce: read mrtmp.kjv12.txt-4-0
DoReduce: read mrtmp.kjv12.txt-0-1
DoReduce: read mrtmp.kjv12.txt-1-1
DoReduce: read mrtmp.kjv12.txt-2-1
DoReduce: read mrtmp.kjv12.txt-3-1
DoReduce: read mrtmp.kjv12.txt-4-1
DoReduce: read mrtmp.kjv12.txt-0-2
DoReduce: read mrtmp.kjv12.txt-1-2
DoReduce: read mrtmp.kjv12.txt-2-2
DoReduce: read mrtmp.kjv12.txt-3-2
DoReduce: read mrtmp.kjv12.txt-4-2
Merge phaseMerge: read mrtmp.kjv12.txt-res-0
Merge: read mrtmp.kjv12.txt-res-1
Merge: read mrtmp.kjv12.txt-res-2
$ sort -n -k2 mrtmp.kjv12.txt | tail -10
unto: 8940
he: 9666
shall: 9760
in: 12334
that: 12577
And: 12846
to: 13384
of: 34434
and: 38850
the: 62075
```

我们看代码(src/main/wc.go)
```Golang
func main() {
	if len(os.Args) != 4 {
		fmt.Printf("%s: see usage comments in file\n", os.Args[0])
	} else if os.Args[1] == "master" {
		if os.Args[3] == "sequential" {
			// part1会运行到这里
			// 这里是运行一个mapreduce集群，5个map进程、3个reduce进程
			mapreduce.RunSingle(5, 3, os.Args[2], Map, Reduce)
		} else {
			mr := mapreduce.MakeMapReduce(5, 3, os.Args[2], os.Args[3])
			// Wait until MR is done
			<-mr.DoneChannel
		}
	} else {
		mapreduce.RunWorker(os.Args[2], os.Args[3], Map, Reduce, 100)
	}
}
```
结合命令行我们可以看出，这里运行了到了mapreduce.RunSignle，点进去看实现。
```Golang
// Run jobs sequentially.
// 启动一个mapreduce任务
// 传参数map进程数、reduce进程数、输入文件名
// map函数、reduce函数
func RunSingle(nMap int, nReduce int, file string,
	Map func(string) *list.List,
	Reduce func(string, *list.List) string) {
	// 初始化一个mapreduce实例
	mr := InitMapReduce(nMap, nReduce, file, "")
	// 划分文件，命名规则，file0、file1、file2
	mr.Split(mr.file)
	// 启动map程序
	for i := 0; i < nMap; i++ {
		DoMap(i, mr.file, mr.nReduce, Map)
	}
	// 启动reduce程序
	for i := 0; i < mr.nReduce; i++ {
		DoReduce(i, mr.file, mr.nMap, Reduce)
	}
	mr.Merge()
}
```
这里不难理解，点进去看一下，DoMap的DoReduce给map函数和reduce函数的输入和输出即可知道如何实现。

## part2

这个部分是要实现一个版本的mapreduce，测试方式是，进入目录mapreduce，然后运行``go test``
```Linux
$ cd src/mapreduce
$ go test
```

通过test_test.go文件的第一个测试（Basic mapreduce），即通过part2。
我们需要修改的文件是master.go。这个目录下每个文件的作用是：worker.go是工作代码，启动工作代码和处理RPC消息的代码是common.go、
我们可以看一下mapreduce的执行过程，再mapreduce.go中的Run()中，先调用split将输入分割成map-job文件，
然后调用RunMaster()来将每个reduce-job文件输出组装成当个输出文件。
观察测试代码TestBasic，我们发现这里启动了两个worker，根据mapreduce的论文，整个分布式计算过程，是master通知worker完成，worker可以
进行map进程也可以进行reduce进程。这里的两个进程应该是让我们交替使用。

所以我们其实只需要在RunMaster中完成对worker的调度，使其运行完成map进程和reduce进程：
```Golang

func (mr *MapReduce) RunMaster() *list.List {
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
			go func(){
				registerChan <- address
			}()
			wg.Done()
		}(doJobArgs, mr.registerChannel)
	}
	wg.Wait()
	wg.Add(mr.nReduce)
	for i := 0; i != mr.nReduce; i++ {
		doJobArgs := DoJobArgs{
			File:          mr.file,
			Operation:     Reduce,
			JobNumber:     i,
			NumOtherPhase: mr.nMap,
		}
		go func(doJobArgs DoJobArgs, registerChan chan string) {
			address := <-registerChan
			call(address, "Worker.DoJob", doJobArgs, nil)
			go func(){
				registerChan <- address
			}()
			wg.Done()
		}(doJobArgs, mr.registerChannel)
	}
	wg.Wait()
	return mr.KillWorkers()
}
```
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



在本部分中，您将完成一个版本的mapreduce，它将工作划分到一组工作线程上，以便利用多个核。
主线把工作交给工人，并等待他们完成。主服务器应该通过RPC与工作服务器通信。
我们给你工作代码(mapreduce/worker.go)、启动工作代码和处理RPC消息的代码(mapreduce/common.go)。
你的工作就是完成master。进入mapreduce包。特别是，您应该在master中修改RunMaster()。
去分发地图，减少工人的工作，只有当所有的工作都完成后才回来。
看看mapreduce.go中的Run()。它调用Split()将输入分割成每个map-job文件，
然后调用RunMaster()来运行map和reduce job，
然后调用Merge()将每个reduce-job输出组装成单个输出文件。
RunMaster只需要告诉工人原始输入文件的名称(mr.file)和工号;每个worker都知道从哪些文件读取其输入，
向哪些文件写入其输出。 每个worker在启动时向master发送一个Register RPC。mapreduce。
go已经实现了master的MapReduce。
为您注册RPC处理程序，并将新工作人员的信息传递给registerchannel先生。
你的RunMaster应该通过读取这个通道来处理新的worker注册。
关于MapReduce作业的信息在MapReduce结构体中，
在MapReduce .go中定义。修改MapReduce结构体以跟踪任何附加的状态(例如，可用的worker的集合)，
并在InitMapReduce()函数中初始化这个附加的状态。master不需要知道哪个Map或Reduce函数被用于工作;
工人将负责执行正确的Map或Reduce代码。
您应该使用Go的单元测试系统运行代码。我们在test_test.go中为您提供了一组测试。在包目录(例如mapreduce目录)中运行单元测试，如下所示:
$ cd mapreduce
当实现通过test_test中的第一个测试(“Basic mapreduce”测试)时，就完成了第二部分。进入mapreduce包。你还不用担心工人的失败。
主服务器应该并行地将rpc发送给worker，以便worker可以同时处理作业。您会发现go语句和go RPC文档对此非常有用。
主人可能不得不等待工人完成工作，然后才能给他们更多的工作。您可能会发现，一旦应答到达，通道可以将等待应答的线程与主服务器同步。通道在Go中的并发性文档中有解释。
跟踪bug最简单的方法是插入log.Printf()语句，使用go test &gt;输出，然后考虑输出是否与您对代码行为的理解相匹配。最后一步(思考)是最重要的。
我们提供的代码将工人作为单个UNIX进程中的线程运行，并且可以利用单个机器上的多个核心。为了在通过网络进行通信的多台机器上运行工作程序，需要进行一些修改。rpc将不得不使用TCP而不是unix域套接字;需要一种方法来启动所有机器上的工作进程;所有的机器都必须通过某种网络文件系统共享存储空间。
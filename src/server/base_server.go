package server

import (
	"global"

	"fmt"
	"net"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"github.com/golang/glog"
)

var ActiveClientCount int32 = 0 //当前客户端连接数
var MaxClientCount int32 = 0
var activeConnStatus bool = false
var autoFreeMemory = true                //默认内存回收选项
var freeMemoryPeriod = time.Second * 600 //单位为秒数,自动内存释放周期
var minFreeMemoryTime = 5                // 手动调用内存释放的最小频率
var lastFreeMemoryTime int64 = 0         //最后一次内存释放时间
var mutex sync.Mutex                     //内存释放同步锁

//常量定义
const ()

type IOPServer interface {
	Initialize() (listenAddress string, clientNumber int32, err error)
	Run(conn *net.TCPConn)
}

/*
* 启动主函数
 */
func MainWork(baseServer IOPServer) error {
	listenAddress, clientNumber, err := baseServer.Initialize()
	glog.Info("Load configure over")
	if err != nil {
		return err
	}
	ProtectServer(baseServer, listenAddress, clientNumber)
	return nil
}

func ProtectServer(baseServer IOPServer, listenAddress string, clientNumber int32) {
	defer func() {
		if err := recover(); err != nil {
			glog.Error("Exception: ProtectServer", err)
			debug.PrintStack()
			fmt.Println("Exception: ProtectServer", err)
			fmt.Println("err ActiveClientCount", ActiveClientCount)
		}
	}()
	startServer(baseServer, listenAddress, clientNumber)
}

func startServer(baseServer IOPServer, listenAddress string, clientNumber int32) (err error) {
	// 创建异步数据处理队列
	ConnnectChan := make(chan *net.TCPConn, 1024)
	//建立连接队列计数
	ClientChanCount := make(chan int32)

	//进行服务地址注册
	TcpAddr, err := net.ResolveTCPAddr("tcp", listenAddress)

	if err != nil {
		glog.Errorln(err.Error())
		return err
	}

	//建立地址监听
	listener, err := net.ListenTCP("tcp", TcpAddr)
	if err != nil {
		glog.Errorln(err.Error())
		return err
	}
	glog.Info("Listen ", listenAddress, " is OK.")
	fmt.Println("服务已完成启动任务，初始化活动全部完成!")

	go func() { //统计当前活动的用户数据
		for {
			ActiveClientCount += <-ClientChanCount
			glog.Infoln("current conn num:", ActiveClientCount)
			if ActiveClientCount > MaxClientCount {
				MaxClientCount = ActiveClientCount
				if MaxClientCount%500 == 0 {
					fmt.Println("MaxClientCount 变更为：", MaxClientCount)
				}

			}
		}
	}()

	// 根据服务设置参数，开启数据服务协程
	for i := int32(0); i < clientNumber; i++ {
		//协程开启，开启处理任务Job
		go MainThread(baseServer, ConnnectChan, ClientChanCount, int(i))
	}

	//自动内存释放

	go ServerGC()

	go showActiveClientCount()
	for global.IsRunning() {
		//开始监听活动
		glog.Infoln("accept")
		conn, err := listener.AcceptTCP()
		if err != nil {
			glog.Errorln(err)
			return err
		}
		glog.Infoln(conn.RemoteAddr().String() + " ==> connected.")

		//监听到连接即刻加入队列
		ConnnectChan <- conn
	}
	listener.Close()
	return
}

func freeMemory() {
	if time.Now().Unix()-lastFreeMemoryTime > 30 {
		lastFreeMemoryTime = time.Now().Unix()
		mutex.Lock()
		defer mutex.Unlock()
		runtime.GC()
		debug.FreeOSMemory()
	}
}

func SetFreeMemory(runFreeMemory bool) {
	//若未启动回收，优先执行一次
	if runFreeMemory && !autoFreeMemory {
		freeMemory()
	}
	autoFreeMemory = runFreeMemory
	glog.Infoln("自动内存释放选项设置：", autoFreeMemory)
}

func SetFreeMemoryPeriod(freeMemorySecond int32) {
	//只是设置为300秒以上
	if freeMemorySecond < 300 {
		freeMemorySecond = 300
	}
	freeMemoryPeriod = time.Second * time.Duration(freeMemorySecond)
	glog.Infoln("自动内存释放周期设置为：", freeMemorySecond)
}

func ServerGC() {
	for {
		if autoFreeMemory {
			freeMemory()
			time.Sleep(freeMemoryPeriod)
		} else {
			//每5秒检查一次状态
			time.Sleep(time.Second * 5)
		}

	}
}

func showActiveClientCount() {
	for {
		if activeConnStatus {
			activeConnStatus = false
			//强制执行一次内存回收活动
			freeMemory()
			//fmt.Println("客户端访问已无连接访问!")
		}
		if ActiveClientCount > 0 {
			activeConnStatus = true
		}
		time.Sleep(time.Second * 5)
	}
}

func ProtectThread(baseServer IOPServer, conn *net.TCPConn) {
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
			glog.Error("Exception: ProtectThread", err)
			fmt.Println("Exception: ProtectThread", err)
			fmt.Println("err ActiveClientCount", ActiveClientCount)
		}
	}()

	baseServer.Run(conn)

}

/*
* 处理客户端连接
 */
func MainThread(baseServer IOPServer, ConnnectChan chan *net.TCPConn, ClientChanCount chan int32, id int) (err error) {
	for conn := range ConnnectChan {
		ClientChanCount <- 1 //当前用户+1
		//主任务活动连接操作

		ProtectThread(baseServer, conn)
		//关闭连接操作
		conn.Close()
		glog.Infoln(conn.RemoteAddr(), "is close.")
		ClientChanCount <- -1 //当前用户-1
	}
	return
}

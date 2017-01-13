package server

import (
	"bufio"
	_ "errors"
	"io"
	"net"
	"os"
	"strings"
	"time"
	"utils/config"
	"utils/msgpk"

	"github.com/golang/glog"
)

//name为key
var ClientMap map[string]*ClientStatus

func init() {
	ClientMap = make(map[string]*ClientStatus)
}

type ClientStatus struct {
	IP     string
	Status string
	Line   string
	Conn   *net.TCPConn
}

type Server struct {
}

func MakeMemory(initialMemory int16) {
	t := time.Now()
	temp := make([]byte, 1024*1024*int32(initialMemory))
	temp[0] = 0
	glog.Infoln("++++++++++系统启动预加载内存", initialMemory, "MB+++++++内存分配耗时>", time.Since(t))
}

func (this *Server) Initialize() (listenAddress string, clientNumber int32, err error) {

	//需要使用BaseServer 者进行实现具体内容
	glog.Info("initialize Config start ")
	listenAddress, clientNumber, err = this.initializeProxy()
	return
}

func (this *Server) Run(conn *net.TCPConn) {
	//主流程 [1] 客户端连接
	this.work(conn)
}

func (this *Server) initializeProxy() (listenAddress string, clientNumber int32, err error) {

	//加载服务端端配置
	var initialMemory int16 = 20
	serverCfg := config.ServerCfg{}
	serverCfg, err = config.LoadServerConfigStruct()
	if err != nil {
		return
	}
	listenAddress = serverCfg.ListenAddress
	clientNumber = serverCfg.MaxClientNumber
	initialMemory = serverCfg.InitialMemory

	//设定对应检测周期
	SetFreeMemoryPeriod(serverCfg.FreeMemory.AutoFreeMemoryPeriod)
	//设定自动启动
	SetFreeMemory(serverCfg.FreeMemory.AutoFreeMemory)
	if initialMemory > 0 {
		MakeMemory(initialMemory)
	}
	glog.Info("Load Server config OK.")

	return
}

/**
*业务主函数
 */
func (this *Server) work(conn *net.TCPConn) {

	if conn == nil {
		glog.Infoln("conn is nil")
		return
	}

	for {

		buff := make([]byte, 64)
		n, err := conn.Read(buff)
		//glog.Infoln(string(buff[:n]), buff[:n], "n:", n)
		if err == io.EOF {

			glog.Infoln("The RemoteAddr:%s is closed!\n", conn.RemoteAddr().String())
			break
		}
		if err != nil {
			glog.Infoln(err)
			break
		}
		if n > 5 {
			action, err := msgpk.Unpack(buff[0:n])
			if err != nil {
				glog.Infoln(err)
				m, err := conn.Write(msgpk.Pack([]byte("fail")))
				if err != nil {
					glog.Infoln(err, m)
					break
				}
				continue
			}
			switch action {
			case "open", "close":
				find := MapFind(conn.RemoteAddr().String())
				if find != "" {
					if val, OK := ClientMap[find]; OK {
						val.Status = action
					}
				}
			case "ping": //客户端心跳包
			default: //默认是客户端注册
				if name, ok := ClientMap[action]; ok && (name.Line == "online") {
					mm, err := conn.Write(msgpk.Pack([]byte("fail")))
					if err != nil {
						glog.Infoln(err, mm)
						break
					}

				}
				client := new(ClientStatus)
				client.IP = conn.RemoteAddr().String()
				client.Status = "close"
				client.Line = "online"
				client.Conn = conn
				ClientMap[action] = client
				k, err := conn.Write(msgpk.Pack([]byte("ok")))
				if err != nil {
					glog.Infoln(err, k)
					break
				}

			}
		}
	}
	findName := MapFind(conn.RemoteAddr().String())
	if findName != "" {
		if val, OK := ClientMap[findName]; OK {
			val.Line = "offline"
		}
	}
	glog.Infoln(conn.RemoteAddr().String() + " ======== OVER ========")
	return
}

func MapFind(ip string) (name string) {
	for index, value := range ClientMap {
		if value.IP == ip {
			return index
		}
	}
	return ""
}

//stdin 接受外部命令
func AcceptCommand() {
	reader := bufio.NewReader(os.Stdin)
	for {
		data := make([]byte, 256)
		n, err := reader.Read(data)
		if err != nil {
			glog.Infoln(err)
		}
		command := string(data[:n])
		//glog.Infoln("command", command)
		go HandleCommand(command)
	}
}

func HandleCommand(command string) {
	cmd := strings.TrimRight(command, "\n")
	array := strings.Split(cmd, " ")
	length := len(array)
	//glog.Infoln("-----------------", array, length, ClientMap)
	switch array[0] {
	case "open", "close":
		for i := 1; i < length; i++ {
			//glog.Infoln(strings.EqualFold(array[i], "yang"), []byte(array[i]), []byte("yang"))
			if client, ok := ClientMap[array[i]]; ok {
				n, err := client.Conn.Write(msgpk.Pack([]byte(array[0])))
				if err != nil {
					glog.Infoln(err, n)
				}
			}
		}
	case "status":
		for name, client := range ClientMap {
			glog.Infoln("Name: ", name, " IP:", client.IP, " Status:", client.Status, " Line:", client.Line)
		}
	default:
		glog.Infoln("unkown command:", array[0])
	}
}

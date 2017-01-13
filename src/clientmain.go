// +build client

package main

import (
	"flag"
	"io"
	"net"
	"os"
	"time"
	"utils/msgpk"

	"github.com/golang/glog"
)

func main() {
	flag.Parse()
	flag.Set("log_dir", "logs")
	flag.Set("alsologtostderr", "true")
	if len(os.Args) != 3 {

		glog.Infoln("Usage:<ip> <name>")
		return
	}
	ip := os.Args[1]
	tcpAddr, err := net.ResolveTCPAddr("tcp", ip)
	if err != nil {
		glog.Infoln(err)
		return
	}
	tcpConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		glog.Infoln(err)
		return
	}
	defer tcpConn.Close()
	//客户端注册
	glog.Infoln(os.Args)
	data := msgpk.Pack([]byte(os.Args[2]))
	//glog.Infoln(data)
	n, err := tcpConn.Write(data)
	if err != nil {
		glog.Infoln(n, err)
		return
	}
	setReadTimeout(tcpConn, time.Duration(3)*time.Second)
	buf := make([]byte, 64)
	m, err := tcpConn.Read(buf)
	if err != nil {
		glog.Infoln(err)
		return
	}
	//glog.Infoln(buf[:m])
	register, err := msgpk.Unpack(buf[:m])
	if err != nil {
		glog.Infoln(err)
		return
	}
	if register == "fail" {
		glog.Infoln("register fail")
		return
	} else if register == "ok" {
		//心跳包
		fd, err := setIOBlock(tcpConn)
		if err != nil {
			glog.Infoln(err)
			return
		}
		go loopingCall(fd)
		for {

			buff := make([]byte, 64)
			n, err := fd.Read(buff)
			//glog.Infoln(buff[:n])
			if err == io.EOF {

				glog.Infoln("The RemoteAddr:%s is closed!\n", tcpConn.RemoteAddr().String())
				break
			}
			if err != nil {
				glog.Infoln(err)
				break
			}
			if n > 5 {
				action, err := msgpk.Unpack(buff[:n])
				if err != nil {
					glog.Infoln(err)
					m, err := fd.Write(msgpk.Pack([]byte("fail")))
					if err != nil {
						glog.Infoln(err, m)
						break
					}
					continue
				}
				switch action {
				case "open", "close":
					k, err := fd.Write(msgpk.Pack([]byte(action)))
					glog.Infoln(action)
					if err != nil {
						glog.Infoln(k, err)
						break
					}
				default: //默认是客户端注册
					m, err := fd.Write(msgpk.Pack([]byte("fail")))
					glog.Infoln(action)
					if err != nil {
						glog.Infoln(err, m)
						break
					}
				}
			}
		}
	}

	glog.Flush()
}

//设置读数据超时
func setReadTimeout(conn *net.TCPConn, t time.Duration) {
	conn.SetReadDeadline(time.Now().Add(t))
}

//定时处理
func loopingCall(conn *os.File) {

	pingTicker := time.NewTicker(10 * time.Second) //定时
	for {
		select {
		case <-pingTicker.C:
			//发送心跳
			n, err := conn.Write([]byte("ping"))
			if err != nil {
				glog.Infoln(n, err)
				pingTicker.Stop()
				return
			}
		}
	}
}

//设置阻塞模式
func setIOBlock(conn *net.TCPConn) (fd *os.File, err error) {

	fd, err = conn.File()
	if err != nil {
		glog.Infoln(err)
		fd.Close()
		return nil, err
	}
	conn.Close()
	return fd, nil
}

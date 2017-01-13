package main

import (
	"flag"
	"github.com/golang/glog"
	"global"
	"server"
)

func MainWork() {

	defer func() {
		if err := recover(); err != nil {
			glog.Error("Exception: MainWork", err)
		}
	}()
	Opserver := new(server.Server)
	server.MainWork(Opserver)
	glog.Info("MainWork Stoped ")
	global.Stop()
}

func main() {
	flag.Parse()
	flag.Set("log_dir", "logs")
	flag.Set("alsologtostderr", "true")
	glog.Info("Server start.")
	global.Start()
	for global.IsRunning() {
		go server.AcceptCommand()
		MainWork()
	}
	global.Stop()
	glog.Info("Server stopped.")
	glog.Flush()
}

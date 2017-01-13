package config

import (
	"encoding/xml"

	"github.com/golang/glog"
	"io/ioutil"
	"runtime"
)

//常量定义
const (
	OPServerPath             string = "./conf/Server.xml"
	MaxClientGuard_X64       int32  = 25000 //最大客户端连接数
	MaxClientGuard_X86       int32  = 8000  //最大客户端连接数
	MinFreeMemoryPeriodGuard int32  = 300   //最小自动回收周期
)

type LFreeMemory struct {
	AutoFreeMemory       bool  `xml:"Auto,attr"`
	AutoFreeMemoryPeriod int32 `xml:"AutoFreeMemoryPeriod,attr"`
}

//Server.xml 结构定义

type ServerCfg struct {
	XMLName         xml.Name    `xml:"Server"`
	ListenAddress   string      `xml:"ListenAddress,attr"`
	MaxClientNumber int32       `xml:"MaxClientNumber,attr"`
	InitialMemory   int16       `xml:"InitialMemory,attr"`
	FreeMemory      LFreeMemory `xml:"FreeMemory"`
}

/*
*	加载客户端配置
 */
func LoadServerConfig() (listenAddress string, maxClientNumber int32, initialMemory int16, err error) {

	server_cfg, err := LoadServerConfigStruct()
	listenAddress = server_cfg.ListenAddress
	initialMemory = server_cfg.InitialMemory
	maxClientNumber = server_cfg.MaxClientNumber
	return
}

/*
*	加载服务端配置
 */
func LoadServerConfigStruct() (server_cfg ServerCfg, err error) {
	glog.Infoln("Loading server configure.")
	data, err := ioutil.ReadFile(OPServerPath)
	if err != nil {
		glog.Errorln(err.Error())
		return
	}

	err = xml.Unmarshal(data, &server_cfg)
	if err != nil {
		glog.Errorln("LoadServerConfigStruct->", err.Error())
		return
	}

	//读取配置
	maxClientNumber := server_cfg.MaxClientNumber
	//检查数据区间是否在保护设置之间  小于预设数值，采用预设数据，大于预设数值，采用最大限制数量15000
	if maxClientNumber < 1 {
		maxClientNumber = 1
	}

	if runtime.GOARCH == "amd64" {

		//最大限制设置
		if maxClientNumber > MaxClientGuard_X64 {
			maxClientNumber = MaxClientGuard_X64
		}

	} else if runtime.GOARCH == "386" {

		//最大限制设置
		if maxClientNumber > MaxClientGuard_X86 {
			maxClientNumber = MaxClientGuard_X86
		}
	}

	server_cfg.MaxClientNumber = maxClientNumber

	glog.Infoln("Client number:", maxClientNumber)
	if server_cfg.FreeMemory.AutoFreeMemory {
		if server_cfg.FreeMemory.AutoFreeMemoryPeriod < MinFreeMemoryPeriodGuard {
			server_cfg.FreeMemory.AutoFreeMemoryPeriod = MinFreeMemoryPeriodGuard
		}
	}
	return
}

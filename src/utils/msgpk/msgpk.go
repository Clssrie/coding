package msgpk

import (
	"errors"

	_ "github.com/golang/glog"
)

//协议的格式
// ZY(2bytes) + len(1bytes) + action + ZY(2bytes)

//客户端注册
// ZY(2bytes) + len(1bytes) + 客户端名称 + ZY(2bytes)
//服务器返回
// ZY(2bytes) + len(1bytes) + ok + ZY(2bytes)
// ZY(2bytes) + len(1bytes) + fail + ZY(2bytes)

//客户端发送心跳包
// ZY(2bytes) + len(1bytes) + ping + ZY(2bytes)

//服务器向客户端发送改变状态命令
// ZY(2bytes) + len(1bytes) + open/close + ZY(2bytes)

//客户端返回
// ZY(2bytes) + len(1bytes) + open/close + ZY(2bytes)

func Unpack(buff []byte) (data string, err error) {
	buffLen := len(buff)
	if buffLen <= 5 {
		return "", errors.New("pack length < 5bytes")
	}
	if string(buff[:2]) != "ZY" {
		return "", errors.New("pack head error")
	}
	length := buff[2]
	action := string(buff[3:(3 + length)])

	if string(buff[3+length:]) != "ZY" {
		return "", errors.New("pack tail error")
	}

	return action, nil
}

func Pack(buff []byte) (data []byte) {
	data = make([]byte, 0)
	data = append(data, []byte("ZY")...)
	data = append(data, byte(len(buff)))
	data = append(data, buff...)
	data = append(data, []byte("ZY")...)
	return data
}

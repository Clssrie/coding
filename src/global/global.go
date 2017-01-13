package global

const (
	status_stop   = 0
	status_start  = 1
	status_runing = 2
)

var iExit int32 = 0

func init() {
	iExit = 0
}

func Start() bool {
	iExit = status_start
	return true
}

func Stop() bool {
	iExit = status_stop
	return true
}

func IsRunning() bool {
	return (iExit != status_stop)
}

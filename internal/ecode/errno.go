package ecode

const (
	Success       = 0
	RoutNotExist  = 10001
	ParamsError   = 10002
	InternalError = 10003
)

var codeMap = map[uint32]string{
	Success:       "success",
	RoutNotExist:  "rout not exist",
	ParamsError:   "param error",
	InternalError: "internal error",
}

func GetErrorMessage(code uint32) string {
	msg, ok := codeMap[code]
	if !ok {
		return "unknown"
	}
	return msg
}

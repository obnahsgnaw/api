package apierr

import (
	"strconv"
)

var (
	None              = newCommonErrCode(0, "success")
	InternalError     = newCommonErrCode(1, "internal error")
	ValidateFailed    = newCommonErrCode(2, "invalid arguments")
	ConflictError     = newCommonErrCode(3, "operate conflict")
	AppMidInvalid     = newCommonErrCode(11, "application identify invalid")
	AuthMidInvalid    = newCommonErrCode(12, "authorization token invalid")
	CryptMidDecFailed = newCommonErrCode(13, "data decrypt failed")
	CryptMidEncFailed = newCommonErrCode(14, "data encrypt failed")
	SignMidInvalid    = newCommonErrCode(15, "signature invalid")
	SignMidGenFailed  = newCommonErrCode(16, "signature generate failed")
	PermMidNoPerm     = newCommonErrCode(17, "no permission")
)

// ErrCode 错误码
type ErrCode struct {
	common    bool
	code      uint32
	projectId string
	msgHandle ErrMsgHandler
}

// ErrMsgHandler error message handler
type ErrMsgHandler func(params []interface{}) string

type Factory struct {
	projectId string
}

// New return a new ErrCode factory
func New(projectId int) *Factory {
	return &Factory{projectId: strconv.Itoa(projectId)}
}

// NewErrCode return a new ErrCode
func (f *Factory) NewErrCode(code uint32, msgHandler ErrMsgHandler) ErrCode {
	return ErrCode{
		projectId: f.projectId,
		code:      code,
		msgHandle: msgHandler,
	}
}

// NewStdErrCode return a new ErrCode with no message handler
func (f *Factory) NewStdErrCode(code uint32) ErrCode {
	return ErrCode{
		projectId: f.projectId,
		code:      code,
	}
}

// NewMsgErrCode return a ErrCode with string message
func (f *Factory) NewMsgErrCode(code uint32, msg string) ErrCode {
	return ErrCode{
		projectId: f.projectId,
		code:      code,
		msgHandle: func(params []interface{}) string {
			return msg
		},
	}
}

// NewCommonErrCode return a common ErrCode
func (f *Factory) NewCommonErrCode(code uint32, msg string) ErrCode {
	return newCommonErrCode(code, msg)
}

func newCommonErrCode(code uint32, msg string) ErrCode {
	return ErrCode{
		projectId: "",
		common:    true,
		code:      code,
		msgHandle: func(params []interface{}) string {
			return msg
		},
	}
}

// Code 值
func (c ErrCode) Code() uint32 {
	if c.common || c.projectId == "" {
		return c.code
	}

	v, _ := strconv.Atoi(c.projectId + "0" + strconv.Itoa(int(c.code)))
	return uint32(v)
}

// Message 描述文字
func (c ErrCode) Message(params []interface{}, replaceMsg string) string {
	if replaceMsg != "" {
		return replaceMsg
	}
	if c.msgHandle != nil {
		return c.msgHandle(params)
	}

	return "internal error"
}

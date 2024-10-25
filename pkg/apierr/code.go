package apierr

import (
	"github.com/obnahsgnaw/api/pkg/apierr/errmsg"
	"strconv"
	"strings"
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
	RpcFailed         = newCommonErrCode(18, "rpc call failed")
)

// ErrCode 错误码
type ErrCode struct {
	common        bool              // common类型code不加项目前缀
	code          uint32            // 错误吗
	projectName   string            // 项目名称
	projectId     string            // 项目id
	messageHandle MessageHandler    // 消息处理器
	tmp           map[string]string // 临时数据
}

// ErrMsgHandler error message handler
type ErrMsgHandler func(params []interface{}) string

// MessageHandler error message handler
type MessageHandler func(e ErrCode, params []interface{}) string

func DefaultMessageHandler(msg *errmsg.LocalMessage, e ErrCode, params []interface{}, defaultMsg string) string {
	target := e.Target("target")
	if target == "" {
		target = strconv.Itoa(int(e.Code()))
	} else {
		target = strconv.Itoa(int(e.Code())) + "." + target
	}
	str := msg.Translate(errmsg.Language(e.Local()), e.projectId+"@"+target, params...)
	if str == target && defaultMsg != "" {
		return defaultMsg
	}
	return str
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
	pjn := c.projectName
	if pjn != "" {
		pjn += ": "
	}
	if replaceMsg != "" {
		return pjn + replaceMsg
	}
	if c.messageHandle != nil {
		return pjn + c.messageHandle(c, params)
	}

	return pjn + "internal error"
}

func (c ErrCode) WithTarget(key, val string) ErrCode {
	if c.tmp == nil {
		c.tmp = make(map[string]string)
	}
	c.tmp[key] = val
	return c
}

func (c ErrCode) Target(key string) string {
	if c.tmp == nil {
		return ""
	}
	v, ok := c.tmp[key]
	if ok {
		return v
	}
	return ""
}

func (c ErrCode) WithLocal(local string) ErrCode {
	return c.WithTarget("local", local)
}

func (c ErrCode) Local() string {
	t := c.Target("local")
	if t == "" {
		return "en"
	}
	return t
}

func (c ErrCode) WithProject(id int, name string) ErrCode {
	c1 := &c
	c1.projectId = strconv.Itoa(id)
	c1.projectName = name
	return *c1
}

// ------------------------------------------------------------------------

type Factory struct {
	projectName string
	projectId   string
}

// New return a new ErrCode factory
func New(projectId int) *Factory {
	return &Factory{projectId: strconv.Itoa(projectId)}
}

func (f *Factory) SetProjectName(name string) {
	if name != "" {
		if !strings.HasSuffix(name, ":") {
			name = name + ":"
		}
		f.projectName = name
	}
}

// NewErrCode return a new ErrCode
func (f *Factory) NewErrCode(code uint32, msgHandler ErrMsgHandler) ErrCode {
	return f.NewErrorCode(code, func(e ErrCode, params []interface{}) string {
		return msgHandler(params)
	})
}

// NewErrorCode return a new ErrCode
func (f *Factory) NewErrorCode(code uint32, msgHandler MessageHandler) ErrCode {
	return ErrCode{
		projectName:   f.projectName,
		projectId:     f.projectId,
		code:          code,
		messageHandle: msgHandler,
	}
}

// NewStdErrCode return a new ErrCode with no message handler
func (f *Factory) NewStdErrCode(code uint32) ErrCode {
	return f.NewErrorCode(code, nil)
}

// NewMsgErrCode return a ErrCode with string message
func (f *Factory) NewMsgErrCode(code uint32, msg string) ErrCode {
	return f.NewErrorCode(code, func(e ErrCode, params []interface{}) string {
		return msg
	})
}

// NewCommonErrCode return a common ErrCode
func (f *Factory) NewCommonErrCode(code uint32, msg string) ErrCode {
	return newCommonErrCode(code, msg)
}

func newCommonErrCode(code uint32, msg string) ErrCode {
	return ErrCode{
		projectName: "",
		projectId:   "",
		common:      true,
		code:        code,
		messageHandle: func(e ErrCode, params []interface{}) string {
			return DefaultMessageHandler(ErrMsg, e, params, msg)
		},
	}
}

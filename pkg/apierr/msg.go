package apierr

import "github.com/obnahsgnaw/api/pkg/apierr/errmsg"

var ErrMsg *errmsg.LocalMessage

func init() {
	ErrMsg = errmsg.New()
}

func SetDefaultMsg(message *errmsg.LocalMessage) {
	ErrMsg.Merge(message)
}

func RegisterErrorMessage(message *errmsg.LocalMessage) {
	ErrMsg.Merge(message)
}

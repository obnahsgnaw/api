package apierr

import (
	"log"
	"testing"
)

func TestErrCode(t *testing.T) {
	f := New(1)
	c := f.NewErrorCode(100, func(e ErrCode, params []interface{}) string {
		log.Print(e.Local())
		log.Print(e.Target("target"))
		return "ok"
	})
	c = c.WithLocal("en")
	c = c.WithTarget("target", "forget")
	log.Println(c.Message(nil, ""))
}

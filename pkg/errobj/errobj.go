package errobj

// Provider trans param to service obj
type Provider func(param Param) interface{}

type Param struct {
	Code    uint32
	Message string
	Errors  []Param
}

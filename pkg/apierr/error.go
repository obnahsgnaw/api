package apierr

import (
	"errors"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// ApiError app error
type ApiError struct {
	StatusCode HttpStatus    // http 状态码
	ErrCode    ErrCode       // 错误码
	RawErr     error         // 原始error
	Message    string        // 错误信息 如果为空则使用code中定义的message
	Replace    []interface{} // 字符串的替换值
	Data       interface{}   // 传递一些数据
}

func (e ApiError) Unwrap() error {
	return e.RawErr
}

func (e ApiError) Error() string {
	return e.ErrCode.Message(e.Replace, e.Message)
}

func (e ApiError) SetTarget(k, v string) {
	e.ErrCode = e.ErrCode.WithTarget(k, v)
}

// NewApiErr 创建新的 app error
func NewApiErr(status HttpStatus, code ErrCode, err error) *ApiError {
	var e *ApiError
	if errors.As(err, &e) && e.StatusCode == status && e.ErrCode.code == code.code { // 重复包裹时
		return e
	}
	return &ApiError{
		StatusCode: status,
		ErrCode:    code,
		RawErr:     err,
	}
}

// NewCreated 创建成功
func NewCreated() *ApiError {
	return NewApiErr(StatusCreated, None, nil)
}

// NewDeleted 删除成功
func NewDeleted() *ApiError {
	return NewApiErr(StatusDeleted, None, nil)
}

// NewValidateError 验证错误
func NewValidateError(message string) *ApiError {
	e := NewApiErr(StatusBadRequest, ValidateFailed, nil)
	e.Message = message
	return e
}

// NewTargetValidateError 验证错误
func NewTargetValidateError(target string) *ApiError {
	return NewApiErr(StatusBadRequest, ValidateFailed.WithTarget("target", target), nil)
}

// NewBadRequestError 验证错误
func NewBadRequestError(code ErrCode, err error) *ApiError {
	return NewApiErr(StatusBadRequest, code, err)
}

// NewUnauthorizedError 未认证错误
func NewUnauthorizedError(code ErrCode, err error) *ApiError {
	return NewApiErr(StatusUnauthorized, code, err)
}

// NewForbiddenError 没权限错误
func NewForbiddenError(code ErrCode, err error) *ApiError {
	return NewApiErr(StatusForbidden, code, err)
}

// NewNotFoundError 未找到错误
func NewNotFoundError(code ErrCode, id string) *ApiError {
	e := NewApiErr(StatusNotFound, code, nil)
	e.Replace = []interface{}{id}
	return e
}

// NewConflictError 更新冲突错误
func NewConflictError() *ApiError {
	return NewApiErr(StatusConflict, ConflictError, nil)
}

// NewLockedError 锁定错误
func NewLockedError(code ErrCode) *ApiError {
	return NewApiErr(StatusLocked, code, nil)
}

// NewCommonInternalError 通用内部错误
func NewCommonInternalError(err error) *ApiError {
	return NewApiErr(StatusInternalServerError, InternalError, err)
}

// NewInternalError 内部错误
func NewInternalError(code ErrCode, err error) *ApiError {
	return NewApiErr(StatusInternalServerError, code, err)
}

// ToStatusError 转换成runtime.HTTPStatusError
func ToStatusError(err error) *runtime.HTTPStatusError {
	if err == nil {
		return nil
	}
	var apiErr *ApiError
	if !errors.As(err, &apiErr) {
		apiErr = NewCommonInternalError(err)
	}
	return &runtime.HTTPStatusError{
		HTTPStatus: apiErr.StatusCode.Value(),
		Err:        apiErr,
	}
}

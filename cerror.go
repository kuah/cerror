package cerror

import (
	"bytes"
	"github.com/pkg/errors"
	"io"
	"net/http"
)

/*
  为了可以在接口的handler 可以 trace 到 cause error 的 “错误代码”
  原理 : 这个包, 利用了errors 库的 Cause 方法, 其原理是 在call errors.wrap 时候, 返回的是一个带error() 和 Cause 方法的interface, 而普通不是由errors 包产生的error, 是没有实现Cause 方法的.
具体可以参考errors 包, 递归获取没有实现Cause 方法的error , 以认定为 错误原因. 基于此, 本包再次进行简单的二次封装, 对error 增加error code 的能力

*/

const ErrorCodeNotFound int = -1001

type cdssError struct {
	error
	code int
}

func (c *cdssError) Error() string {
	return c.error.Error()
}
func (c *cdssError) Code() int { return c.code }
func SetCode(err error, code int) error {
	if err == nil {
		return nil
	}
	return &cdssError{
		error: err,
		code:  code,
	}
}
func GetCode(err error) int {
	type coder interface {
		Code() int
	}
	myErr, ok := errors.Cause(err).(coder)
	if ok {
		return myErr.Code()
	}
	return ErrorCodeNotFound
}

func Cause(err error) error {
	return errors.Cause(err)
}
func ErrorWithResp(response *http.Response) error {
	if response == nil {
		return nil
	}
	respBody, _ := io.ReadAll(response.Body)
	defer response.Body.Close()
	response.Body = io.NopCloser(bytes.NewBuffer(respBody))
	responseBodyStr := string(respBody)
	return &cdssError{
		error: errors.New(responseBodyStr),
		code:  response.StatusCode,
	}
}
func New(message string, code int) error {
	return &cdssError{
		error: errors.New(message),
		code:  code,
	}
}

func Wrap(err error, message string, code int) error {
	return errors.Wrap(&cdssError{
		error: err,
		code:  code,
	}, message)
}

func WrapResp(response *http.Response, message string) error {
	return errors.Wrap(ErrorWithResp(response), message)
}

package applicationError

import (
	"fmt"
	"net/http"
	"runtime"
)

type httpError struct {
	Message       string
	Code          int
	ClientMessage string
	Stack         string
}

func New(status int, msg, clientMsg string) *httpError {
	clientMessage := clientMsg
	if clientMessage == "" {
		clientMessage = msg
	}
	err := httpError{
		Message:       msg,
		Code:          status,
		Stack:         fmt.Sprintf("Error: %s", msg),
		ClientMessage: clientMessage,
	}
	return err.stackTrace()
}

func (e httpError) Error() string {
	return e.Message
}

var funcInfoFormat = "Stack Trace: {%s:%d} [%s]"

func getFuncInfo(pc uintptr, file string, line int) string {
	f := runtime.FuncForPC(pc)
	if f == nil {
		return fmt.Sprintf(funcInfoFormat, file, line, "unknwon")
	}
	return fmt.Sprintf(funcInfoFormat, file, line, f.Name())
}

var wrapFormat = "%s\n%s" // "error \n {file:line} [func name] msg"

func (e *httpError) stackTrace() *httpError {
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		e.Stack = fmt.Sprintf(wrapFormat, e.Stack, e.Error())
	}
	e.Stack = fmt.Sprintf(wrapFormat, e.Stack, getFuncInfo(pc, file, line))
	return e
}

func Wrap(err error) error {
	if e, ok := err.(*httpError); ok {
		return e.stackTrace()
	}
	// NOTE: Set status with 500 when error is not application error
	msg := err.Error()
	httpErr := httpError{
		Message: msg,
		Code:    http.StatusInternalServerError,
		Stack:   fmt.Sprintf("Error: %s", msg),
	}

	return httpErr.stackTrace()
}

func UnWrap(err error) *httpError {
	if e, ok := err.(*httpError); ok {
		return e
	}
	// NOTE: Set status with 500 when error is not application error
	msg := err.Error()
	httpErr := httpError{
		Message: msg,
		Code:    http.StatusInternalServerError,
		Stack:   fmt.Sprintf("Error: %s", msg),
	}

	return httpErr.stackTrace()
}

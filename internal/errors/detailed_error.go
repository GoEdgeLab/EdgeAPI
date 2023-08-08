package errors

type DetailedError struct {
	msg  string
	code string
}

func (this *DetailedError) Error() string {
	return this.msg
}

func (this *DetailedError) Code() string {
	return this.code
}

func NewDetailedError(code string, errString string) *DetailedError {
	return &DetailedError{
		msg:  errString,
		code: code,
	}
}

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

func NewDetailedError(code string, error string) *DetailedError {
	return &DetailedError{
		msg:  error,
		code: code,
	}
}

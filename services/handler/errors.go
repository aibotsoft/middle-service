package handler

type SurebetError struct {
	Err         error
	Msg         string
	Permanent   bool
	ServiceName string
}

func (e SurebetError) Error() string {
	if e.Err != nil {
		return e.Msg + ": " + e.Err.Error()
	}
	return e.Msg
}

func (e SurebetError) Unwrap() error { return e.Err }

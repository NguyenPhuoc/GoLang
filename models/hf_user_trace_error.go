package models

type HfUserTraceError struct {
	UserId string
	Data   string
}

func (HfUserTraceError) TableName() string {
	return "hf_user_trace_error"
}
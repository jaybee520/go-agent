package vo

import "google.golang.org/protobuf/reflect/protoreflect"

type LogMessageVO struct {
	System    string `json:"system"`
	ProcessId string `json:"processId"`
	TaskId    string `json:"taskId"`
	TaskName  string `json:"taskName"`
	DetailsId string `json:"detailsId"`
	ExeSeq    string `json:"exeSeq"`
	Msg       string `json:"msg"`
}

func (l LogMessageVO) ProtoReflect() protoreflect.Message {
	//TODO implement me
	panic("implement me")
}

package vo

import "time"

type MessageVO struct {
	System    string `json:"system"`
	ProcessId string `json:"processId"`
	TaskId    string `json:"taskId"`
	TaskName  string `json:"taskName"`
	DetailsId string `json:"detailsId"`
	//InstDetailsId string    `json:"instDetailsId"`
	ExeSeq string        `json:"exeSeq"`
	Time   time.Duration `json:"time"`
	Param  string        `json:"param"`
	Type   string        `json:"type"`
	Key    string        `json:"key"`
}

package vo

type ReturnMessageVO struct {
	System      string `json:"system"`
	ProcessId   string `json:"processId"`
	TaskId      string `json:"taskId"`
	DetailsId   string `json:"detailsId"`
	ExeSeq      string `json:"exeSeq"`
	Type        string `json:"type"`
	Time        string `json:"time"`
	Status      string `json:"status"`
	Drpcode     string `json:"drpcode"`
	Msg         string `json:"msg"`
	OutPutParam string `json:"outPutParam"`
	Key         string `json:"key"`
}

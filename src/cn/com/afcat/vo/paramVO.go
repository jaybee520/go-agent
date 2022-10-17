package vo

type Framework struct {
	ServerType        string
	ExecServer_Passwd string
	ScriptType        string
	ExecuteWay        string
	ExecServer_User   string
	ScriptPro         string
	ExecServer_IP     string
	FileDownloadUrl   string
	JobID             string
}
type Business struct {
	OutputPara []OutputPara
	InputPara  []InputPara
	EnvPara    []EnvPara
}
type OutputPara struct {
	ParamName string `json:"paramName"`
	ParamData string `json:"paramData"`
	ParamType string `json:"paramType"`
	Required  string `json:"required"`
}
type InputPara struct {
	ParamName string `json:"paramName"`
	ParamData string `json:"paramData"`
	ParamType string `json:"paramType"`
	Required  string `json:"required"`
}
type EnvPara struct {
	ParamName string `json:"paramName"`
	ParamData string `json:"paramData"`
	ParamType string `json:"paramType"`
	Required  string `json:"required"`
}
type ParamVO struct {
	Framework Framework `json:"framework"`
	Business  Business  `json:"business"`
	TimeOut   string    `json:"timeOut"`
}

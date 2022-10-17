package task

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/sftp"
	"go-agent/src/cn/com/afcat/mq"
	"go-agent/src/cn/com/afcat/util"
	"go-agent/src/cn/com/afcat/vo"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var logger = log.Default()
var workDir = "/tmp/work/"

var remoteBaseWorkDir = "/tmp/remote_work/"

func TaskExeService(messageVO vo.MessageVO) {
	if "CMD" == messageVO.Type {
		err := os.MkdirAll(workDir, 0777)
		if err != nil {
			fmt.Println("创建工作目录", err)
			panic("工作目录创建失败")
		}
		var paramVO vo.ParamVO
		json.Unmarshal([]byte(messageVO.Param), &paramVO)
		framework := paramVO.Framework
		returnMessageVO := vo.ReturnMessageVO{
			System:    "UYW",
			ProcessId: messageVO.ProcessId,
			TaskId:    messageVO.TaskId,
			DetailsId: messageVO.DetailsId,
			ExeSeq:    messageVO.ExeSeq,
			Type:      "REPLY",
		}
		defer func() {
			if err := recover(); err != nil {
				returnMessageVO.Status = "1"
				returnMessageVO.Msg = fmt.Sprintf("%v", err)
				returnMessageVO.Time = time.Now().Format("2006-01-02 15:04:05")
				mq.SendMq(returnMessageVO)
			}
		}()
		// 下载远程脚本
		downloadRemoteScript(framework)
		remoteWorkDir := remoteBaseWorkDir + paramVO.Framework.JobID + "/"
		// 执行远程脚本
		if "shell" == framework.ScriptType {
			// 生成执行脚本
			localScriptPath, scriptName := generateShellExeScript(paramVO, remoteWorkDir)
			status, paras := ExeCmd(messageVO, paramVO, remoteWorkDir, localScriptPath, remoteWorkDir+scriptName, "sh "+remoteWorkDir+scriptName)
			marshal, _ := json.Marshal(paras)
			returnMessageVO.OutPutParam = string(marshal)
			returnMessageVO.Status = strconv.Itoa(status)
		} else if "python" == framework.ScriptType {
			localScriptPath, scriptName := generatePythonExeScript(paramVO)
			status, paras := ExeCmd(messageVO, paramVO, remoteWorkDir, localScriptPath, remoteWorkDir+scriptName, "python "+remoteWorkDir+"Minions")
			marshal, _ := json.Marshal(paras)
			returnMessageVO.OutPutParam = string(marshal)
			returnMessageVO.Status = strconv.Itoa(status)
		}
		returnMessageVO.Time = time.Now().Format("2006-01-02 15:04:05")
		mq.SendMq(returnMessageVO)
	}
}

func downloadRemoteScript(framework vo.Framework) {
	resp, err := http.Get(framework.FileDownloadUrl)
	if err != nil {
		println("脚本下载失败 error: %+v", err)
		panic("脚本下载失败")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Printf("读取Body失败 error: %+v", err)
		panic("下载的脚本读取失败")
	}
	os.MkdirAll(workDir+framework.JobID, os.ModePerm)
	downloadName := workDir + framework.JobID + string(os.PathSeparator) + "Minions"
	file, err := os.OpenFile(downloadName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		fmt.Println("Open file Failed", err)
		panic("脚本写入本地失败")
	}
	defer func() {
		file.Close()
	}()
	_, err1 := file.Write(body)
	if err1 != nil {
		log.Println("脚本写入失败", err1)
		panic(err)
	}
}

func ExeCmd(messageVO vo.MessageVO, v vo.ParamVO, remoteWorkDir string, localScriptPath string, remoteScriptPath string, exeCmd string) (int, []vo.OutputPara) {
	passwd := util.PasswdDecode(v.Framework.ExecServer_Passwd)
	sshClient, err := util.Connect(v.Framework.ExecServer_User, passwd, v.Framework.ExecServer_IP, 22)
	if err != nil {
		log.Println("连接创建失败", err)
		panic("远程连接创建失败")
	}
	defer sshClient.Close()
	sftpClient := sendScript(v, remoteWorkDir, sshClient, localScriptPath)
	defer func() {
		sftpClient.Remove(remoteWorkDir + "Minions")
		sftpClient.Remove(remoteScriptPath)
		sftpClient.Remove(remoteWorkDir + "returnfile")
		sftpClient.RemoveDirectory(remoteWorkDir)
	}()
	logMessageVO := vo.LogMessageVO{System: messageVO.System, ProcessId: messageVO.ProcessId, TaskId: messageVO.TaskId, TaskName: messageVO.TaskName, DetailsId: messageVO.DetailsId, ExeSeq: messageVO.ExeSeq}
	// 执行脚本
	task, err := util.ExecTask(sshClient, exeCmd, logMessageVO)
	if err != nil {
		log.Println("远程执行脚本失败", err)
		errorMess := fmt.Sprintf("远程执行脚本失败%v", err)
		panic(errorMess)
	}
	if 0 != task {
		log.Println("远程执行脚本失败", err)
		return task, nil
	}
	// 获取返回参数
	outPutParam := readReturnParam(v, remoteWorkDir, sftpClient)
	return 0, outPutParam
}

func readReturnParam(v vo.ParamVO, remoteWorkDir string, sftpClient *sftp.Client) []vo.OutputPara {
	var outPutParam = []vo.OutputPara{}
	util.PullFile(sftpClient, remoteWorkDir+"returnfile", workDir+v.Framework.JobID)
	defer func() {
		os.Remove(workDir + v.Framework.JobID + "/returnfile")
		os.Remove(workDir + v.Framework.JobID)
	}()
	exists, _ := PathExists(workDir + v.Framework.JobID + "/returnfile")
	if exists {
		fileContent, err := ioutil.ReadFile(workDir + v.Framework.JobID + "/returnfile")
		if err != nil {
			log.Println(err)
			panic("获取返回参数失败")
		}
		split := strings.Split(string(fileContent), "\n")
		for _, value := range split {
			if "" != value && strings.Contains(value, "=") {
				param := strings.Split(value, "=")
				para := vo.OutputPara{ParamName: param[0], ParamData: param[1]}
				outPutParam = append(outPutParam, para)
			}
		}
	}
	return outPutParam
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func sendScript(v vo.ParamVO, remoteWorkDir string, sshClient *ssh.Client, scriptPath string) *sftp.Client {
	// 传输脚本
	sftpClient, err := util.GetSftpClient(sshClient)
	if err != nil {
		log.Println("文件传输失败", err)
		panic("文件传输失败")
	}
	localMinName := workDir + v.Framework.JobID + string(os.PathSeparator) + "Minions"
	defer func() {
		os.Remove(localMinName)
		os.Remove(scriptPath)
	}()
	util.PushFile(sftpClient, localMinName, remoteWorkDir)
	util.PushFile(sftpClient, scriptPath, remoteWorkDir)
	return sftpClient
}

func generateShellExeScript(v vo.ParamVO, remoteWorkDir string) (string, string) {
	fileName := v.Framework.JobID + ".sh"
	filePath := workDir + v.Framework.JobID + string(os.PathSeparator) + fileName
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		fmt.Println("Open file Failed", err)
		panic("生成shell执行脚本创建失败")
	}
	defer func() {
		file.Close()
	}()
	file.WriteString("#!/usr/bin/env bash\n")
	for _, value := range v.Business.InputPara {
		file.WriteString(fmt.Sprintf("%s=%s\n", value.ParamName, value.ParamData))
	}
	file.WriteString(fmt.Sprintf(". %sMinions\n", remoteWorkDir))
	for _, value := range v.Business.OutputPara {
		file.WriteString(fmt.Sprintf("echo %s=\"${%s}\" >> %sreturnfile\n", value.ParamName, value.ParamName, remoteWorkDir))
	}
	return filePath, fileName
}

func generatePythonExeScript(v vo.ParamVO) (string, string) {
	fileName := "param.py"
	filePath := workDir + v.Framework.JobID + string(os.PathSeparator) + fileName
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		fmt.Println("Open file Failed", err)
		panic("生成python脚本创建失败")
	}
	defer func() {
		file.Close()
	}()
	file.WriteString("#!/usr/bin/env python\n# -*- coding: utf-8 -*-\nimport os\nRootPath = os.path.dirname(os.path.abspath(__file__))\nSystemSep = os.sep\n")
	file.WriteString("def OutputPara(DictPara):\n    if isinstance(DictPara,dict):\n        for key,value in DictPara.items():\n            with open(RootPath + SystemSep + \"returnfile\",\"a+\") as f:\n                f.write(\"%s=%s\\n\"%(key,value))\n    else:\n        print(\"[ x ] 返回输出参数必须是字典类型\")\n        exit(1)\n")
	param := make(map[string]string, len(v.Business.InputPara))
	for _, value := range v.Business.InputPara {
		//file.WriteString(fmt.Sprintf("%s=%s\n", value.ParamName, value.ParamData))
		param[value.ParamName] = value.ParamData
	}
	file.WriteString("InputPara=")
	marshal, _ := json.Marshal(param)
	file.WriteString(string(marshal))
	return filePath, fileName
}

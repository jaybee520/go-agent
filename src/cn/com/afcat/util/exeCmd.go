package util

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

func ExecCmd() {
	retryTimes := 3
	var retryInterval time.Duration = 3
	user := "afcat"
	host := "10.24.24.11"

	//部分场景下重试登录
	shouldRetry := true
	for i := 1; i <= retryTimes && shouldRetry; i++ {
		//执行命令
		shouldRetry = RunSSHCommand(user, host)
		if !shouldRetry {
			return
		}
		time.Sleep(retryInterval * time.Second)
	}
	if shouldRetry {
		fmt.Println("\n失败,请重试或检查")
	}
}
func shouldRetryByOutput(output string) bool {
	if strings.Contains(output, "错误") { //匹配到"错误"就重试.这里只是Demo，请根据实际情况设置。
		return true
	}
	return false
}
func GetAndFilterOutput(reader *bufio.Reader) (shouldRetry bool) {
	var sumOutput string
	outputBytes := make([]byte, 200)
	for {
		n, err := reader.Read(outputBytes)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(err)
			sumOutput += err.Error()
		}
		output := string(outputBytes[:n])
		fmt.Print("# " + output) //输出屏幕内容
		sumOutput += output
		if shouldRetryByOutput(output) {
			shouldRetry = true
		}
	}
	if shouldRetryByOutput(sumOutput) {
		shouldRetry = true
	}
	return
}
func RunSSHCommand(user, host string) (shouldRetry bool) {
	//获取执行命令
	cmd := exec.Command("ssh", fmt.Sprintf("%s@%s", user, host))
	cmd.Stdin = os.Stdin

	var wg sync.WaitGroup
	wg.Add(2)
	//捕获标准输出
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}
	readout := bufio.NewReader(stdout)
	go func() {
		defer wg.Done()
		shouldRetryTemp := GetAndFilterOutput(readout)
		if shouldRetryTemp {
			shouldRetry = true
		}
	}()

	//捕获标准错误
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	readerr := bufio.NewReader(stderr)
	go func() {
		defer wg.Done()
		shouldRetryTemp := GetAndFilterOutput(readerr)
		if shouldRetryTemp {
			shouldRetry = true
		}
	}()

	//执行命令
	cmd.Run()
	wg.Wait()
	return
}

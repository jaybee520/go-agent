package util

import (
	"bufio"
	"fmt"
	"github.com/pkg/sftp"
	"go-agent/src/cn/com/afcat/logService"
	"go-agent/src/cn/com/afcat/vo"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"path/filepath"
	"time"
)

func Connect(user, password, host string, port int) (*ssh.Client, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		sshClient    *ssh.Client
		err          error
	)
	// get auth method
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(password))

	clientConfig = &ssh.ClientConfig{
		User:    user,
		Auth:    auth,
		Timeout: 30 * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	// connet to ssh
	addr = fmt.Sprintf("%s:%d", host, port)

	if sshClient, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}
	return sshClient, nil
}

func GetSftpClient(sshClient *ssh.Client) (*sftp.Client, error) {
	var (
		sftpClient *sftp.Client
		err        error
	)
	// create sftp client
	if sftpClient, err = sftp.NewClient(sshClient); err != nil {
		return nil, err
	}

	return sftpClient, nil
}

func PushFile(sftpClient *sftp.Client, localFilePath string, remoteDir string) {
	srcFile, err := os.Open(localFilePath)
	if err != nil {
		log.Println(err)
		panic(err)
	}
	defer srcFile.Close()

	var remoteFileName = filepath.Base(localFilePath)
	err = sftpClient.MkdirAll(remoteDir)
	if err != nil {
		log.Println(err)
		panic(err)
	}
	dstFile, err := sftpClient.Create(path.Join(remoteDir, remoteFileName))
	if err != nil {
		log.Println(err)
		panic(err)
	}
	defer dstFile.Close()

	ff, err := ioutil.ReadAll(srcFile)
	if err != nil {
		fmt.Println("ReadAll error : ", localFilePath)
		log.Fatal(err)
	}
	dstFile.Write(ff)

	log.Println("文件远程传输完成!")
}

func PushDirectory(sftpClient *sftp.Client, localPath string, remotePath string) {
	localFiles, err := ioutil.ReadDir(localPath)
	if err != nil {
		log.Fatal("read dir list fail ", err)
	}
	for _, backupDir := range localFiles {
		localFilePath := path.Join(localPath, backupDir.Name())
		remoteFilePath := path.Join(remotePath, backupDir.Name())
		if backupDir.IsDir() {
			sftpClient.Mkdir(remoteFilePath)
			PushDirectory(sftpClient, localFilePath, remoteFilePath)
		} else {
			PushFile(sftpClient, path.Join(localPath, backupDir.Name()), remotePath)
		}
	}
	fmt.Println(localPath + " copy directory to remote server finished!")
}

func PullFile(sftpClient *sftp.Client, remoteFilePath string, localDir string) {
	if _, err := sftpClient.Stat(remoteFilePath); os.IsNotExist(err) {
		log.Println("文件不存在")
		return
	}

	srcFile, err := sftpClient.Open(remoteFilePath)
	if err != nil {
		log.Println(err)
		panic(err)
	}
	defer srcFile.Close()

	var localFileName = path.Base(remoteFilePath)
	dstFile, err := os.Create(path.Join(localDir, localFileName))
	if err != nil {
		log.Println(err)
		panic(err)
	}
	defer dstFile.Close()

	if _, err = srcFile.WriteTo(dstFile); err != nil {
		log.Println(err)
		panic(err)
	}

	log.Println("远程文件拉取完成!")
}

// 远程执行脚本
func ExecTask(client *ssh.Client, remoteCmd string, vo vo.LogMessageVO) (int, error) {
	session, err := client.NewSession()
	if err != nil {
		fmt.Println("创建会话失败", err)
		panic(err)
	}
	defer session.Close()

	cmdReader, err := session.StdoutPipe()
	if err != nil {
		log.Println(err)
		panic(err)
	}
	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			vo.Msg = scanner.Text() + "\n"
			//conf.SendLog(vo)
			logService.SendLog(vo)
			log.Println(scanner.Text())
		}
	}()
	errPipe, err := session.StderrPipe()
	if err != nil {
		log.Println(err)
		panic(err)
	}
	errScanner := bufio.NewScanner(errPipe)
	go func() {
		for errScanner.Scan() {
			text := errScanner.Text()
			vo.Msg = text + "\n"
			logService.SendLog(vo)
			log.Println(text)
		}
	}()

	err1 := session.Run(remoteCmd)
	if err1 != nil {
		fmt.Println("远程执行脚本失败", err1)
		return 1, err1
	} else {
		fmt.Println("远程执行脚本成功")
		return 0, nil
	}
}

/*func main() {
	sshHost := "home.xxx.cn"
	sshUser := "x"
	sshPassword := "xxxxxx"
	sshType := "password" //password 或者 key
	sshKeyPath := ""      //ssh id_rsa.id 路径"
	sshPort := 22

	//创建sshp登陆配置
	config := &ssh.ClientConfig{
		Timeout:         time.Second, //ssh 连接time out 时间一秒钟, 如果ssh验证错误 会在一秒内返回
		User:            sshUser,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //这个可以， 但是不够安全
		//HostKeyCallback: hostKeyCallBackFunc(h.Host),
	}
	if sshType == "password" {
		config.Auth = []ssh.AuthMethod{ssh.Password(sshPassword)}
	} else {
		config.Auth = []ssh.AuthMethod{publicKeyAuthFunc(sshKeyPath)}
	}

	//dial 获取ssh client
	addr := fmt.Sprintf("%s:%d", sshHost, sshPort)
	sshClient, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		log.Fatal("创建ssh client 失败", err)
	}
	defer sshClient.Close()

	//创建ssh-session
	session, err := sshClient.NewSession()
	if err != nil {
		log.Fatal("创建ssh session 失败", err)
	}
	defer session.Close()
	//执行远程命令
	combo, err := session.CombinedOutput("whoami; cd /; ls -al;echo https://github.com/dejavuzhou/felix")
	if err != nil {
		log.Fatal("远程执行cmd 失败", err)
	}
	log.Println("命令输出:", string(combo))

}

func publicKeyAuthFunc(kPath string) ssh.AuthMethod {
	keyPath, err := homedir.Expand(kPath)
	if err != nil {
		log.Fatal("find key's home dir failed", err)
	}
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		log.Fatal("ssh key file read failed", err)
	}
	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatal("ssh key signer failed", err)
	}
	return ssh.PublicKeys(signer)
}
*/

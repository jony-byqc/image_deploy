package ssh

import (
	"bytes"
	"fmt"
	"github.com/jony-byqc/image_deploy/pkg"
	"github.com/jony-byqc/image_deploy/utils"
	"github.com/pkg/sftp"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path"
	"time"
)

type CliSsh struct {
	user       string
	pwd        string
	ip         string
	port       int
	sshType    string
	sshClient  *ssh.Client
	sftpClient *sftp.Client
}

func NewSSHClient(user, pwd, ip string, port int) *CliSsh {
	return &CliSsh{
		user:    user,
		pwd:     pwd,
		ip:      ip,
		port:    port,
		sshType: pkg.PasswordType,
	}
}

// 不使用 HostKey， 使用密码
func (c *CliSsh) getConfig_nokey() *ssh.ClientConfig {
	config := &ssh.ClientConfig{
		User: c.user,
		Auth: []ssh.AuthMethod{
			ssh.Password(c.pwd),
		},
		Timeout:         30 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return config
}
func (c *CliSsh) Connect() error {
	// 创建ssh登陆配置
	config := &ssh.ClientConfig{
		Timeout: time.Second, //ssh 连接timeout时间一秒钟，如果ssh验证错误 会在1秒内返回
		User:    c.user,      //指定ssh连接用户
		// HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	if c.sshType == pkg.PasswordType {
		config.Auth = []ssh.AuthMethod{ssh.Password(c.pwd)}
	}

	// dial获取ssh Client
	addr := fmt.Sprintf("%s:%d", c.ip, c.port)
	sshClient, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		log.Fatal("创建ssh client 失败", err)
	}
	c.sshClient = sshClient
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return fmt.Errorf("new sftp client error: %w", err)
	}
	c.sftpClient = sftpClient
	return nil
}

func (c *CliSsh) Close() error {
	if c.sshClient != nil {
		if err := c.sftpClient.Close(); err != nil {
			return err
		}
	}
	if c.sftpClient != nil {
		if err := c.sftpClient.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (c *CliSsh) check() error {
	if c.sshClient == nil {
		if err := c.Connect(); err != nil {
			return err
		}
	}
	return nil
}

func (c *CliSsh) Run(cmd string) (string, error) {
	if err := c.check(); err != nil {
		return "", err
	}

	session, err := c.sshClient.NewSession()
	if err != nil {
		return "", fmt.Errorf("create new session error: %w", err)
	}
	defer session.Close()
	buf, err := session.CombinedOutput(cmd)
	return string(buf), err
}

func (c *CliSsh) DownloadFile(remoteFile, localPath string) error {
	source, err := c.sftpClient.Open(remoteFile)
	if err != nil {
		return fmt.Errorf("sftp client open file error: %w", err)
	}
	defer source.Close()
	localFile := path.Join(localPath, path.Base(remoteFile))
	os.MkdirAll(localPath, os.ModePerm)
	target, err := os.OpenFile(localFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("open local file error: %w", err)
	}
	defer target.Close()
	_, err = io.Copy(target, source)
	if err != nil {
		return fmt.Errorf("write file error: %w", err)
	}
	return nil
}

func (c *CliSsh) DownloadFileOrDir(remotePath, localPath string) error {
	if c.sshClient == nil {
		if err := c.Connect(); err != nil {
			return err
		}
	}
	// 如是文件直接下载
	if utils.IsFile(remotePath) {
		return c.DownloadFile(remotePath, localPath)
	}
	// 如是目录，递归下载
	remoteFiles, err := c.sftpClient.ReadDir(remotePath)
	if err != nil {
		return fmt.Errorf("read path failed: %w", err)
	}
	for _, item := range remoteFiles {
		remoteFilePath := path.Join(remotePath, item.Name())
		localFilePath := path.Join(localPath, item.Name())
		if item.IsDir() {
			err = os.MkdirAll(localFilePath, os.ModePerm)
			if err != nil {
				return err
			}
			err = c.DownloadFileOrDir(remoteFilePath, localFilePath) // 递归本函数
			if err != nil {
				return err
			}
		} else {
			err = c.DownloadFile(path.Join(remotePath, item.Name()), localPath)
			if err != nil {
				return err
			}
		}
	}
	return nil

}

func (c *CliSsh) UploadFile(localFile, remotePath string, sameLevel bool) error {
	file, err := os.Open(localFile)
	if nil != err {
		return fmt.Errorf("open local file failed: %w", err)
	}
	defer file.Close()
	var remoteFileName = ""
	if !sameLevel {
		remoteFileName = path.Base(localFile)
		c.sftpClient.MkdirAll(remotePath)
	}
	ftpFile, err := c.sftpClient.Create(path.Join(remotePath, remoteFileName))
	if nil != err {
		return fmt.Errorf("Create remote path failed: %w", err)
	}
	defer ftpFile.Close()
	fileByte, err := ioutil.ReadAll(file)
	if nil != err {
		return fmt.Errorf("read local file failed: %w", err)
	}
	ftpFile.Write(fileByte)
	return nil
}

func (c *CliSsh) UploadFileOrDir(localPath, remotePath string) error {
	if c.sshClient == nil {
		if err := c.Connect(); err != nil {
			return err
		}
	}
	// 如是文件直接上传
	if utils.IsFile(localPath) {
		return c.UploadFile(localPath, remotePath, true)
	}
	// 如是目录，递归上传
	localFiles, err := ioutil.ReadDir(localPath)
	if err != nil {
		return fmt.Errorf("read path failed: %w", err)
	}
	for _, item := range localFiles {
		localFilePath := path.Join(localPath, item.Name())
		remoteFilePath := path.Join(remotePath, item.Name())
		if item.IsDir() {
			err = c.sftpClient.MkdirAll(remoteFilePath)
			if err != nil {
				return err
			}
			err = c.UploadFileOrDir(localFilePath, remoteFilePath) // 递归本函数
			if err != nil {
				return err
			}
		} else {
			err = c.UploadFile(path.Join(localPath, item.Name()), remotePath, false)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *CliSsh) SessionRequestPty() (*ssh.Session, error) {

	session, err := c.sshClient.NewSession()
	if err != nil {
		return nil, err
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		return nil, err
	}
	return session, nil
}

func (c *CliSsh) RunCmdS(cmds []string) error {
	session, err := c.SessionRequestPty()
	if err != nil {
		return err
	}
	if session == nil {
		return fmt.Errorf("ssh session is nil")
	}
	defer session.Close()

	stdinBuf, err := session.StdinPipe()
	if err != nil {
		return err
	}

	var outbt, errbt bytes.Buffer
	session.Stdout = &outbt

	session.Stderr = &errbt
	err = session.Shell()
	if err != nil {
		return err
	}
	for _, cmd := range cmds {
		cmd = cmd + "\n"
		stdinBuf.Write([]byte(cmd))
	}

	log.Info("执行命令完毕", outbt.String()+errbt.String())
	return nil
}

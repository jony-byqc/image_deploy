package server

import (
	"github.com/jony-byqc/image_deploy/internal/config"
	"github.com/jony-byqc/image_deploy/pkg"
	"github.com/jony-byqc/image_deploy/pkg/ssh"
	"github.com/jony-byqc/image_deploy/utils"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

type Client struct {
}

func NewClient() *Client {
	return &Client{}
}

func (s *Client) Open() {
	deploymentConfig, err := config.ParseConfig()
	if err != nil {
		log.Error("获取配置文件错误", err)
		return
	}
	s.Run(deploymentConfig)
}

func (s *Client) Close() {

}

func (s *Client) Run(deploymentConfig config.Deployment) {

	switch deploymentConfig.Type {
	case pkg.PullOnly:
	case pkg.DeployOnly:
	case pkg.PullAndDeploy:
		s.PullAndDeployManager(deploymentConfig)
	default:
		log.Error("未识别的部署类型:", deploymentConfig.Type)
	}
}

func (s *Client) PullAndDeployManager(deploymentConfig config.Deployment) {

	for i := range deploymentConfig.SSHConfig {
		s.PullAndDeploy(deploymentConfig.SSHConfig[i], deploymentConfig.Deploy.TargetFolder)
	}

}

func (s *Client) PullAndDeploy(sshConfig config.SSHConfig, remotePath string) {
	CliSsh := ssh.NewSSHClient(sshConfig.User, sshConfig.Password, sshConfig.Ip, sshConfig.Port)
	CliSsh.Connect()
	localDirFiles, _ := ioutil.ReadDir("./")
	for _, f := range localDirFiles {
		//log.Println(f.Name())
		_, err := CliSsh.UploadFileOrDir(f.Name(), utils.StitchingString(remotePath, "/", f.Name()))
		if err != nil {
			log.Errorf(f.Name(), "upload failed", err)
		}
	}
	CliSsh.Close()
}

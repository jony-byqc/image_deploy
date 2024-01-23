package server

import (
	"fmt"
	"github.com/jony-byqc/image_deploy/internal/config"
	"github.com/jony-byqc/image_deploy/pkg"
	"github.com/jony-byqc/image_deploy/pkg/image"
	"github.com/jony-byqc/image_deploy/pkg/ssh"
	"github.com/jony-byqc/image_deploy/utils"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"path"
	"strings"
)

type Client struct {
	composePath   []string
	loadImageTars []string
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
		if err := s.PullImage(); err != nil {
			log.Error(err)
		}
	case pkg.DeployOnly:
		s.DeployManager(deploymentConfig)
	case pkg.PullAndDeploy:
		if err := s.PullImage(); err != nil {
			log.Error(err)
			return
		}
		s.DeployManager(deploymentConfig)
	default:
		log.Error("未识别的部署类型:", deploymentConfig.Type)
	}
}

func (s *Client) PullImage() error {
	images, composePaths := utils.FindCompose()
	s.composePath = utils.FormatPath(composePaths)

	pullerConfig := image.Config{
		NeedBar:  true,
		UseCache: false,
	}
	cliPuller := image.NewCliPuller(&pullerConfig)
	if len(images) < 1 {
		return fmt.Errorf("当前目录下未找到image镜像")
	}
	for _, v := range images {
		tag, err := cliPuller.ParseImageTag(v)
		if err != nil {
			return fmt.Errorf("error when parse image uri: %s\n", v)
		}
		dockerTarName := strings.ReplaceAll(tag.Repo, "/", "_") + "_" + tag.Img + "_" + tag.Tag + ".tar"
		s.loadImageTars = append(s.loadImageTars, dockerTarName)
		cliPuller.DownloadDockerImage(tag, "", "")
	}

	return nil
}

func (s *Client) DeployManager(deploymentConfig config.Deployment) {
	_, composePaths := utils.FindCompose()
	s.composePath = utils.FormatPath(composePaths)
	for i := range deploymentConfig.SSHConfig {
		s.Deploy(deploymentConfig.SSHConfig[i], deploymentConfig.Deploy.TargetFolder)
	}

}

func (s *Client) Deploy(sshConfig config.SSHConfig, targetDir string) {
	CliSsh := ssh.NewSSHClient(sshConfig.User, sshConfig.Password, sshConfig.Ip, sshConfig.Port)
	CliSsh.Connect()
	localDirFiles, _ := ioutil.ReadDir("./")
	for _, f := range localDirFiles {
		err := CliSsh.UploadFileOrDir(f.Name(), utils.StitchingString(targetDir, "/", f.Name()))
		if err != nil {
			log.Errorf(f.Name(), "upload failed", err)
		}
	}

	var (
		composeNum = len(s.composePath)
		imagesNum  = len(s.loadImageTars)
		loadCmds   = make([]string, len(s.loadImageTars))
		upCmds     = make([]string, composeNum)
	)

	if composeNum <= 0 {
		log.Error("无 需部署的compose文件")
		return
	}

	for i := 0; i < imagesNum; i++ {
		tarPath := path.Join(targetDir, "/", s.loadImageTars[i])
		loadCmds[i] = utils.StitchingString(pkg.Docker, pkg.Load, pkg.I, tarPath)
	}

	for i := 0; i < composeNum; i++ {
		composePath := path.Join(targetDir, s.composePath[i])
		upCmds[i] = utils.StitchingString(pkg.DockerCompose, pkg.F, composePath, pkg.Upd)
	}

	cmds := append(loadCmds, upCmds...)

	err := CliSsh.RunCmdS(cmds)
	if err != nil {
		log.Error("deploy exec cmds failed", err)
		return
	}
	CliSsh.Close()
}

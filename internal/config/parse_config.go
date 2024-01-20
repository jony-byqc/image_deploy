package config

import (
	_ "embed"
	"github.com/jony-byqc/image_deploy/utils"
	log "github.com/sirupsen/logrus"
	pEnv "github.com/traefik/paerser/env"
	pFlag "github.com/traefik/paerser/flag"
	"gopkg.in/yaml.v3"
	"io/fs"
	"io/ioutil"
	"os"
)

//go:embed deployment.yaml.example
var embedConfig []byte

// ParseConfig 解析部署配置文件优先：命令行参数 > 环境变量 > 当前目录配文件
func ParseConfig() (config Deployment, err error) {
	var (
		data []byte
	)
	config = NewConfig()
	defaultNamePrefix := "DEPLOY_"

	path := "deployment.yaml"

	// 从本地读取 读不到从二进制静态文件包中读取
	if ok, _ := utils.PathExists(path); ok {
		data, err = ioutil.ReadFile(path)
		if err != nil {
			return
		}
	} else {
		// config 写入当前目录
		_ = ioutil.WriteFile(path, embedConfig, fs.FileMode(0755))
		data = embedConfig
	}

	log.Println("Using configuration at:", path)

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return
	}

	err = pEnv.Decode(os.Environ(), defaultNamePrefix, &config)
	if err != nil {
		return
	}
	err = pFlag.Decode(os.Args[1:], &config)
	if err != nil {
		log.Println(err)
		return
	}

	return
}

package config

type Deployment struct {
	DockerUser     string      `yaml:"user"`
	DockerPassword string      `yaml:"password"`
	Type           string      `yaml:"type"`
	SSHConfig      []SSHConfig `yaml:"ssh_config"`
	Deploy         Deploy      `yaml:"deploy"`
}

type SSHConfig struct {
	Ip       string `yaml:"ip"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type Deploy struct {
	TargetFolder string `yaml:"target_folder"`
}

func NewConfig() Deployment {
	return Deployment{
		DockerUser:     "",
		DockerPassword: "",
		Type:           "",
		SSHConfig: []SSHConfig{
			{
				Ip:       "192.168.60.157",
				Port:     22,
				User:     "wjq",
				Password: "123456",
			},
		},
		Deploy: Deploy{
			TargetFolder: "/compose",
		},
	}

}

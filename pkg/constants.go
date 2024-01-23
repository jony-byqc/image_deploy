package pkg

const (
	PasswordType  = "password"
	PullOnly      = "pull"
	DeployOnly    = "deploy"
	PullAndDeploy = "pullAndDeploy"
)

const (
	DockerCompose = "docker-compose "
	Docker        = "docker "
	Load          = "load "
	I             = "-i "
	F             = "-f "
	Upd           = " up -d"
	Down          = " down"
)
const (
	DefaultImageTag      = "latest"
	DefaultImageRegistry = "registry-1.docker.io"
	DefaultImageRepo     = "library"
	DefaultEmptyJson     = `{
	"created": "1970-01-01T00:00:00Z",
	"container_config": {
		"Hostname": "",
		"Domainname": "",
		"User": "",
		"AttachStdin": false,
		"AttachStdout": false,
		"AttachStderr": false,
		"Tty": false,
		"OpenStdin": false,
		"StdinOnce": false,
		"Env": null,
		"Cmd": null,
		"Image": "",
		"Volumes": null,
		"WorkingDir": "",
		"Entrypoint": null,
		"OnBuild": null,
		"Labels": null
	}
}`
)

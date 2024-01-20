package image

type Config struct {
	//Proxy    string
	NeedBar  bool
	UseCache bool
}

type ImageTag struct {
	ImagUri    string
	Img        string
	Tag        string
	Registry   string
	Repo       string
	Repository string
	AuthUrl    string
	RegService string
	RepoTags   string
}

type ImageContext struct {
	Config   string   `json:"Config"`
	RepoTags []string `json:"RepoTags"`
	Layers   []string `json:"Layers"`
}

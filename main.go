package main

import "github.com/jony-byqc/image_deploy/internal/server"

func main() {
	serverRun := server.NewClient()
	serverRun.Open()
	//is, ps := utils.FindCompose()
	//fmt.Println(is, ps)
}

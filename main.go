package main

import (
	"git-audit/cmd"
	"git-audit/config"
)

var (
	APPName    = "git-audit"
	Maintainer = "ops@xxxx.io"
	APPVersion = "v0.2"
	BuildTime  = "200601021504"
	GitCommit  = "ccccccccccccccc"
)

func main() {
	cmd.Run()
}

func init() {
	config.APPName = APPName
	config.Maintainer = Maintainer
	config.APPVersion = APPVersion
	config.BuildTime = BuildTime
	config.GitCommit = GitCommit
}

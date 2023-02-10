package service

import (
	"encoding/json"
	"time"

	"git-audit/config"

	"github.com/hpcloud/tail"
)

// ssh 方式拉取代码的报警服务
func SSHService() {
	tails := NewTailFile(config.Config.GitAudit.SSHFile)
	var (
		line *tail.Line
		ok   bool
	)

	for {
		line, ok = <-tails.Lines //遍历chan，读取日志内容
		if !ok {
			LogSugar.Infof("tail file close reopen, filename:%s\n", tails.Filename)
			time.Sleep(time.Second)
			continue
		}
		LogSugar.Infof("line: %v", line.Text)
		var info Model

		err := json.Unmarshal([]byte(line.Text), &info)
		if err != nil {
			LogSugar.Infof("json: %v", err)
			return
		}

		LogSugar.Infof("info struct: %v", info)
		GitAuditAction(config.DB, info)

	}

}

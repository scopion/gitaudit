package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"git-audit/config"

	"github.com/hpcloud/tail"
)

func HTTPService() {
	tails := NewTailFile(config.Config.GitAudit.HTTPFile)
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
		var infoInt ModleInt

		err := json.Unmarshal([]byte(line.Text), &infoInt)
		if err != nil {
			LogSugar.Infof("json: %v", err)
			return
		}
		info := Model{
			Command:       infoInt.Command,
			CorrelationId: infoInt.CorrelationId,
			GlProjectPath: infoInt.GlProjectPath,
			RemoteIp:      infoInt.RemoteIp,
			Time:          infoInt.Time,
			UserName:      infoInt.UserName,
		}
		info.UserId = fmt.Sprintf("user-%v", infoInt.UserId)
		info.Command = strings.ReplaceAll(info.Command, "_", "-")
		infoInt.GlKeyType = "http"

		LogSugar.Infof("info struct: %v", info)
		GitAuditAction(config.DB, info)

	}

}

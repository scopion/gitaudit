package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"git-audit/config"

	"github.com/hpcloud/tail"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var LogSugar *zap.SugaredLogger

type Model struct {
	Command       string `json:"command"`
	CorrelationId string `json:"correlation_id"`
	GlKeyType     string `json:"gl_key_type"`
	GlProjectPath string `json:"gl_project_path"`
	RemoteIp      string `json:"remote_ip"`
	Time          string `json:"time"`
	UserId        string `json:"user_id"`
	UserName      string `json:"username"`
}

// userid 是int
type ModleInt struct {
	Command       string `json:"action"`
	CorrelationId string `json:"correlation_id"`
	GlKeyType     string `json:"gl_key_type"`
	GlProjectPath string `json:"path"`
	RemoteIp      string `json:"remote_ip"`
	Time          string `json:"time"`
	UserId        int    `json:"user_id"`
	UserName      string `json:"username"`
}

type Alter struct {
	Level    int    `json:"level"`
	Nownum   int    `json:"nownum"`
	UserId   string `json:"user_id"`
	UserName string `json:"username"`
	Status   string `json:"alterstatus"`
}

func isNotNull(content string) bool {
	for _, s := range content {
		if string(s) != "" {
			return true
		}
	}
	return false
}

type FeishuChatbot struct {
	webhook string
	secret  string
}

func (f *FeishuChatbot) sendText(msg string) error {
	ts := f.getTS()
	sign, _ := GenSign(f.secret, ts)
	data := map[string]interface{}{"msg_type": "text", "timestamp": ts, "sign": sign}

	if isNotNull(msg) {
		data["content"] = map[string]string{"text": msg}
	} else {
		fmt.Printf("text类型：%v", data)
		log.Printf("text类型，消息内容不能为空! ")
		return errors.New("text类型，消息内容不能为空!")
	}
	result, err := SendJson(f.webhook, "/", "POST", data)
	if err != nil {
		return err
	}
	log.Printf("机器人通知response: %v", string(result))
	return nil
}

func (f *FeishuChatbot) getTS() int64 {
	timestamp := time.Now().Unix()
	return timestamp
}

func NewFeishuCbot() *FeishuChatbot {
	return &FeishuChatbot{
		config.Config.GitAudit.Webhook,
		config.Config.GitAudit.Secret,
	}

}

func GenSign(secret string, timestamp int64) (string, error) {
	//timestamp + key 做sha256, 再进行base64 encode
	stringToSign := fmt.Sprintf("%v", timestamp) + "\n" + secret
	var data []byte
	h := hmac.New(sha256.New, []byte(stringToSign))
	_, err := h.Write(data)
	if err != nil {
		return "", err
	}
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signature, nil
}

func SendJson(url, path, method string, data interface{}) ([]byte, error) {
	url = url + path
	jsonStr, _ := json.Marshal(data)
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonStr))
	if err != nil {
		panic(err)
	}
	req.Header.Add("content-type", "application/json")
	defer req.Body.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	return result, nil
}

func NewTailFile(file string) (tails *tail.Tail) {
	tailConfig := tail.Config{
		ReOpen:    true,                                 // 重新打开
		Follow:    true,                                 // 是否跟随
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2}, // 从文件的哪个地方开始读
		MustExist: false,                                // 文件不存在不报错
		Poll:      true,
	}
	tails, err := tail.TailFile(file, tailConfig)
	if err != nil {
		LogSugar.Errorf("tail file close reopen, filename:%s\n", tails.Filename)
	}
	return
}

func UpdateDB(db *gorm.DB, m Model) {
	sql := "insert into git(command,correlation_id,gl_key_type,gl_project_path,remote_ip,time,user_id,username,updatetime) " +
		"values(?, ?, ?, ?, ?, ?, ?, ?, ?)  " +
		"on duplicate key update command=VALUES(command),gl_key_type=VALUES(gl_key_type)," +
		"gl_project_path=VALUES(gl_project_path),remote_ip=VALUES(remote_ip),time=VALUES(time)," +
		"user_id=VALUES(user_id),username=VALUES(username),updatetime=VALUES(updatetime)"
	tx := db.Begin()

	result := tx.Exec(sql, m.Command, m.CorrelationId, m.GlKeyType, m.GlProjectPath, m.RemoteIp, m.Time, m.UserId, m.UserName, time.Now())
	if result.Error != nil {
		LogSugar.Errorf("updatedb 记录入库失败, 用户： %v 失败原因：%v", m.UserId, result.Error)
		tx.Rollback()
	}

	tx.Commit()
	LogSugar.Infof("updatedb 记录入库成功, 用户： %v 用户名：%v", m.UserId, m.UserName)
}

func GetAlter(db *gorm.DB, m Model) (alterinfo Alter) {
	userId := m.UserId

	sql := "select * from gitalter where user_id = ? "
	result := db.Raw(sql, userId).Scan(&alterinfo)
	if result.Error != nil {
		LogSugar.Errorf("get alert failed: %v", result.Error)
	}

	return
}

func checkNum(db *gorm.DB, m Model) {
	username := m.UserName
	userId := m.UserId
	type GlCount struct {
		GlProjectPath string `json:"gl_project_path"`
		Count         int    `json:"count"`
		UserId        string `json:"user_id"`
	}
	var glCount GlCount
	LogSugar.Infof("user_id: %v", userId)

	// SELECT count(distinct gl_project_path) as count  from git WHERE user_id = 'user-94' and command = 'git-upload-pack' and DATE_SUB(CURDATE(), INTERVAL 7 DAY) <= DATE(TIME)
	sql := "SELECT count(distinct gl_project_path) as count from git WHERE user_id = ? and command = 'git-upload-pack' and DATE_SUB(CURDATE(), INTERVAL 7 DAY) <= DATE(TIME)"
	result := db.Raw(sql, userId).Scan(&glCount)
	if result.Error != nil {
		LogSugar.Errorf(" checknum select failed: %v", result.Error)
	}
	nownum := glCount.Count

	sql2 := "insert into gitalter(user_id,username,nownum,updatetime) " +
		"values(?, ?, ?, ?)  " +
		"on duplicate key update username=VALUES(username),nownum=VALUES(nownum),updatetime=VALUES(updatetime)"
	tx := db.Begin()
	result = tx.Exec(sql2, userId, username, nownum, time.Now())
	if result.Error != nil {
		LogSugar.Errorf("alert 记录入库失败, 用户： %v 失败原因：%v", m.UserId, result.Error)
		tx.Rollback()
	}
	tx.Commit()
	LogSugar.Infof("alert 记录入库成功, 用户： %v 用户名：%v", userId, username)

}

func GitAuditAction(db *gorm.DB, info Model) {

	feishu := NewFeishuCbot()
	if info.Command == "git-upload-pack" {
		UpdateDB(config.DB, info)
		checkNum(config.DB, info)
		alterresult := GetAlter(config.DB, info)
		num := alterresult.Nownum
		level := alterresult.Level
		if num >= level {
			msg := fmt.Sprintf("用户名: %v, ID: %v \n本周内拉取仓库数量为 [ %v ], 预设条件为 [ %v ], 请及时处理",
				info.UserName,
				info.UserId,
				alterresult.Nownum,
				alterresult.Level,
			)
			LogSugar.Infof("飞书消息: %v", msg)
			err := feishu.sendText(fmt.Sprintf("%v%v", msg, config.Config.GitAudit.AT))
			if err != nil {
				LogSugar.Error("error: %v", err)
			}

		}
	} else {
		LogSugar.Info("非拉取操作")
	}

}

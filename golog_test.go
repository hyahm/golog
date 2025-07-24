package golog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/fatih/color"
)

type MpMessageList struct {
	ID              int64      `xorm:"'id'"`
	TypeID          int64      `xorm:"'type_id'"`
	Title           string     `xorm:"'title'"`
	Summary         string     `xorm:"'summary'"`
	UpdateAt        *time.Time `xorm:"'update_at'"`
	Platform        int        `xorm:"'platform'"`
	Scope           int        `xorm:"'scope'"`
	CreateTime      time.Time  `xorm:"'create_at'"`
	ContentType     int        `xorm:"'content_type'"`
	Content         string     `xorm:"'content'"`
	AuditTime       time.Time  `xorm:"'audit_time'"`
	AuditStatus     int        `xorm:"'audit_status'"`
	PushTime        time.Time  `xorm:"'push_time'"`
	ClickVolume     int        `xorm:"'click_volume'"`
	SendTime        time.Time  `xorm:"'send_time'"`
	SendStatus      int        `xorm:"'send_status'"`
	SendCount       int        `xorm:"'send_count'"`
	ReceptionVolume int        `xorm:"'reception_volume'"`
	SendType        int        `xorm:"'send_type'"`
	UID             int64      `xorm:"'uid'"`
	Deleted         bool       `xorm:"'deleted'"`
	MsgID           string     `xorm:"'msg_id'"`
}

func TestInitLogger(t *testing.T) {
	defer Sync()

	ShowBasePath = true
	DefaultUnit = Hour
	WarnHandler = func(ctime, hostname, line, msg string, label map[string]string) {
		cfg := elasticsearch.Config{
			Addresses: []string{
				"https://es.hyahm.com",
			},
			Username: "elastic",
			Password: "OVIGr-sdoTIdfcaLVTHD",
		}

		es, err := elasticsearch.NewClient(cfg)
		if err != nil {
			log.Fatalf("Error creating the client: %s", err)
		}

		type Doc struct {
			Message  string `json:"message"`
			Time     string `json:"time"`
			Hostname string `json:"hostname"`
			Level    string `json:"level"`
		}
		// 创建一个文档
		doc := Doc{
			Message:  msg,
			Time:     ctime,
			Hostname: hostname,
			Level:    "error",
		}
		b, _ := json.Marshal(doc)
		res, err := es.Index("log", bytes.NewReader(b)) // 索引文档

		if err != nil {
			log.Fatalf("Error indexing document: %s", err)
		}

		// 执行创建索引请求

		fmt.Println(res)
	}
	Warn("aaaaaa")
	// golog.InitLogger("log/a.log", 1024, false, 10)
	a := NewLog("log/a.log", 1024, false, 10)
	a.Debugf("foo", "aaaa", "bb")
	a.Warn(color.New(color.BgYellow).Sprint("aaaa"), color.New(color.BgBlue).Sprint("bbbb"))
	Level = DEBUG
	// test()
	a.Error("bar")

}

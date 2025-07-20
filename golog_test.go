package golog

import (
	"testing"
	"time"

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
	ErrorHandler = func(ctime, hostname, line, msg string, label map[string]string) {
		t.Log("你是怎么做到的")
	}
	Error("aaaaaa")
	// golog.InitLogger("log/a.log", 1024, false, 10)
	a := NewLog("log/a.log", 1024, true, 10)
	for range 100 {
		a.Info("foo", "aaaa", "bb")
	}
	a.Warn(color.New(color.BgYellow).Sprint("aaaa"), color.New(color.BgBlue).Sprint("bbbb"))
	Level = DEBUG
	// test()
	a.Error("bar")

}

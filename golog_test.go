package golog

import (
	"testing"

	// "github.com/elastic/go-elasticsearch/v8"
	// "github.com/elastic/go-elasticsearch"
	"github.com/fatih/color"
)

func TestInitLogger(t *testing.T) {
	defer Sync()

	ShowBasePath = true
	DefaultUnit = Hour
	WarnHandler = func(ctime, hostname, line, msg string, label map[string]string) {
		// cfg := elasticsearch.Config{
		// 	Addresses: []string{
		// 		"https://es.hyahm.com",
		// 	},
		// 	Username: "elastic",
		// 	Password: "OVIGr-sdoTIdfcaLVTHD",
		// }

		// es, err := elasticsearch.NewClient(cfg)
		// if err != nil {
		// 	log.Fatalf("Error creating the client: %s", err)
		// }

		// type Doc struct {
		// 	Message  string `json:"message"`
		// 	Time     string `json:"time"`
		// 	Hostname string `json:"hostname"`
		// 	Level    string `json:"level"`
		// }
		// // 创建一个文档
		// doc := Doc{
		// 	Message:  msg,
		// 	Time:     ctime,
		// 	Hostname: hostname,
		// 	Level:    "error",
		// }
		// b, _ := json.Marshal(doc)
		// res, err := es.Index("log", bytes.NewReader(b)) // 索引文档

		// if err != nil {
		// 	log.Fatalf("Error indexing document: %s", err)
		// }

		// // 执行创建索引请求

		// fmt.Println(res)
	}
	Warn("警告")
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

package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"53it.net/zql" // http://git.53it.net/zuoxiupeng/zql or https://github.com/shiguanghuxian/zql
	"gopkg.in/mgo.v2"
)

// Data Ajax相应数据根结构
type Data struct {
	State string      `json:"state"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
}

func main() {
	// 系统日志显示文件和行号
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	rootDir := flag.String("d", "./", "The root directory of the website")
	port := flag.String("p", "80", "Web port")
	ipAddress := flag.String("a", "0.0.0.0", "Site IP address")
	help := flag.Bool("h", false, "Use the help")
	flag.Parse()
	// 帮助
	if *help {
		flag.Usage()
		os.Exit(1)
	}
	// 静态路径处理
	staticPath := GetRootDir() + "/view/" // 静态文件路径
	if *rootDir != "" {
		staticPath = *rootDir
	}
	// 启动http服务
	StartHTTP(*ipAddress, *port, staticPath)
}

// StartHTTP 启动http服务
func StartHTTP(ipAddr, port, staticPath string) {
	// 防止空参数
	if ipAddr == "" {
		ipAddr = "0.0.0.0"
	}
	if port == "" {
		port = "80"
	}
	if staticPath == "" {
		staticPath = "./"
	}
	// 端口监听ip和端口
	listenAddress := ipAddr + ":" + port // 启动ip和端口
	// 静态文件服务器
	http.Handle("/", http.FileServer(http.Dir(staticPath)))
	// ajax请求数据表数据
	http.HandleFunc("/convert", AjaxConvert)
	// 启动http服务
	log.Println("正在启动http服务:", listenAddress)
	err := http.ListenAndServe(listenAddress, nil)
	if err != nil {
		log.Println("http服务启动失败:", err)
		os.Exit(1)
	}
	log.Println("启动http服成功")
}

// AjaxConvert 处理ajax请求
func AjaxConvert(w http.ResponseWriter, r *http.Request) {
	// 接收参数
	valS := r.URL.Query()
	cType := valS.Get("type")
	zqlStr := valS.Get("zql")
	// 最终发送的数据包
	data := &Data{State: "1", Msg: "服务端错误"}
	// 判断是否未传参数
	if cType == "" {
		data.Msg = "参数'type'不能为空"
	} else if zqlStr == "" {
		data.Msg = "参数'zql'不能为空"
	} else {
		// 表前缀
		prefix := valS.Get("prefix")
		// 创建zql对象
		zqlObj, err := zql.New(prefix, zqlStr)
		if err != nil {
			data.Msg = err.Error()
		} else {
			data.State = "0" // 返回状态置为0
			data.Msg = "zql转换成功"
			switch cType {
			case "mongodb":
				mgoDb := new(mgo.Session).DB("dbname")
				mgoStr, err := zqlObj.GetMongoQueryStr(mgoDb, "")
				if err != nil {
					data.State = "1"
					data.Msg = err.Error()
				} else {
					data.Data = mgoStr
				}
				break
			case "influxdb":
				infStr, err := zqlObj.GetInfluxdbQuery("")
				if err != nil {
					data.State = "1"
					data.Msg = err.Error()
				} else {
					data.Data = infStr
				}
				break
			case "elasticsearch":
				esStr, err := zqlObj.GetElasticQueryStr()
				if err != nil {
					data.State = "1"
					data.Msg = err.Error()
				} else {
					data.Data = esStr
				}
				break
			default:
				data.State = "1"
				data.Msg = "参数'type'无法识别，请检查是否错误"
				break
			}
		}
	}
	// 转json
	wdata, err := json.Marshal(data)
	if err != nil {
		log.Println("ajax输出转json失败:" + err.Error())
	}
	// 发送数据
	w.Write(wdata)
}

// GetRootDir 获取程序跟目录
func GetRootDir() string {
	// 获取当前根目录
	file, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "."
	}
	return file
}

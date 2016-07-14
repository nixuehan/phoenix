//author 逆雪寒
//version 0.9.1
package main

import (
	"flag"
	"log"
	"fmt"
	"net/http"
	"runtime"
	"bytes"
	"strconv"
	"time"
	"html/template"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
)

var (
	Host  string
	Port  string
	Mongodb string
)

func init() {
	log.SetFlags(log.LstdFlags)
	flag.StringVar(&Host, "host", "localhost", "bound ip. default:localhost")
	flag.StringVar(&Port, "port", "8884", "port. default:8884")
	flag.StringVar(&Mongodb, "cs", "127.0.0.1:27017", "Mongodb Connection String. example:mongodb://db1.example.net,db2.example.net:2500/?replicaSet=test&connectTimeoutMS=300000")
	flag.Parse()

}

type Models struct {
	mongo *mgo.Session
}

func toInt(v string) int {
	if s, err := strconv.Atoi(v); err == nil {
		return s
	}else{
		return 0
	}
}

func NewModels() (*Models,error){
	mongo, err := mgo.Dial(Mongodb)  //连接数据库
	if err != nil {
		log.Println("fail:mongodb connection failed")
		return nil,err
	}
	mongo.SetMode(mgo.Primary, true)

	return &Models{mongo},nil
}

const (
	PhoenixDB = "Phoenix"
)

type ApiMessage struct{
	ExecutionTime interface{}
	Year interface{}
	Month interface{}
	Day interface{}
	Milli interface{}
	Path interface{}
}

func(this *Models) apiLogByYMD(groupName string,year int,month int,day int,path string) string{
	db := this.mongo.DB(PhoenixDB)
	collection := db.C(groupName)

	result := &ApiMessage{}

	iter := collection.Find(bson.M{"year":year,"month":month,"day":day,"path":path}).Iter()

	buf := bytes.NewBufferString("")
  	for iter.Next(&result) {
  		buf.WriteString(fmt.Sprintf("[%d,%.2f],",result.Milli,result.ExecutionTime))
    }
    return buf.String()
}

func(this *Models) apiLogByYM(groupName string,year int,month int,path string) string{
	db := this.mongo.DB(PhoenixDB)
	collection := db.C(groupName)

	result := &ApiMessage{}

	iter := collection.Find(bson.M{"year":year,"month":month,"path":path}).Iter()

	buf := bytes.NewBufferString("")
  	for iter.Next(&result) {
  		buf.WriteString(fmt.Sprintf("[%d,%.2f],",result.Milli,result.ExecutionTime))
    }
    return buf.String()
}

func(this *Models) apiSlowByYM(groupName string,year int,month int) string{
	db := this.mongo.DB(PhoenixDB)
	collection := db.C(groupName + "Slowlog")

	result := &ApiMessage{}

	iter := collection.Find(bson.M{"year":year,"month":month}).Sort("-executiontime").Iter()

	buf := bytes.NewBufferString("")
  	for iter.Next(&result) {
  		buf.WriteString(fmt.Sprintf("(%.2f)<a href='/api?year=%d&month=%d&day=%d&groupName=%s&path=%s'>%s</a><br/>",result.ExecutionTime,result.Year,result.Month,result.Day,groupName,result.Path,result.Path))
    }
    return buf.String()
}

func(this *Models) apiSlowByYMD(groupName string,year int,month int,day int) string{
	db := this.mongo.DB(PhoenixDB)
	collection := db.C(groupName + "Slowlog")

	result := &ApiMessage{}

	iter := collection.Find(bson.M{"year":year,"month":month,"day":day}).Sort("-executiontime").Iter()

	buf := bytes.NewBufferString("")
  	for iter.Next(&result) {
  		buf.WriteString(fmt.Sprintf("(%.2f)<a href='/api?year=%d&month=%d&day=%d&groupName=%s&path=%s'>%s</a><br/>",result.ExecutionTime,result.Year,result.Month,result.Day,groupName,result.Path,result.Path))
    }
    return buf.String()
}

func(this *Models) Close(){
	this.mongo.Close()
}

const (
	MonitorHeader = `<!doctype html>
			<html lang="en">
			<head>
			  <title>phoenix api性能监控系统</title>
			  <script type="text/javascript" src="http://cdn.hcharts.cn/jquery/jquery-1.8.3.min.js"></script>
			  <script type="text/javascript" src="http://cdn.hcharts.cn/highcharts/highcharts.js"></script>
			  <script type="text/javascript" src="http://cdn.hcharts.cn/highcharts/exporting.js"></script>
			  <script>`

	MonitorFooter = `</script>
			</head>
			<body>
			  <div id="container" style="min-width:700px;height:400px"></div>
			</body>
			</html>`
)

const (
	Header = `<!doctype html>
			<html lang="en">
			<head><title>phoenix api性能监控系统</title></head>
			<body><div id="container" style="min-width:700px;height:400px">`

	Footer = `
			</div>
			</body>
			</html>`
)


func monitor(res http.ResponseWriter, req *http.Request) {
	t := template.New("phoenix")

	body := `
		<a href="/slow?groupName=phoenixBee">bee项目</a>
	`

	outPutBuf := bytes.NewBufferString(Header)
	outPutBuf.WriteString(body)
	outPutBuf.WriteString(Footer)

	t.Parse(outPutBuf.String())
	t.Execute(res,nil)
}

func slow(res http.ResponseWriter, req *http.Request) {
	model,err := NewModels()
	if err != nil{
		return
	}
	defer model.Close()

	req.ParseForm()

	groupName := req.Form["groupName"]
	if len(groupName) == 0{
		http.Error(res,"groupName can not be empty",444)
		return
	}

	year := req.Form["year"]
	month := req.Form["month"]
	day := req.Form["day"]

	startN := time.Now()
	y,m,_ := startN.Date()
	mm := int(m)

	if len(year) != 0 {
		y = toInt(year[0])
	}

	if len(month) != 0 {
		mm = toInt(month[0])
	}

	var apiData string
	if len(day) != 0 {
		apiData = model.apiSlowByYMD(groupName[0],y,mm,toInt(day[0]))
	}else {
		apiData = model.apiSlowByYM(groupName[0],y,mm)
	}

	t := template.New("phoenix")
	outPutBuf := bytes.NewBufferString(Header)
	outPutBuf.WriteString(apiData)
	outPutBuf.WriteString(Footer)

	t.Parse(outPutBuf.String())
	t.Execute(res,nil)
}

func api(res http.ResponseWriter, req *http.Request) {
	model,err := NewModels()
	if err != nil{
		return
	}
	defer model.Close()

	req.ParseForm()

	groupName := req.Form["groupName"]
	year := req.Form["year"]
	month := req.Form["month"]
	day := req.Form["day"]
	path := req.Form["path"]

	if len(year) == 0 || len(month) == 0 || len(path) == 0 || len(groupName) == 0{
		http.Error(res,"year and month and path and groupName can not be empty",444)
		return
	}

	var apiData,maxZoom string
	if len(day) != 0 {
		apiData = model.apiLogByYMD(groupName[0],toInt(year[0]),toInt(month[0]),toInt(day[0]),path[0])
	}else {
		apiData = model.apiLogByYM(groupName[0],toInt(year[0]),toInt(month[0]),path[0])
		maxZoom = "minRange: 14 * 24 * 3600000,"
	}


	body := `$(function () {
    $('#container').highcharts({
        title: {
            text: '`+path[0]+`',
            x: -20 //center
        },
        subtitle: {
            text: 'Source: github.com',
            x: -20
        },
        xAxis: {
            type: 'datetime',
            `+maxZoom+`
        },
        yAxis: {
            title: {
                text: '执行时间 (秒)'
            },
            plotLines: [{
                value: 0,
                width: 1,
                color: '#808080'
            }]
        },
        tooltip: {                                                              
            formatter: function() {                                             
                    return '<b>'+ this.series.name +'</b><br/>'+                
                    Highcharts.dateFormat('%Y-%m-%d %H:%M:%S', this.x) +'<br/>'+
                    '耗费秒数:' + Highcharts.numberFormat(this.y, 2);                         
            }                                                                   
        },
        legend: {
            layout: 'vertical',
            align: 'right',
            verticalAlign: 'middle',
            borderWidth: 0
        },
        series: [{
            name: 'api',
            data: [`+apiData+`]
        }]
    });
});`

	t := template.New("phoenix")
	outPutBuf := bytes.NewBufferString(MonitorHeader)
	outPutBuf.WriteString(body)
	outPutBuf.WriteString(MonitorFooter)

	t.Parse(outPutBuf.String())
	t.Execute(res,nil)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	http.HandleFunc("/",monitor)
	http.HandleFunc("/api",api)
	http.HandleFunc("/slow",slow)
	log.Println("Success:Phoenix admin on port 8884")
	http.ListenAndServe(Host + ":" + Port, nil)
}

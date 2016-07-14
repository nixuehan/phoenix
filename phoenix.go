//author 逆雪寒
//version 0.9.1
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
	"runtime"
	"math"
	"strconv"
    "gopkg.in/mgo.v2"
)

var (
	Host  string
	Port  string
	Mongodb string
)

var (
	Queue chan map[string]interface{}
)

func init() {
	log.SetFlags(log.LstdFlags)
	flag.StringVar(&Host, "host", "localhost", "bound ip. default:localhost")
	flag.StringVar(&Port, "port", "8888", "port. default:8888")
	flag.StringVar(&Mongodb, "cs", "127.0.0.1:27017", "Mongodb Connection String. example:mongodb://db1.example.net,db2.example.net:2500/?replicaSet=test&connectTimeoutMS=300000")
	flag.Parse()

	Queue = make(chan map[string]interface{},5000)
}
  
func toFloat(v string) float64 {
	if s, err := strconv.ParseFloat(v, 32); err == nil {
		return s
	}else{
		return 0
	}
}

type Models struct {
	mongo *mgo.Session
}

func NewModels() (*Models,error){
	mongo, err := mgo.Dial(Mongodb)
	if err != nil {
		return nil,err
	}
	mongo.SetMode(mgo.Primary, true)

	return &Models{mongo},nil
}


const (
	PhoenixDB = "Phoenix"
	SlowQueryTime = 0.6 * 100
)

type ApiMessageCollections struct{
	ExecutionTime interface{}
	Year interface{}
	Month interface{}
	Day interface{}
	Milli interface{}
	Path interface{}
}

func(this *Models) ApiCollecte(api ApiMessageCollections,groupName string) error{
	db := this.mongo.DB(PhoenixDB)
	collection := db.C(groupName)
	return collection.Insert(&api)
}

func(this *Models) ApiSlowLog(api ApiMessageCollections,groupName string) error{
	db := this.mongo.DB(PhoenixDB)
	collection := db.C(groupName + "Slowlog")
	return collection.Insert(&api)
}

func(this *Models) Close(){
	this.mongo.Close()
}

func Shopping(path string,et float64,groupName string) {
	startN := time.Now()
	year,month,day := startN.Date()

	food := make(map[string]interface{})
	food["executionTime"] = math.Trunc(et*1e2)*1e-2
	food["year"] = year
	food["month"] = int(month)
	food["day"] = day
	food["milli"] = startN.Unix() * 1000
	food["path"] = path
	food["groupName"] = groupName

	Queue <- food
}

func Cooking(food map[string]interface{}) error{
	model,err := NewModels()

	if err != nil{
		return fmt.Errorf("fail:mongodb connection failed")
	}
	defer model.Close()

	field := ApiMessageCollections{food["executionTime"],food["year"],food["month"],food["day"],food["milli"],food["path"]}

	executionTime := food["executionTime"].(float64) * 100
	if executionTime > SlowQueryTime {
		model.ApiSlowLog(field,food["groupName"].(string))
	}
	return model.ApiCollecte(field,food["groupName"].(string))
}

func Monitor() {
	go func(){
		for{
			food := <- Queue
			if err := Cooking(food);err != nil{
				log.Printf("error:%v",err)
			}
		}
	}()
}

type WaitForYou struct{}

func (this *WaitForYou) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	req.ParseForm()

	ac := req.URL.Path

	if ac == "/dota" {
		title := req.Form["title"]
		timeUsed := req.Form["timeUsed"]
		groupName := req.Form["groupName"]

		if len(title) == 0 || len(timeUsed) == 0 || len(groupName) == 0{
			res.Write([]byte("NO"))
			return
		}

  		Shopping(title[0],toFloat(timeUsed[0]),groupName[0])

		res.Write([]byte("OK"))

	}else if ac == "/ping" {
		res.Write([]byte("pong"))
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	Monitor()

	s := &http.Server{
		Addr:           Host + ":" + Port,
		Handler:        &WaitForYou{},
		ReadTimeout:    31 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("Success:Phoenix has been started")
	log.Fatal(s.ListenAndServe())

}

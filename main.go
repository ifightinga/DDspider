package main

import (
	"context"
	"fmt"
	"github.com/hajimehoshi/oto"
	"github.com/robfig/cron/v3"
	"github.com/tosone/minimp3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

var mgo  *mongo.Client
var err  error
var player *oto.Player


type BaiduApiAudio struct {
	DataByte   []byte `json:"data_byte"`
	InsertTime int64  `json:"insert_time"`
	IsPlay     bool   `json:"is_play"`
	DataString string `json:"data_string"`
}

var BaiduChan = make(chan BaiduApiAudio, 30)

func InitMongo()  {
	clientOption := options.Client().ApplyURI("mongodb://admin:admin@123.206.118.99:27017")
	mgo, err = mongo.Connect(context.TODO(), clientOption)
	if err != nil {
		panic("initialize mgo failed")
	}
	if err2 := mgo.Ping(context.TODO(), nil); err2 != nil {
		fmt.Println(err2)
	}
}

func GetMongoDataJson() []BaiduApiAudio {
	var res =  []BaiduApiAudio{}
	filter := bson.M{
		"isplay":false,
	}
	c ,_ := mgo.Database("wechat").Collection("baiduApiAudio").Find(context.TODO(), filter, nil)
	c.All(context.TODO(), &res)
	c.Close(context.Background())
	return res
}

func DoPlayAudio(b BaiduApiAudio)  {
	fmt.Println(time.Unix(b.InsertTime,0), "信息===>：", b.DataString)
		_, databyte,_ := minimp3.DecodeFull(b.DataByte)
		for i:=0; i<3; i++{
			player.Write(databyte)
			time.Sleep(3 * time.Second)
		}
		mgo.Database("wechat").Collection("baiduApiAudio").
			UpdateOne(context.TODO(), bson.M{"databyte":b.DataByte}, bson.M{"$set":bson.M{"isplay":true}})
}

func InChanData()  {
	reslist := GetMongoDataJson()
	for _,v := range reslist{
		BaiduChan <- v
	}
}

func OutChanData(c <-chan BaiduApiAudio)  {
	contextPlayer, _ := oto.NewContext(16000, 1, 2, 2048)
	player = contextPlayer.NewPlayer()
	for vv:= range c{
		DoPlayAudio(vv)
		//fmt.Println("播放：", vv.DataString)
		//time.Sleep(5* time.Second)
	}
	player.Close()
	contextPlayer.Close()
}

func main()  {
	fmt.Println("======程序已启动，请务关闭程序，若漏听信息，可于本界面查看文字版信息======\nDesignBy => Fighting")
	InitMongo()
	go OutChanData(BaiduChan)

	c2 := cron.New(cron.WithSeconds())
	c2.AddFunc("*/60 * * * * ?", InChanData)
	c2.Start()
	defer c2.Stop()
	select {

	}

}
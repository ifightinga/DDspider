package main

import (
	"encoding/json"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

var (
	Latitude        string
	Longitude       string
	StationId       string
	ApiVersion      string
	AppClientId     string
	CategoryId      string
	InsertArryTitle []string
)

type TargetRequest struct {
	CategoriesUrl       string
	CategoriesDetailUrl string
	ProductDetailUrl    string
}

//定义一下请求返回值的struct，便于使用json反序列解析
type UrlResponseURL struct {
	Success    bool            `json:"success"`
	Code       int             `json:"code"`
	Msg        string          `json:"msg"`
	Data       DataResponseURL `json:"data"`
	ServerTime int             `json:"server_time"`
}
type DataResponseURL struct {
	Cate []CateResponseURL `json:"cate"`
}
type CateResponseURL struct {
	Id               string `json:"id"`
	Name             string `json:"name"`
	CategoryImageUrl string `json:"category_image_url"`
}
type UrlResponseGOOD struct {
	Success    bool             `json:"success"`
	Code       int              `json:"code"`
	Msg        string           `json:"msg"`
	Data       DataResponseGOOD `json:"data"`
	ServerTime int              `json:"server_time"`
}
type DataResponseGOOD struct {
	AdId                 string             `json:"ad_id"`
	CategoryName         string             `json:"category_name"`
	ServiceSupportsCache string             `json:"service_supports_cache"`
	Cate                 []CateResponseGOOD `json:"cate"`
	CacheLife_time       int                `json:"cache_life_time"`
}
type CateResponseGOOD struct {
	Id       string
	Name     string
	ImageUrl string
	Products []GoodsResponse
}
type GoodsResponse struct {
	Id            string `json:"id"`
	ProductName   string `json:"product_name"`
	OriginPrice   string `json:"origin_price"`
	Price         string `json:"price"`
	VipPrice      string `json:"vip_price"`
	Spec          string `json:"spec"`
	SmallImage    string `json:"small_image"`
	TotalSales    int    `json:"total_sales"`
	MonthSales    int    `json:"month_sales"`
	Status        int    `json:"status"`
	NetWeight     int    `json:"net_weight"`
	NetWeightUnit string `json:"net_weight_unit"`
	Oid           int    `json:"oid"`
	StockNumber   int    `json:"stock_number"`
}

//结构体参数初始化，发起HTTP请求
func (t *TargetRequest) init_spider() *TargetRequest {
	t.CategoriesUrl = "https://maicai.api.ddxq.mobi/homeApi/categories"
	t.CategoriesDetailUrl = "https://maicai.api.ddxq.mobi/homeApi/categoriesdetail"
	t.ProductDetailUrl = "https://maicai.api.ddxq.mobi/productApi/productDetail"
	return t
}

func (t *TargetRequest) spider_req() (*http.Request, *[]CateResponseURL) {
	fmt.Printf("请输入您的经度：")
	fmt.Scanln(&Longitude)
	fmt.Printf("请输入您的维度：")
	fmt.Scanln(&Latitude)
	fmt.Printf("请输入需要爬取的站点ID：")
	fmt.Scanln(&StationId)
	if len(Longitude) <=2 || len(Latitude) <= 1 || len(StationId) <=6 {
		fmt.Printf("【警告】请正确的输入：Longitude，Latitude，StationId等爬取信息！\n【==========Design By Fighting==========】")
		time.Sleep(10*time.Second)
		panic("输入的经纬度及坐标无效！")
	}
	ApiVersion = "8.9.0"
	AppClientId = "3"
	CategoryId = "5a69bc6a936edf9b3f8bf82a"
	client := http.Client{}
	req, _ := http.NewRequest("GET", t.CategoriesUrl, nil)
	req.Header.Add("user-agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1")
	q := req.URL.Query()
	q.Add("app_client_id", AppClientId)
	q.Add("api_version", ApiVersion)
	q.Add("category_id", CategoryId)
	q.Add("latitude", Latitude)
	q.Add("longitude", Longitude)
	q.Add("station_id", StationId)
	//参数设置完成之后，一定要记得Encode
	req.URL.RawQuery = q.Encode()
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("请求错误，错误：%v\n", err.Error())
	}
	resbyte, _ := ioutil.ReadAll(res.Body)
	var datares UrlResponseURL
	jsonerr := json.Unmarshal(resbyte, &datares)
	if jsonerr != nil {
		fmt.Printf("Json解析错误，错误：%v", jsonerr)
	}
	cateresponse := datares.Data.Cate
	return req, &cateresponse
}

func (t *TargetRequest) save_cates(req *http.Request, cates *[]CateResponseURL) {
	client := http.Client{}
	req2, _ := http.NewRequest("GET", t.CategoriesDetailUrl, nil)
	InsertArryTitle = append(InsertArryTitle, "CategoryName", "CategoryName2", "Id", "ProductName", "OriginPrice",
		"Price", "VipPrice", "Spec", "SmallImage", "TotalSales", "MonthSales", "Status", "NetWeight", "NetWeightUnit",
		"Oid", "StockNumber")
	//设置计数器
	count := 1
	//EXCEL操作：如下是excel操作，初始化日期信息，打开excel，并写入第一条数据：title数据！
	nowdate := time.Now().Format("2006-01-02")
	f := excelize.NewFile()
	f.SetSheetRow("sheet1", "A"+strconv.Itoa(count), &InsertArryTitle)
	fmt.Printf("第%d行d数据：<<标题信息>>写入excel成功！\n",count)
	for _, v := range *cates {
		q2 := req.URL.Query()
		q2.Set("category_id", v.Id)
		req2.URL.RawQuery = q2.Encode()
		res, err := client.Do(req2)
		if err != nil {
			fmt.Printf("请求错误：%v", err.Error())
		}
		resbyte, _ := ioutil.ReadAll(res.Body)
		//fmt.Println(string(resbyte))
		var datares UrlResponseGOOD
		jsonerr := json.Unmarshal(resbyte, &datares)
		if jsonerr != nil {
			fmt.Printf("JSON解析错误：%v", jsonerr)
		}
		//第二次循环，循环二级分类
		for _, v := range datares.Data.Cate {
			//第三次循环，循环商品信息
			for _, vv := range v.Products {
				count++
				insertdata := []interface{}{datares.Data.CategoryName,v.Name, vv.Id, vv.ProductName, vv.OriginPrice, vv.Price, vv.VipPrice, vv.Spec,
					vv.SmallImage, vv.TotalSales, vv.MonthSales, vv.Status, vv.NetWeight, vv.NetWeightUnit}
				//EXCEL操作：在循环中插入每一条核心数据！
				f.SetSheetRow("sheet1","A"+strconv.Itoa(count), &insertdata)
				fmt.Printf("第%d数据：%v-%v-%v写入excel成功！\n", count, datares.Data.CategoryName,v.Name, vv.ProductName)
				println("==============================================================================", insertdata)
			}
		}
	}
	//EXCEL操作：整个循环结束之后，保存excel，并进行错误处理！
	if err := f.SaveAs("./" + nowdate + "：叮咚调研.xlsx"); err != nil {
		fmt.Println(err)
	}
}



func main() {
	// 睡一下
	//time.Sleep(2 * time.Second)
	//初始化实例对象
	a := TargetRequest{}
	//初始化爬虫的url信息
	a.init_spider()
	//第一次爬取获取基础信息，并设置请求信息
	req, cateresponse := a.spider_req()
	//存储基本信息，并进行第二次爬取
	a.save_cates(req, cateresponse)
}

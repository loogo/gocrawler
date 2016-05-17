package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/loogo/gocrawler/database"
	_ "github.com/mattn/go-sqlite3"
)

func crawler(url string, c chan jdProduct) int {
	doc, err := goquery.NewDocument(url)

	if err != nil {
		log.Fatal(err)
	}
	root := doc.Find("#plist .gl-item")
	root.Each(func(i int, s *goquery.Selection) {
		go func() {
			gsku, _ := s.Find(".gl-i-wrap.j-sku-item").Attr("data-sku")
			skudoc, _ := goquery.NewDocument(fmt.Sprintf("https://item.jd.com/%s.html", gsku))

			pname := s.Find(".p-name").Text()
			pimg, exist := skudoc.Find("#spec-n1 img").Attr("src")

			pprice := getprice(gsku)
			if !exist {
				pimg = "Not Exist!"
			}
			p := pprice[0]["p"]
			price, _ := strconv.ParseFloat(p, 64)
			c <- jdProduct{name: pname, img: pimg, price: price}
		}()
	})
	return root.Length()
}
func getprice(sku string) (price []map[string]string) {
	url := fmt.Sprintf("http://p.3.cn/prices/mgets?skuIds=J_%s&type=1", sku)
	response, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal(body, &price)
	if err != nil {
		fmt.Println(err)
	}
	return
}

func main() {
	database.CreateDb()
	now := time.Now()
	url := "https://list.jd.com/list.html?cat=9987,653,655"
	c := make(chan jdProduct)
	length := crawler(url, c)
	var buffer bytes.Buffer
	buffer.WriteString(`
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="utf-8">
			<meta http-equiv="X-UA-Compatible" content="IE=edge">
			<meta name="viewport" content="width=device-width, initial-scale=1">
			<title>SUI 模板</title>
			<link rel="stylesheet" href="http://g.alicdn.com/sui/sui3/0.0.18/css/sui.min.css">
		</head>
		<body>
	`)
	buffer.WriteString("<table class=\"table\"\n")
	for i := 0; i < length; i++ {
		data := <-c
		buffer.WriteString(fmt.Sprintf(
			`<tr>
                <td>%s</td>
                <td>%f</td>
                <td><img src="http:%s"/></td>
            </tr>`, data.name, data.price, data.img))
		database.Insert(data.name, data.img, data.price)
	}
	buffer.WriteString("</table>\n")
	buffer.WriteString(`
		<script type="text/javascript" src="http://g.alicdn.com/sj/lib/jquery/dist/jquery.min.js"></script>
			<script type="text/javascript" src="http://g.alicdn.com/sui/sui3/0.0.18/js/sui.min.js"></script>
		</body>
		</html>
	`)
	ioutil.WriteFile("jd.html", buffer.Bytes(), os.ModePerm)
	fmt.Println(time.Since(now))
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", "jd.html").Start()
	case "windows", "darwin":
		err = exec.Command("open", "jd.html").Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	fmt.Println(err)
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/loogo/gocrawler/database"
	// _ "github.com/mattn/go-sqlite3"
	_ "github.com/go-sql-driver/mysql"
)

type result struct {
	HTML    string
	HasMore bool
}

var ajaxURL = "http://shop.haocaisong.cn/shop/ajax/mall.php"

func crawler(url string, c chan hcProduct) int {
	count := 0
	document, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}
	cata1 := document.Find("#cate_list ul.types li")
	cata1.Each(func(j int, cata1Sel *goquery.Selection) {
		cata2Url, _ := cata1Sel.Find("a").Attr("href")
		cata2Url = url + cata2Url
		document, err = goquery.NewDocument(cata2Url)
		cata2 := document.Find("#cate2_container_ ul.swiper-wrapper li")
		cata2.Each(func(k int, cata2Sel *goquery.Selection) {
			i := 1
			href, _ := cata2Sel.Find("a").Attr("href")
			for {
				rawurl := fmt.Sprintf("%s%s&page=%d", ajaxURL, href, i)
				fmt.Println(rawurl)
				response, err := http.Get(rawurl)
				if err != nil {
					fmt.Println(err)
				}
				defer response.Body.Close()
				body, err := ioutil.ReadAll(response.Body)
				if err != nil {
					fmt.Println(err)
				}
				var jsonRes result
				err = json.Unmarshal(body, &jsonRes)
				if err != nil {
					fmt.Println(err)
				}

				if !jsonRes.HasMore || len(jsonRes.HTML) == 0 {
					break
				}
				if len(jsonRes.HTML) > 0 {

					htmlReader := strings.NewReader(jsonRes.HTML)
					doc, err := goquery.NewDocumentFromReader(htmlReader)

					if err != nil {
						log.Fatal(err)
					}
					root := doc.Find("li")

					root.Each(func(i int, s *goquery.Selection) {
						count++
						// fmt.Println(count)
						go func() {
							productID, _ := s.Attr("id")
							img, exist := s.Find(".gi img").Attr("src")
							if exist {
								img = strings.Split(img, "@")[0]
							} else {
								img = "Not Exist!"
							}
							info := s.Find(".intro")

							name := info.Find("h3.f15").Text()
							spec := info.Find("p.f14").Text()
							price := info.Find("em.f16").Parent().Text()

							hc := hcProduct{name: name, img: img, price: price, spec: spec, product_id: productID}
							fmt.Println(hc)

							downloadImg(img)
							c <- hc
						}()
					})
				}
				i++
			}
		})
	})

	return count
}
func downloadImg(url string) {
	response, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer response.Body.Close()

	imgArray := strings.Split(url, "/")
	imgName := imgArray[len(imgArray)-1]
	file, err := os.Create("images/" + imgName)
	if err != nil {
		log.Fatal(err)
	}
	// Use io.Copy to just dump the response body to the file. This supports huge files
	_, err = io.Copy(file, response.Body)
	if err != nil {
		log.Fatal(err)
	}
	file.Close()
}

func main() {
	cfg := loadconfig()
	db := database.MySQL{DataSourceName: cfg.DataSourceName}
	db.CreateDb()
	now := time.Now()
	url := "http://shop.haocaisong.cn/shop/mall.php"
	c := make(chan hcProduct)
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
	pIDs := make(map[string]string, length)
	for i := 0; i < length; i++ {
		data := <-c
		price := strings.Replace(data.price, "元", "@", 1)

		price = strings.Split(price, "@")[0][2:]

		imgURL := strings.Split(data.img, "/")
		imgID := imgURL[len(imgURL)-1]
		buffer.WriteString(fmt.Sprintf(
			`<tr>
				<td>%s</td>
				<td>%s</td>
                <td>%s</td>
                <td>%s</td>
				<td>%s</td>
				<td>%s</td>
                <td><img src="%s"/></td>
            </tr>`, data.product_id, data.name, data.spec, data.price, price, imgID, data.img))
		if _, ok := pIDs[data.product_id]; !ok {
			db.Insert(data.name, data.img, data.price, data.spec, data.product_id, price, imgID)
			pIDs[data.product_id] = ""
		}
	}
	buffer.WriteString("</table>\n")
	buffer.WriteString(`
		<script type="text/javascript" src="http://g.alicdn.com/sj/lib/jquery/dist/jquery.min.js"></script>
			<script type="text/javascript" src="http://g.alicdn.com/sui/sui3/0.0.18/js/sui.min.js"></script>
		</body>
		</html>
	`)
	ioutil.WriteFile("haocai.html", buffer.Bytes(), os.ModePerm)
	fmt.Println(time.Since(now))
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", "haocai.html").Start()
	case "windows", "darwin":
		err = exec.Command("open", "haocai.html").Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	fmt.Println(err)
}

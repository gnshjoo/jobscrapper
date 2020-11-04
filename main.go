package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

var baseURL string = "https://kr.indeed.com/jobs?q=python&limit=50"

func checkError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln(res.StatusCode)
	}

}

func getPages() int {
	pages := 0
	res, err := http.Get(baseURL)
	checkError(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkError(err)

	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		pages = s.Find("a").Length()
	})

	return pages
}

func getpage(page int) {
	pageURL := baseURL + "&start=" + strconv.Itoa(page*50)
	res, err := http.Get(pageURL)
	checkError(err)
	checkCode(res)

}

func main() {
	totalPages := getPages()

	for i := 0; i < totalPages; i++ {
		getpage(i)
	}
}

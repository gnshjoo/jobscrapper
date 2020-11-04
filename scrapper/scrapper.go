package scrapper

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type extractdJob struct {
	id       string
	title    string
	location string
	salary   string
	summary  string
}

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

func getPages(url string) int {
	pages := 0
	res, err := http.Get(url)
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

func getpage(page int, url string, mainC chan<- []extractdJob) {
	var jobs []extractdJob
	c := make(chan extractdJob)
	fmt.Println("Requesting " + url + "&start=" + strconv.Itoa(page*50))
	pageURL := url + "&start=" + strconv.Itoa(page*50)
	res, err := http.Get(pageURL)
	checkError(err)
	checkCode(res)

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkError(err)

	searchCards := doc.Find(".jobsearch-SerpJobCard")

	searchCards.Each(func(i int, card *goquery.Selection) {
		go extractJob(card, c)

	})

	for i := 0; i < searchCards.Length(); i++ {
		job := <-c
		jobs = append(jobs, job)
	}

	mainC <- jobs

}

func CleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), "")
}

func extractJob(card *goquery.Selection, c chan<- extractdJob) {
	id, _ := card.Attr("data-jk")
	title := CleanString(card.Find(".title>a").Text())
	location := CleanString(card.Find(".sjcl").Text())
	salary := CleanString(card.Find(".salaryText").Text())
	summary := CleanString(card.Find(".summary").Text())
	c <- extractdJob{
		id:       id,
		title:    title,
		location: location,
		salary:   salary,
		summary:  summary,
	}
}

func writeJobs(jobs []extractdJob) {
	file, err := os.Create("jobs.csv")
	checkError(err)

	w := csv.NewWriter(file)
	defer w.Flush()

	headers := []string{"Link", "Title", "Location", "Salary", "Summary"}

	wErr := w.Write(headers)
	checkError(wErr)

	for _, job := range jobs {
		jobSlice := []string{"https://kr.indeed.com/viewjob?jk=" + job.id, job.title, job.location, job.salary, job.summary}
		wErr := w.Write(jobSlice)
		checkError(wErr)
	}
}

// Scrape Indeed
func Scrape(term string) {
	var baseURL string = "https://kr.indeed.com/jobs?q=" + term + "&limit=50"
	var jobs []extractdJob
	c := make(chan []extractdJob)
	totalPages := getPages(baseURL)

	for i := 0; i < totalPages; i++ {
		go getpage(i, baseURL, c)
	}

	for i := 0; i < totalPages; i++ {
		extractdJob := <-c
		jobs = append(jobs, extractdJob...)
	}

	writeJobs(jobs)
	fmt.Println("Done, extracted :", len(jobs))
}

package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/julienschmidt/httprouter"
)

var waitGroup sync.WaitGroup
var news_map map[string]NewsMap

type SitemapIndex struct {
	Locations []string `xml:"sitemap>loc"`
}

type News struct {
	Titles    []string `xml:"url>news>title"`
	Keywords  []string `xml:"url>news>keywords"`
	Locations []string `xml:"url>loc"`
}

type Config struct {
	Port int `yaml:"config, port"`
}

type NewsMap struct {
	Keyword  string
	Location string
}

type NewsAggPage struct {
	Title string
	News  map[string]NewsMap
}

func indexHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprintf(w, "<h1>Hello Worl2d</h1>")
}

func fetchNewsLoction(c chan News, Location string) {

	defer waitGroup.Done()
	var n News

	resp, _ := http.Get(Location)
	bytes, _ := ioutil.ReadAll(resp.Body)
	xml.Unmarshal(bytes, &n)

	resp.Body.Close()

	c <- n
}

func newAggHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	var s SitemapIndex

	news_map := make(map[string]NewsMap)
	news_chan := make(chan News, 30)

	resp, _ := http.Get("https://www.washingtonpost.com/news-sitemap-index.xml")
	bytes, _ := ioutil.ReadAll(resp.Body)

	xml.Unmarshal(bytes, &s)

	resp.Body.Close()

	for _, Location := range s.Locations {
		go fetchNewsLoction(news_chan, Location)
		waitGroup.Add(1)
	}

	waitGroup.Wait()
	close(news_chan)

	for elem := range news_chan {
		for idx, _ := range elem.Titles {
			news_map[elem.Titles[idx]] = NewsMap{elem.Keywords[idx], elem.Locations[idx]}
		}
	}

	p := NewsAggPage{Title: "Amazing News Aggregator", News: news_map}
	t, _ := template.ParseFiles("basictemplating.html")
	t.Execute(w, p)
}

func newurl(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	key := ps.ByName("key")
	fmt.Fprintf(w, "<h1>Welcome to "+key+"</h1>")
}

func main() {
	router := httprouter.New()
	router.GET("/", indexHandler)
	router.GET("/agg/", newAggHandler)
	router.GET("/newurl/:key", newurl)

	http.ListenAndServe(":8000", router)
}

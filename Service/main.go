package Service

import (
	"net/http"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"regexp"
	"sync"
	"strconv"
	"strings"

	"github.com/op/go-logging"
	"golang.org/x/net/html"
	"io"
)

var (
	log = logging.MustGetLogger("service")
)

func (s Server) Listen() {
	http.HandleFunc("/", ArrayUrl)
	http.ListenAndServe(":9990", nil)
}

func ArrayUrl(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		defer r.Body.Close()
		bytes, err := ioutil.ReadAll(r.Body)
		log.Infof("POST %s", string(bytes))

		if err != nil {
			fmt.Fprintf(w, "Can't read bytes. err: %v", err)
			return
		}

		var urls []string
		err = json.Unmarshal(bytes, &urls)
		if err != nil {
			log.Errorf("Can't unmarshal data. err: %s", err)
		}

		items := make(chan *Item, len(urls))

		go ParseUrl(urls, items)

		resp := toSlice(items)

		jsonItems, err := json.Marshal(resp)
		if err != nil {
			log.Errorf("Can't marshal array Items: %s", resp)
		}

		status, err := w.Write(jsonItems)
		if err != nil {
			log.Errorf("Can't send response %s \nstatus %d, err: %s", items, status, err)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			log.Infof("Status: %d", status)

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
		}

	default:
		log.Infof("Sorry, only GET and POST methods are supported.")
	}
}

func ParseUrl(urls []string, items chan *Item) {
	wg := sync.WaitGroup{}
	for _, url := range urls {
		reg, _ := regexp.Compile("^http(s?)://([\\w.]+)\\.([a-z]{2,6}\\.?)(/[\\w.]*)*[/a-z-\\d]*/?$")
		if !reg.MatchString(url) {
			log.Errorf("Invalid URL: %s", url)
			continue
		}

		wg.Add(1)
		go func() {
			items <- BuildItem(url)
			defer wg.Done()
		}()
	}
	wg.Wait()
	defer close(items)
}

func BuildItem(url string) *Item {
	res, err := http.Get(url)
	defer res.Body.Close()
	if err != nil {
		log.Errorf("Can't GET from url: %s \nerr: %s", url, err)
	}
	status, err := strconv.ParseInt(strings.Split(res.Status, " ")[0], 10, 32)
	contentType := strings.Split(res.Header.Get("content-type"), ";")[0]
	contentLength := res.ContentLength
	meta := Meta{
		Status: status,
	}
	if len(contentType) != 0 {
		meta.ContentType = contentType
	}
	if contentLength > 0 {
		meta.ContentLength = contentLength
	}

	tags, length := CountTag(res.Body)

	if meta.ContentLength < 1 {
		meta.ContentLength = length
	}

	elements := []Element{}
	for k, v := range tags {
		elements = append(elements, Element{Tag: k, Count: v})
	}

	item := Item{
		Url:      url,
		Meta:     meta,
		Elements: elements,
	}
	return &item
}

func CountTag(body io.Reader) (map[string]int64, int64) {
	result := map[string]int64{}
	tok:=html.NewTokenizer(body)
	for {
		tt := tok.Next()
		if tt == html.ErrorToken {
			return result, 0
		}
		if tt == html.StartTagToken || tt == html.SelfClosingTagToken{
			t := tok.Token()
			if _, ok := result[t.Data]; !ok {
				result[t.Data] = 1
				continue
			}
			result[t.Data] += 1
		}
	}
	bytes, _ := ioutil.ReadAll(body)
	length := int64(len(bytes))

	return result, length
}

func toSlice(c chan *Item) []*Item {
	s := make([]*Item, 0)
	for i := range c {
		s = append(s, i)
	}
	return s
}

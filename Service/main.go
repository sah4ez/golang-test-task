package Service

import (
	"net/http"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"regexp"
	"sync"
	"strconv"

	"github.com/op/go-logging"
)

var (
	log = logging.MustGetLogger("service")
)

type Server struct {
	Host string
	Port string
}

type urls []string

type Item struct {
	Url      string    `json:"url"`
	Meta     Meta      `json:"meta"`
	Elements []Element `json:"elements"`
}

type Meta struct {
	Status        int64  `json:"status"`
	ContentType   string `json:"content-type"`
	ContentLength int64  `json:"content-length"`
}

type Element struct {
	Tag   string `json:"tag-name"`
	Count int64  `json:"count"`
}

func (s Server) Listen() {
	addr := fmt.Sprintf("%s:%s", s.Host, s.Port)
	log.Infof("Start Listen %s", addr)
	http.HandleFunc("/", ArrayUrl)
	http.ListenAndServe(addr, nil)
}

func ArrayUrl(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		log.Infof("GET")
	case "POST":
		log.Infof("POST")
		defer r.Body.Close()
		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Fprintf(w, "Can't read bytes. err: %v", err)
			return
		}

		var urls []string
		err = json.Unmarshal(bytes, &urls)
		if err != nil {
			log.Errorf("Can't unmarshal data. err: %s", err)
		}

		items := make(chan interface{}, len(urls))

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
		}else{
			log.Infof("Status: %d", status)

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
		}


	default:
		log.Infof("Sorry, only GET and POST methods are supported.")
	}
}

func ParseUrl(urls []string, items chan interface{}) {
	wg := sync.WaitGroup{}
	for _, url := range urls {
		reg, _ := regexp.Compile("^(https?://)?([\\w.]+)\\.([a-z]{2,6}\\.?)(/[\\w.]*)*/?$")
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

func BuildItem(url string) Item {
	res, err := http.Get(url)
	defer res.Body.Close()
	if err != nil {
		log.Errorf("Can't GET from url: %s \nerr: %s", url, err)
	}
	status, err := strconv.ParseInt(res.Status, 10, 32)
	meta := Meta{
		Status:        status,
		ContentType:   res.Header.Get("content-type"),
		ContentLength: res.ContentLength,
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Errorf("Can't read body from url {}", url)
	}
	tags := CountTag(string(body))

	elements := []Element{}
	for k, v := range tags {
		elements = append(elements, Element{Tag: k, Count: v})
	}

	item := Item{
		Url:      url,
		Meta:     meta,
		Elements: elements,
	}
	return item
}

func CountTag(body string) map[string]int64 {
	result := map[string]int64{}
	regex := regexp.MustCompile("<([^>/ ]+)")
	tags := regex.FindAllString(body, -1)
	for _, tag := range tags {
		tagName := tag[1:]
		if _, ok := result[tagName]; !ok {
			result[tagName] = 1
			continue
		}
		result[tagName] += 1
	}
	return result
}

func toSlice(c chan interface{}) []interface{} {
	s := make([]interface{}, 0)
	for i := range c {
		s = append(s, i)
	}
	return s
}

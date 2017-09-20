package Service

import (
	"net/http"
	"fmt"
	"github.com/op/go-logging"
	"github.com/xeipuuv/gojsonschema"
	"io/ioutil"
	"encoding/json"
	"regexp"
	"sync"
	"strconv"
)

var (
	log = logging.MustGetLogger("service")
	schemaLoader = gojsonschema.NewReferenceLoader("file:///./Schema.json")
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
	Status        int64    `json:"status"`
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

		go func() {
			wg := sync.WaitGroup{}
			for _, url := range urls {
				wg.Add(1)
				go func() {
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
					tags := countTag(string(body))

					elements := []Element{}
					for k, v := range tags {
						elements = append(elements, Element{Tag: k, Count: v})
					}

					item := Item{
						Url:      url,
						Meta:     meta,
						Elements: elements,
					}
					items <- item
					defer wg.Done()
				}()
			}
			wg.Wait()
			defer close(items)
		}()

		resp := toSlice(items)

		jsonItems, _ := json.Marshal(resp)

		document := gojsonschema.NewReferenceLoader(string(jsonItems))
		result, err := gojsonschema.Validate(schemaLoader, document)
		if err != nil {
			log.Errorf("Error in process validation")
		}

		if !result.Valid(){
			log.Errorf("Invalid JSON document")
		}

		status, err := w.Write(jsonItems)
		if err != nil {
			log.Errorf("Can't send response %s \nstatus %d, err: %s", items, status, err)
		}
		log.Infof("Status: %d", status)

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)

	default:
		log.Infof("Sorry, only GET and POST methods are supported.")
	}
}

func countTag(body string) map[string]int64 {
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

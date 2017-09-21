package Service

type Server struct {
}

type urls []string

type Item struct {
	Url      string    `json:"url"`
	Meta     Meta      `json:"meta"`
	Elements []Element `json:"elements,omitempty"`
}

type Meta struct {
	Status        int64  `json:"status"`
	ContentType   string `json:"content-type,omitempty"`
	ContentLength int64  `json:"content-length,omitempty"`
}

type Element struct {
	Tag   string `json:"tag-name"`
	Count int64  `json:"count"`
}

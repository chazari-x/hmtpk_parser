package model

type Schedule struct {
	Date    string   `json:"date"`
	Lessons []Lesson `json:"lesson"`
	Href    string   `json:"href"`
}

type Lesson struct {
	Num      string `json:"num"`
	Time     string `json:"time"`
	Name     string `json:"name"`
	Room     string `json:"room"`
	Location string `json:"location"`
	Group    string `json:"group"`
	Subgroup string `json:"subgroup"`
	Teacher  string `json:"teacher"`
}

type Option struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type Announces struct {
	Announces []Announce `json:"announces"`
	LastPage  int        `json:"last_page"`
}

type Announce struct {
	Path  string `json:"path"`
	Date  string `json:"date"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

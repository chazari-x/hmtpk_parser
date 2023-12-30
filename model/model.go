package model

type Schedule struct {
	Date    string   `json:"date"`
	Lessons []Lesson `json:"lesson"`
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

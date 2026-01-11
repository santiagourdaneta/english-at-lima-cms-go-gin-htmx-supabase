package models

type Sentence struct {
	ID      int    `json:"id,omitempty"`
	English string `json:"english"`
	Spanish string `json:"spanish"`
}

type Quiz struct {
	ID       int    `json:"id,omitempty"`
	Question string `json:"question"`
	Opt1     string `json:"opt1"`
	Opt2     string `json:"opt2"`
	Opt3     string `json:"opt3"`
	Correct  string `json:"correct"`
}

type Resource struct {
	ID    int    `json:"id,omitempty"`
	Title string `json:"title"`
	URL   string `json:"url"`
	Type  string `json:"type"`
}

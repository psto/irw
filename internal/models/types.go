package models

type Track struct {
	Path       string
	Type       string
	Interval   float64
	Afactor    float64
	DueDate    int64
	IsFinished int
	Priority   int
}

type Session struct {
	Date     string
	Duration int
	Reviewed int
	Finished int
}

type ZkNote struct {
	AbsPath  string `json:"absPath"`
	Metadata struct {
		Priority *int `json:"priority"`
	} `json:"metadata"`
}

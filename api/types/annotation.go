package types

type Annotation struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	IsRegex bool   `json:"isRegex"`
}

type Annotations []Annotation

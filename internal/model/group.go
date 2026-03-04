package model

type Group struct {
	Name      string     `json:"name"`
	Instances []Instance `json:"instances"`
}

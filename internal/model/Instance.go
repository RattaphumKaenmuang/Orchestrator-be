package model

type Instance struct {
	ID       string
	Provider string
	CPU      int
	RAM      int
	GPU      bool
	Status   string
}

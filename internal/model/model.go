package model

const (
	TypeAction = "action"
	TypeSelect = "single-select"
	TypeText   = "text"
)

type Dependency struct {
	ID          string
	Type        string
	Name        string
	Description string
}

type Dependencies []Dependency

type ProjectRequest struct {
	Project      string
	Language     string
	SpringBoot   string
	Group        string
	Artifact     string
	PackageName  string
	Packaging    string
	JavaVersion  string
	Name         string
	Description  string
	Dependencies string
}

package model

type DepGraph struct {
	Vendor   string
	Name     string
	Version  string
	Language string
	Parents  []*DepGraph
	Children []*DepGraph
}

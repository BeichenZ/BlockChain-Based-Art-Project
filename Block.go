package main

type Block struct {
	CurrentHash string
	PreviousHash string
	OPS []Operation
}

type Operation struct {
	Command string
	Shapetype string
	ShapeSvgString string
	Fill string
	Stroke string
}
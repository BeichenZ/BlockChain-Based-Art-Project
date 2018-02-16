package main

import (
	"encoding/json"
	"net/http"

	"./blockartlib"
	shared "./shared"
)

//User defines model for storing account details in database
type SVGPayload struct {
	SVGs []shared.FullSvgInfo
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/echo", echoHandler)

	http.ListenAndServe(":5000", mux)
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	//Marshal or convert user object back to json and write to response
	svgPayload := SVGPayload{blockartlib.GetListOfOps()}
	s, err := json.Marshal(svgPayload)
	if err != nil {
		panic(err)
	}

	//Set Content-Type header so that clients will know how to read response
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set(
		"Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization",
	)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.WriteHeader(http.StatusOK)
	//Write json response back to response
	w.Write(s)
}

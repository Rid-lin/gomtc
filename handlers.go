package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// func handleIndex() http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		fmt.Fprintf(w,
// 			`<html>
// 			<head>
// 			<title>golang-netflow-to-squid</title>
// 			</head>
// 			<body>
// 			Более подробно на https://github.com/Rid-lin/gonflux
// 			</body>
// 			</html>
// 			`)
// 	}
// }

func handleIndex(w http.ResponseWriter, r *http.Request) {
	indextmpl, err := template.ParseFiles("assets/index.html", "assets/header.html", "assets/menu.html", "assets/right.html", "assets/footer.html")
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	err = indextmpl.ExecuteTemplate(w, "index", nil)
	if err != nil {
		fmt.Fprint(w, err.Error())
	}
}

func (data *transport) handleReport(w http.ResponseWriter, r *http.Request) {

	indextmpl, err := template.ParseFiles("assets/index.html", "assets/header.html", "assets/menu.html", "assets/right.html", "assets/footer.html")
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	err = indextmpl.ExecuteTemplate(w, "index", nil)
	if err != nil {
		fmt.Fprint(w, err.Error())
	}
}

func (data *transport) handleFlow(w http.ResponseWriter, r *http.Request) {

	indextmpl, err := template.ParseFiles("assets/flow.html", "assets/header.html", "assets/menu.html", "assets/report.html", "assets/footer.html")
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	err = indextmpl.ExecuteTemplate(w, "flow", nil)
	if err != nil {
		fmt.Fprint(w, err.Error())
	}
}

func (data *transport) getmacHandler() http.HandlerFunc {
	var (
		request  request
		Response ResponseType
	)

	return func(w http.ResponseWriter, r *http.Request) {
		request.Time = r.URL.Query().Get("time")
		request.IP = r.URL.Query().Get("ip")
		Response = data.GetInfo(&request)
		log.Debugf(" | Request:'%v','%v' response:'%v'", request.Time, request.IP, Response.Mac)
		responseJSON, err := json.Marshal(Response)
		if err != nil {
			log.Errorf("Error Marshaling mac'%v'to JSON:'%v'", Response.Mac, err)
		}
		// fmt.Fprint(w, mac)
		_, err2 := w.Write(responseJSON)
		if err2 != nil {
			log.Errorf("Error send response:%v", err2)
		}
	}
}

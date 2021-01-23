package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type LineOfData struct {
	ip,
	mac,
	timeout,
	hostName,
	comment string
	timeoutInt int64
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w,
		`<html>
			<head>
			<title>go-macfrommikrotik</title>
			</head>
			<body>
			Более подробно на https://github.com/Rid-lin/gonsquid
			</body>
			</html>
			`)
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
		_, err2 := w.Write(responseJSON)
		if err2 != nil {
			log.Errorf("Error send response:%v", err2)
		}
	}
}

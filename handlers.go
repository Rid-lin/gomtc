package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func (transport *Transport) handleRequest() {
	http.HandleFunc("/", logreq(transport.handleIndex))
	http.HandleFunc("/withfriends/", logreq(transport.handleWithFriends))
	http.HandleFunc("/dayDetail", logreq(transport.handleDayDetail))
	http.HandleFunc("/log/", logreq(transport.handleLog))
	http.HandleFunc("/runparse/", logreq(transport.handleRunParse))
	http.HandleFunc("/editalias/", logreq(transport.handleEditAlias))
	http.HandleFunc("/", logreq(handleIndex))
	http.HandleFunc("/getmac", logreq(transport.handlerGetMac()))
	http.HandleFunc("/setstatusdevices", logreq(transport.handlerSetStatusDevices))
	http.HandleFunc("/getstatusdevices", logreq(transport.handlerGetStatusDevices))
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(cfg.AssetsPath))))

	log.Infof("gomtc listens to:%v", cfg.BindAddr)

	go func() {
		err := http.ListenAndServe(cfg.BindAddr, nil)
		if err != nil {
			log.Fatal("http-server returned error:", err)
		}
	}()

}

func logreq(f func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("path: %s", r.URL.Path)

		f(w, r)
	})
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w,
		`<html>
			<head>
			<title>go-macfrommikrotik</title>
			</head>
			<body>
			Более подробно на https://github.com/Rid-lin/gomtc
			</body>
			</html>
			`)
}

func (data *Transport) handlerGetMac() http.HandlerFunc {
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

func (data *Transport) handlerSetStatusDevices(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		errorResponse(w, "Content Type is not application/json", http.StatusUnsupportedMediaType)
		return
	}

	result := map[string]bool{}
	var unmarshalErr *json.UnmarshalTypeError

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&result)
	if err != nil {
		if errors.As(err, &unmarshalErr) {
			errorResponse(w, "Bad Request. Wrong Type provided for field "+unmarshalErr.Field, http.StatusBadRequest)
		} else {
			errorResponse(w, "Bad Request "+err.Error(), http.StatusBadRequest)
		}
		return
	}
	data.syncStatusDevices(result)

	errorResponse(w, "Recived", http.StatusOK)
	log.Println(result)
}

func (data *Transport) handlerGetStatusDevices(w http.ResponseWriter, r *http.Request) {
	defaultLine := LineOfData{}

	data.RLock()
	dataToDSend := data.ipToMac

	defaultLine.HourlyQuota = data.HourlyQuota
	defaultLine.DailyQuota = data.DailyQuota
	defaultLine.MonthlyQuota = data.MonthlyQuota
	data.RUnlock()

	dataToDSend["default"] = defaultLine

	json_data, err := json.Marshal(dataToDSend)
	if err != nil {
		log.Errorf("Error witn Marshaling to JSON status of all devices:(%v)", err)
	}
	fmt.Fprint(w, string(json_data))
}

func errorResponse(w http.ResponseWriter, message string, httpStatusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode)
	resp := make(map[string]string)
	resp["message"] = message
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Error(err)
	}
	_, err = w.Write(jsonResp)
	if err != nil {
		log.Error(err)
	}
}

func parseDataFromURL(r *http.Request) RequestForm {

	var request RequestForm

	request.path = r.URL.Path

	u, err := url.Parse(r.URL.String())
	if err != nil {
		log.Error(err)
	}
	m, _ := url.ParseQuery(u.RawQuery)
	// Checking the availability of data from the URL. To show today if there is no data.
	if len(m["date_from"]) > 0 {
		request.dateFrom = m["date_from"][0]
	} else {
		request.dateFrom = time.Now().Format("2006-01-02")
	}
	if len(m["date_to"]) > 0 {
		request.dateTo = m["date_to"][0]
	} else {
		request.dateTo = time.Now().Add(24 * time.Hour).Format("2006-01-02")
	}
	if request.dateFrom == "" {
		request.dateFrom = time.Now().Format("2006-01-02")
	}
	if request.dateTo == "" {
		request.dateTo = time.Now().Add(24 * time.Hour).Format("2006-01-02")
	}
	if len(m["report"]) > 0 {
		request.report = m["report"][0]
		// } else {
		// request.report = "index"
	}

	return request
}

func (data *Transport) handleIndex(w http.ResponseWriter, r *http.Request) {

	request := parseDataFromURL(r)
	log.Debug("request=", request)

	path := data.AssetsPath
	// Starting template processing to display the page in the browser
	indextmpl, err := template.ParseFiles(
		path+"/index.html",
		path+"/header.html",
		path+"/footer.html")
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	DisplayData := data.reportTrafficHourlyByLogins(request, false)

	err = indextmpl.ExecuteTemplate(w, "index", DisplayData)
	if err != nil {

		fmt.Fprintf(w, "Что-то пошло не так, произошла ошибка при выполнении запроса. Проверьте налиие логов за запрашиваемый период\n%v", err.Error())
	}
}

func (data *Transport) handleWithFriends(w http.ResponseWriter, r *http.Request) {

	request := parseDataFromURL(r)
	log.Debug("request=", request)

	path := data.AssetsPath
	// Starting template processing to display the page in the browser
	indextmpl, err := template.ParseFiles(
		path+"/indexwf.html",
		path+"/header.html",
		path+"/footer.html")
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	DisplayData := data.reportTrafficHourlyByLogins(request, true)
	DisplayData.SizeOneMegabyte = data.SizeOneMegabyte

	err = indextmpl.ExecuteTemplate(w, "indexwf", DisplayData)
	if err != nil {
		if strings.Contains(fmt.Sprint(err), "index out of range") {
			fmt.Fprintf(w, "Проверьте налиие логов за запрашиваемый период<br> или подождите несколько минут.")
		} else {
			fmt.Fprintf(w, "Что-то пошло не так, произошла ошибка при выполнении запроса. <br> %v", err.Error())

		}
	}
}

func (t *Transport) handleEditAlias(w http.ResponseWriter, r *http.Request) {
	// path := r.URL.Path
	// alias := r.URL.Fragment
	t.RLock()
	alias := r.FormValue("alias")
	nowDay := findOutTheCurrentDay(time.Now().Unix(), t.Location)
	nowDayStr := time.Unix(nowDay, 0).In(t.Location).Format("2006-01-02")
	key := KeyMapOfReports{
		Alias:   alias,
		DateStr: nowDayStr,
	}

	lineOfDisplay, ok := t.dataChashe[key]
	if !ok {
		lineOfDisplay = ValueMapOfReports{}
	}
	t.RUnlock()

	DisplayDataUser := DisplayDataUserType{
		Header:    "Редактирование пользователя",
		Copyright: "GoSquidLogAnalyzer <i>© 2020</i> by Vladislav Vegner",
		Mail:      "mailto:vegner.vs@uttist.ru",
		LineOfDisplay: LineOfDisplay{
			Alias: alias,
			DeviceType: DeviceType{
				HostName: lineOfDisplay.HostName,
				TypeD:    lineOfDisplay.TypeD,
			},
			PersonType: PersonType{
				Name:     lineOfDisplay.Name,
				Position: lineOfDisplay.Position,
				Company:  lineOfDisplay.Company,
				Comments: lineOfDisplay.Comments,
			},
			QuotaType: QuotaType{
				HourlyQuota:  lineOfDisplay.HourlyQuota,
				DailyQuota:   lineOfDisplay.DailyQuota,
				MonthlyQuota: lineOfDisplay.MonthlyQuota,
				Blocked:      lineOfDisplay.Blocked,
			},
		},
	}

	AssetsPath := t.AssetsPath
	// Starting template processing to display the page in the browser
	indextmpl, err := template.ParseFiles(
		AssetsPath+"/editalias.html",
		AssetsPath+"/header.html",
		AssetsPath+"/footer.html")
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	err = indextmpl.ExecuteTemplate(w, "editalias", DisplayDataUser)
	if err != nil {

		fmt.Fprintf(w, "Что-то пошло не так, произошла ошибка при выполнении запроса. Проверьте налиие логов за запрашиваемый период\n%v", err.Error())
	}
}

func (t *Transport) handleLog(w http.ResponseWriter, r *http.Request) {
	path := t.AssetsPath
	// Starting template processing to display the page in the browser
	indextmpl, err := template.ParseFiles(
		path+"/log.html",
		path+"/header.html",
		path+"/footer.html")
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	DisplayData := &DisplayDataType{
		Header: "Лог работы",
		Logs:   t.logs,
	}

	err = indextmpl.ExecuteTemplate(w, "log", DisplayData)
	if err != nil {

		fmt.Fprintf(w, "Что-то пошло не так, произошла ошибка при выполнении запроса. <br> %v", err.Error())
	}
}

func (t *Transport) handleRunParse(w http.ResponseWriter, r *http.Request) {
	t.timer.Stop()
	t.timer.Reset(1 * time.Second)

	http.Redirect(w, r, "/", 302)
}

// func (data *Transport) handleLog(w http.ResponseWriter, r *http.Request) {
// 	path := data.AssetsPath
// 	// Starting template processing to display the page in the browser
// 	indextmpl, err := template.ParseFiles(
// 		path+"/log.html",
// 		path+"/header.html",
// 		path+"/footer.html")
// 	if err != nil {
// 		fmt.Fprint(w, err.Error())
// 		return
// 	}

// 	DisplayData := ""

// 	err = indextmpl.ExecuteTemplate(w, "log", DisplayData)
// 	if err != nil {

// 		fmt.Fprintf(w, "Что-то пошло не так, произошла ошибка при выполнении запроса. Проверьте налиие логов за запрашиваемый период\n%v", err.Error())
// 	}
// }
func (data *Transport) handleDayDetail(w http.ResponseWriter, r *http.Request) {

	request := parseDataFromURL(r)
	log.Debug("request=", request)

	path := data.AssetsPath
	// Starting template processing to display the page in the browser
	indextmpl, err := template.ParseFiles(
		path+"/dayDetail.html",
		path+"/header.html",
		path+"/footer.html")
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	DisplayData := data.reportTrafficHourlyByLogins(request, false)

	err = indextmpl.ExecuteTemplate(w, "dayDetail", DisplayData)
	if err != nil {

		fmt.Fprintf(w, "Что-то пошло не так, произошла ошибка при выполнении запроса. Проверьте налиие логов за запрашиваемый период\n%v", err.Error())
	}
}

package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func (transport *Transport) handleRequest(cfg *Config) {
	http.HandleFunc("/", logreq(transport.handleIndex))
	http.HandleFunc("/withfriends/", logreq(transport.handleWithFriends))
	http.HandleFunc("/log/", logreq(transport.handleLog))
	http.HandleFunc("/runparse/", logreq(transport.handleRunParse))
	http.HandleFunc("/editalias/", logreq(transport.handleEditAlias))
	// http.HandleFunc("/dayDetail/", logreq(transport.handleDayDetail))
	// http.HandleFunc("/getmac", logreq(transport.handlerGetMac()))
	// http.HandleFunc("/setstatusdevices/", logreq(transport.handlerSetStatusDevices))
	// http.HandleFunc("/getstatusdevices/", logreq(transport.handlerGetStatusDevices))
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(cfg.AssetsPath))))

	log.Infof("gomtc listens HTTP on:'%v'", cfg.ListenAddr)

	go func() {
		err := http.ListenAndServe(cfg.ListenAddr, nil)
		if err != nil {
			log.Fatal("http-server returned error:", err)
		}
	}()

}

func logreq(f func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("access:%s?%s", r.URL.Path, r.URL.RawQuery)

		f(w, r)
	})
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
	assetsPath := data.AssetsPath
	DisplayData := data.reportTrafficHourlyByLogins(request, false)

	fmap := template.FuncMap{
		"FormatSize": FormatSize,
	}
	t := template.Must(template.New("index").Funcs(fmap).ParseFiles(
		assetsPath+"/index.html",
		assetsPath+"/header.html",
		assetsPath+"/footer.html"))
	err := t.Execute(w, DisplayData)
	if err != nil {
		if strings.Contains(fmt.Sprint(err), "index out of range") {
			fmt.Fprintf(w, "Проверьте налиие логов за запрашиваемый период<br> или подождите несколько минут.")
		} else {
			fmt.Fprintf(w, "Что-то пошло не так, произошла ошибка при выполнении запроса. <br> %v", err.Error())
		}
	}
}

func (data *Transport) handleWithFriends(w http.ResponseWriter, r *http.Request) {

	request := parseDataFromURL(r)
	assetsPath := data.AssetsPath
	DisplayData := data.reportTrafficHourlyByLogins(request, true)

	fmap := template.FuncMap{
		"FormatSize": FormatSize,
	}
	t := template.Must(template.New("indexwf").Funcs(fmap).ParseFiles(
		assetsPath+"/indexwf.html",
		assetsPath+"/header.html",
		assetsPath+"/footer.html"))
	err := t.Execute(w, DisplayData)
	if err != nil {
		if strings.Contains(fmt.Sprint(err), "index out of range") {
			fmt.Fprintf(w, "Проверьте налиие логов за запрашиваемый период<br> или подождите несколько минут.")
		} else {
			fmt.Fprintf(w, "Что-то пошло не так, произошла ошибка при выполнении запроса. <br> %v", err.Error())
		}
	}
}

func (t *Transport) handleEditAlias(w http.ResponseWriter, r *http.Request) {
	alias := r.FormValue("alias")
	// t.RLock()
	// nowDay := findOutTheCurrentDay(time.Now().Unix(), t.Location)
	// nowDayStr := time.Unix(nowDay, 0).In(t.Location).Format("2006-01-02")
	// key := KeyMapOfReports{
	// 	Alias:   alias,
	// 	DateStr: nowDayStr,
	// }

	// lineOfDisplay, ok := t.dataCashe[key]
	// if !ok {
	// 	lineOfDisplay = ValueMapOfReports{}
	// }
	// t.RUnlock()
	InfoOfDevice := t.aliasToDevice(alias)

	DisplayDataUser := DisplayDataUserType{
		Header:           "Редактирование пользователя",
		Copyright:        "GoSquidLogAnalyzer <i>© 2020</i> by Vladislav Vegner",
		Mail:             "mailto:vegner.vs@uttist.ru",
		Alias:            alias,
		InfoOfDeviceType: InfoOfDevice,
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

func FormatSize(size, SizeOneKilobyte uint64) string {
	var Size float64
	var Suffix string
	if size > (SizeOneKilobyte * SizeOneKilobyte * SizeOneKilobyte) {
		Size = float64(size) / float64(SizeOneKilobyte*SizeOneKilobyte*SizeOneKilobyte)
		Suffix = "Gb"
	} else if size > (SizeOneKilobyte * SizeOneKilobyte) {
		Size = float64(size) / float64(SizeOneKilobyte*SizeOneKilobyte)
		Suffix = "Mb"
	} else if size > (SizeOneKilobyte) {
		Size = float64(size) / float64(SizeOneKilobyte)
		Suffix = "Kb"
	} else {
		Size = float64(size)
		Suffix = "b"
	}
	return fmt.Sprintf("%3.2f.%s", Size, Suffix)
}

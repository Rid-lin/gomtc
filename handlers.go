package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func (transport *Transport) handleRequest(cfg *Config) {
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(cfg.AssetsPath))))

	http.HandleFunc("/", logreq(transport.handleIndex))
	http.HandleFunc("/wf/", logreq(transport.handleWithFriends))
	http.HandleFunc("/log/", logreq(transport.handleLog))
	http.HandleFunc("/runparse", logreq(transport.handleRunParse))
	http.HandleFunc("/editalias/", logreq(transport.handleEditAlias))

	log.Infof("gomtc listens HTTP on:'%v'", cfg.ListenAddr)

	go func() {
		err := http.ListenAndServe(cfg.ListenAddr, nil)
		if err != nil {
			log.Fatal("http-server returned error:", err)
			transport.exitChan <- os.Kill
		}
	}()

}

func logreq(f func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("access:%s%s?%s", r.Host, r.URL.Path, r.URL.RawQuery)

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

func (data *Transport) handleShowReport(w http.ResponseWriter, withfriends bool, preffix string, r *http.Request) {

	request := parseDataFromURL(r)
	request.referURL = r.Host + r.URL.Path
	request.path = r.URL.Path
	data.RLock()
	assetsPath := data.AssetsPath
	data.RUnlock()
	DisplayData := data.reportTrafficHourlyByLogins(request, withfriends)

	fmap := template.FuncMap{
		"FormatSize": FormatSize,
	}
	t := template.Must(template.New("index"+preffix).Funcs(fmap).ParseFiles(
		assetsPath+"/index"+preffix+".html",
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

func (data *Transport) handleIndex(w http.ResponseWriter, r *http.Request) {
	data.handleShowReport(w, false, "", r)
}

func (data *Transport) handleWithFriends(w http.ResponseWriter, r *http.Request) {
	data.handleShowReport(w, true, "wf", r)
}

// func (data *Transport) handleIndex(w http.ResponseWriter, r *http.Request) {

// 	request := parseDataFromURL(r)
// 	request.referURL = r.Host + r.URL.Path
// 	request.path = r.URL.Path
// 	data.RLock()
// 	assetsPath := data.AssetsPath
// 	data.RUnlock()
// 	DisplayData := data.reportTrafficHourlyByLogins(request, false)

// 	fmap := template.FuncMap{
// 		"FormatSize": FormatSize,
// 	}
// 	t := template.Must(template.New("index").Funcs(fmap).ParseFiles(
// 		assetsPath+"/index.html",
// 		assetsPath+"/header.html",
// 		assetsPath+"/footer.html"))
// 	err := t.Execute(w, DisplayData)
// 	if err != nil {
// 		if strings.Contains(fmt.Sprint(err), "index out of range") {
// 			fmt.Fprintf(w, "Проверьте налиие логов за запрашиваемый период<br> или подождите несколько минут.")
// 		} else {
// 			fmt.Fprintf(w, "Что-то пошло не так, произошла ошибка при выполнении запроса. <br> %v", err.Error())
// 		}
// 	}
// }

// func (data *Transport) handleWithFriends(w http.ResponseWriter, r *http.Request) {

// 	request := parseDataFromURL(r)
// 	request.referURL = r.Host + r.URL.Path
// 	request.path = r.URL.Path
// 	data.RLock()
// 	assetsPath := data.AssetsPath
// 	data.RUnlock()
// 	DisplayData := data.reportTrafficHourlyByLogins(request, true)

// 	fmap := template.FuncMap{
// 		"FormatSize": FormatSize,
// 	}
// 	t := template.Must(template.New("indexwf").Funcs(fmap).ParseFiles(
// 		assetsPath+"/indexwf.html",
// 		assetsPath+"/header.html",
// 		assetsPath+"/footer.html"))
// 	err := t.Execute(w, DisplayData)
// 	if err != nil {
// 		if strings.Contains(fmt.Sprint(err), "index out of range") {
// 			fmt.Fprintf(w, "Проверьте налиие логов за запрашиваемый период<br> или подождите несколько минут.")
// 		} else {
// 			fmt.Fprintf(w, "Что-то пошло не так, произошла ошибка при выполнении запроса. <br> %v", err.Error())
// 		}
// 	}
// }

func (data *Transport) handleEditAlias(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		alias := r.FormValue("alias")

		data.RLock()
		assetsPath := data.AssetsPath
		SizeOneKilobyte := data.SizeOneKilobyte
		data.RUnlock()

		InfoOfDevice := data.aliasToDevice(alias)
		// InfoOfDevice := data.getInfoOfDeviceFromMT(alias)
		// TempInfoOfDevice := data.aliasToDevice(alias)
		// InfoOfDevice.ShouldBeBlocked = TempInfoOfDevice.ShouldBeBlocked

		DisplayDataUser := DisplayDataUserType{
			Header:           "Редактирование пользователя",
			Copyright:        "GoSquidLogAnalyzer <i>© 2020</i> by Vladislav Vegner",
			Mail:             "mailto:vegner.vs@uttist.ru",
			Alias:            alias,
			SizeOneKilobyte:  SizeOneKilobyte,
			InfoOfDeviceType: InfoOfDevice,
		}

		fmap := template.FuncMap{
			"FormatSize": FormatSize,
		}
		t := template.Must(template.New("editalias").Funcs(fmap).ParseFiles(
			assetsPath+"/editalias.html",
			assetsPath+"/header.html",
			assetsPath+"/footer.html"))
		err := t.Execute(w, DisplayDataUser)
		if err != nil {
			fmt.Fprintf(w, "Что-то пошло не так, произошла ошибка при выполнении запроса:%v", err.Error())
		}
	} else if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, `Что-то пошло не так, произошла ошибка при выполнении запроса. 
			<br> %v
			<br> Перенаправление...
			<br> Если ничего не происходит нажмите <a href="/">сюда</a>`, err.Error())
			time.Sleep(5 * time.Second)
			http.Redirect(w, r, "/", 302)

		}
		params := r.Form
		alias := params["alias"][0]
		device := data.aliasToDevice(alias)
		parseParamertToDevice(&device, params)

		if err := data.setDevice(device); err != nil {
			fmt.Fprintf(w, `Произошла ошибка при сохранении. 
			<br> %v
			<br> Перенаправление...
			<br> Если ничего не происходит нажмите <a href="/">сюда</a>`, err.Error())
			time.Sleep(5 * time.Second)
			http.Redirect(w, r, "/", 302)
			return
		}

		data.updateInfoOfDeviceFromMT(alias)

		http.Redirect(w, r, "/", 302)
		log.Printf("%v(%v)%v", alias, device, params)
	}
}

func parseParamertToDevice(device *InfoOfDeviceType, params url.Values) {
	if len(params["TypeD"]) > 0 {
		device.TypeD = params["TypeD"][0]
	} else {
		device.TypeD = "other"
	}
	if len(params["name"]) > 0 {
		device.Name = params["name"][0]
	} else {
		device.Name = ""
	}
	if len(params["col"]) > 0 {
		device.Position = params["col"][0]
	} else {
		device.Position = ""
	}
	if len(params["com"]) > 0 {
		device.Company = params["com"][0]
	} else {
		device.Company = ""
	}
	if len(params["comment"]) > 0 {
		device.Comment = params["comment"][0]
	} else {
		device.Comment = ""
	}
	if len(params["disabled"]) > 0 {
		device.Disabled = paramertToBool(params["disabled"][0])
	} else {
		device.Disabled = false
	}
	if len(params["quotahourly"]) > 0 {
		device.HourlyQuota = paramertToUint(params["quotahourly"][0])
	} else {
		device.HourlyQuota = 0
	}
	if len(params["quotadaily"]) > 0 {
		device.DailyQuota = paramertToUint(params["quotadaily"][0])
	} else {
		device.DailyQuota = 0
	}
	if len(params["quotamonthly"]) > 0 {
		device.MonthlyQuota = paramertToUint(params["quotamonthly"][0])
	} else {
		device.MonthlyQuota = 0
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
	referURL := r.FormValue("refer")
	t.timer.Stop()
	t.timer.Reset(1 * time.Second)

	http.Redirect(w, r, "/"+referURL, 302)
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

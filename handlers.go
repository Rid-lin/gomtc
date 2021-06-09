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

	// http.HandleFunc("/", logreq(transport.handleIndex))
	// http.HandleFunc("/wf/", logreq(transport.handleWithFriends))
	http.HandleFunc("/", logreq(transport.handleIndex))
	http.HandleFunc("/wf/", logreq(transport.handleIndexWithFriends))
	http.HandleFunc("/log/", logreq(transport.handleLog))
	http.HandleFunc("/runparse", logreq(transport.handleRunParse))
	http.HandleFunc("/editalias/", logreq(transport.handleEditAlias))

	log.Infof("gomtc listens HTTP on:'%v'", cfg.ListenAddr)

	err := http.ListenAndServe(cfg.ListenAddr, nil)
	if err != nil {
		log.Fatal("http-server returned error:", err)
		transport.exitChan <- os.Kill
	}

}

func logreq(f func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("access:%s%s?%s", r.Host, r.URL.Path, r.URL.RawQuery)

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

// func (data *Transport) handleShowReport(w http.ResponseWriter, withfriends bool, preffix string, r *http.Request) {

// 	request := parseDataFromURL(r)
// 	request.referURL = r.Host + r.URL.Path
// 	request.path = r.URL.Path
// 	data.RLock()
// 	assetsPath := data.AssetsPath
// 	data.RUnlock()
// 	DisplayData := data.reportTrafficHourlyByLogins(request, withfriends)

// 	fmap := template.FuncMap{
// 		"FormatSize": FormatSize,
// 	}
// 	t := template.Must(template.New("index"+preffix).Funcs(fmap).ParseFiles(
// 		assetsPath+"/index"+preffix+".html",
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

func (data *Transport) handleIndex(w http.ResponseWriter, r *http.Request) {
	data.handleNewReport(w, false, "", r)
}

func (data *Transport) handleIndexWithFriends(w http.ResponseWriter, r *http.Request) {
	data.handleNewReport(w, true, "wf", r)
}

func (t *Transport) handleNewReport(w http.ResponseWriter, withfriends bool, preffix string, r *http.Request) {

	request := parseDataFromURL(r)
	request.referURL = r.Host + r.URL.Path
	request.path = r.URL.Path
	t.RLock()
	assetsPath := t.AssetsPath
	t.RUnlock()
	DisplayData, err := t.reportTrafficHourlyByLoginsNew(request, withfriends)
	if err != nil {
		fmt.Fprintf(w, "Проверьте налиие логов за запрашиваемый период<br> или подождите несколько минут.")

	}

	fmap := template.FuncMap{
		"FormatSize": FormatSize,
	}
	templ := template.Must(template.New("index"+preffix).Funcs(fmap).ParseFiles(
		assetsPath+"/index"+preffix+".html",
		assetsPath+"/header.html",
		assetsPath+"/footer.html"))
	err = templ.Execute(w, DisplayData)
	if err != nil {
		if strings.Contains(fmt.Sprint(err), "index out of range") {
			fmt.Fprintf(w, "Проверьте налиие логов за запрашиваемый период<br> или подождите несколько минут.")
		} else {
			fmt.Fprintf(w, "Что-то пошло не так, произошла ошибка при выполнении запроса. <br> %v", err.Error())
		}
	}
}

// func (data *Transport) handleIndex(w http.ResponseWriter, r *http.Request) {
// 	data.handleShowReport(w, false, "", r)
// }

// func (data *Transport) handleWithFriends(w http.ResponseWriter, r *http.Request) {
// 	data.handleShowReport(w, true, "wf", r)
// }

func (t *Transport) handleEditAlias(w http.ResponseWriter, r *http.Request) {
	t.RLock()
	assetsPath := t.AssetsPath
	SizeOneKilobyte := t.SizeOneKilobyte
	p := parseType{
		SSHCredentials:   t.sshCredentials,
		QuotaType:        t.QuotaType,
		BlockAddressList: t.BlockAddressList,
		Location:         t.Location,
	}
	t.RUnlock()

	if r.Method == "GET" {
		alias := r.FormValue("alias")
		aliasS := t.getAliasS(alias)
		// InfoOfDevice := data.aliasToDevice(alias)

		DisplayDataUser := DisplayDataUserType{
			Header:          "Редактирование пользователя",
			Copyright:       "GoSquidLogAnalyzer <i>© 2020</i> by Vladislav Vegner",
			Mail:            "mailto:vegner.vs@uttist.ru",
			SizeOneKilobyte: SizeOneKilobyte,
			InfoType:        aliasS,
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
		device := t.getAliasS(alias)
		// device := data.aliasToDevice(alias)
		fromFormToDevice(&device, params)
		// if err := data.setDevice(device); err != nil {
		// if err := devices.updateInfo(info.convertToDevice(quotaDef)); err != nil {
		// 	fmt.Fprintf(w, `Произошла ошибка при сохранении.
		// 	<br> %v
		// 	<br> Перенаправление...
		// 	<br> Если ничего не происходит нажмите <a href="/">сюда</a>`, err.Error())
		// 	time.Sleep(5 * time.Second)
		// 	http.Redirect(w, r, "/", 302)
		// 	return
		// }
		// _ = info.sendByAll(p, quotaDef)
		device.Update(p)
		t.Lock()
		t.devices[KeyDevice{ip: device.activeAddress, mac: device.activeMacAddress}] = device.DeviceType
		t.Unlock()

		// t.getDevices()
		var refer string
		if len(params["reffer"]) > 0 {
			refer = params["alias"][0]
		}
		http.Redirect(w, r, "/"+refer, 302)
		log.Printf("%v(%v)%v", alias, device, params)
	}
}

func fromFormToDevice(device *InfoType, params url.Values) {
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
	t.timerParse.Stop()
	t.timerParse.Reset(1 * time.Second)

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

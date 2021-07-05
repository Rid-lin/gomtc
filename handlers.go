package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	. "git.vegner.org/vsvegner/gomtc/internal/config"

	log "github.com/sirupsen/logrus"
)

func (t *Transport) handleRequest(cfg *Config) {
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(cfg.AssetsPath))))

	http.HandleFunc("/", logreq(t.handleIndex))
	http.HandleFunc("/wf/", logreq(t.handleIndexWithFriends))
	http.HandleFunc("/log/", logreq(t.handleLog))
	http.HandleFunc("/runparse", logreq(t.handleRunParse))
	http.HandleFunc("/editalias/", logreq(t.handleEditAlias))

	http.HandleFunc("/api/v1/blocked/", logreq(t.handleBlocked))
	http.HandleFunc("/api/v1/devices/", logreq(t.handleShowDevices))
	http.HandleFunc("/api/v1/aliases/", logreq(t.handleShowAliases))
	http.HandleFunc("/api/v1/getmac/", logreq(t.handleGetMac))

	log.Infof("gomtc listens HTTP on:'%v'", cfg.ListenAddr)

	err := http.ListenAndServe(cfg.ListenAddr, nil)
	if err != nil {
		log.Fatal("http-server returned error:", err)
		t.exitChan <- os.Kill
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
		request.dateFrom = time.Now().In(Location).Format(DateLayout)
	}
	if len(m["date_to"]) > 0 {
		request.dateTo = m["date_to"][0]
	} else {
		request.dateTo = time.Now().Add(24 * time.Hour).Format(DateLayout)
	}
	if request.dateFrom == "" {
		request.dateFrom = time.Now().In(Location).Format(DateLayout)
	}
	if request.dateTo == "" {
		request.dateTo = time.Now().Add(24 * time.Hour).Format(DateLayout)
	}
	if len(m["report"]) > 0 {
		request.report = m["report"][0]
	}
	var dateFrom, dateTo time.Time
	if len(m["direct"]) > 0 {
		dateFrom, err = time.ParseInLocation(DateLayout, request.dateFrom, Location)
		if err != nil {
			dateFrom = time.Now()
		}
		dateTo, err = time.ParseInLocation(DateLayout, request.dateFrom, Location)
		if err != nil {
			dateTo = time.Now()
		}
		if m["direct"][0] == ">" {
			dateFrom = dateFrom.AddDate(0, 0, 1)
		} else if m["direct"][0] == "<" {
			dateFrom = dateFrom.AddDate(0, 0, -1)
		}
		request.dateFrom = dateFrom.In(Location).Format(DateLayout)
	}
	if len(m["direct_to"]) > 0 {
		dateFrom, err = time.Parse(DateLayout, request.dateFrom)
		if err != nil {
			dateFrom = time.Now()
		}
		dateTo, err = time.Parse(DateLayout, request.dateFrom)
		if err != nil {
			dateTo = time.Now()
		}
		if m["direct_to"][0] == ">" {
			dateTo = dateTo.AddDate(0, 0, 1)
		} else if m["direct_to"][0] == "<" {
			dateTo = dateTo.AddDate(0, 0, -1)
		}
		request.dateTo = dateTo.In(Location).Format(DateLayout)
	}
	return request
}

func (t *Transport) handleIndex(w http.ResponseWriter, r *http.Request) {
	t.handleNewReport(w, false, r)
}

func (data *Transport) handleIndexWithFriends(w http.ResponseWriter, r *http.Request) {
	data.handleNewReport(w, true, r)
}

func (t *Transport) handleNewReport(w http.ResponseWriter, withfriends bool, r *http.Request) {
	t.RLock()
	assetsPath := t.AssetsPath
	t.RUnlock()

	request := parseDataFromURL(r)
	request.referURL = r.Host + r.URL.Path
	request.path = r.URL.Path
	DisplayData, err := t.reportDailyHourlyByMac(request, withfriends)
	if err != nil {
		fmt.Fprintf(w, "Проверьте налиие логов за запрашиваемый период<br> или подождите несколько минут.")

	}

	fmap := template.FuncMap{
		"FormatSize": FormatSize,
	}
	templ := template.Must(template.New("index").Funcs(fmap).ParseFiles(
		assetsPath+"/index.html",
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

func (t *Transport) handleEditAlias(w http.ResponseWriter, r *http.Request) {
	t.RLock()
	assetsPath := t.AssetsPath
	SizeOneKilobyte := t.SizeOneKilobyte
	t.RUnlock()

	if r.Method == "GET" {
		alias := r.FormValue("alias")
		aliasS := t.Aliases[alias]

		DisplayDataUser := DisplayDataUserType{
			Header:          "Редактирование устройства",
			Copyright:       "GoSquidLogAnalyzer <i>© 2020</i> by Vladislav Vegner",
			Mail:            "mailto:vegner.vs@uttist.ru",
			SizeOneKilobyte: SizeOneKilobyte,
			InfoType: InfoType{
				InfoName:   aliasS.AliasName,
				PersonType: aliasS.PersonType,
				QuotaType:  aliasS.QuotaType,
				DeviceType: t.devices[aliasS.KeyArr[0]],
			},
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
		aliasName := params["alias"][0]
		alias := t.Aliases[aliasName]
		alias.UpdateFromForm(params)

		var refer string
		if len(params["reffer"]) > 0 {
			refer = params["alias"][0]
		}
		http.Redirect(w, r, "/"+refer, 302)
		log.Tracef("%v(%v)%v", aliasName, alias, params)
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
		// TODO Поменть на выгрузку из SQL
		// TODO Дописать метод записи в SQL
		// Logs:   t.logs,
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

func (t *Transport) handleBlocked(w http.ResponseWriter, r *http.Request) {
	var arr []string
	t.Lock()
	for key, d := range t.Aliases {
		if d.ShouldBeBlocked {
			arr = append(arr, key)
		}
	}
	t.Unlock()
	// Just send out the JSON version of 'arr'
	// j, _ := json.Marshal(arr)
	// _, _ = w.Write(j)
	renderJSON(w, arr)
}

func (t *Transport) handleShowDevices(w http.ResponseWriter, r *http.Request) {
	var arr []DeviceType
	t.Lock()
	for _, d := range t.devices {
		arr = append(arr, d)
	}
	t.Unlock()
	// Just send out the JSON version of 'arr'
	// j, _ := json.MarshalIndent(arr, "\t", "")
	// _, _ = w.Write(j)
	renderJSON(w, arr)
}

func (t *Transport) handleGetMac(w http.ResponseWriter, r *http.Request) {
	params := r.FormValue("ip")
	var arr []ResponseType
	t.Lock()
	for _, d := range t.devices {
		rt := ResponseType{
			IP:       d.ActiveAddress,
			Mac:      d.ActiveMacAddress,
			Comments: d.Comment,
			HostName: d.HostName,
		}
		if d.ActiveAddress == params && params != "" {
			j, _ := json.MarshalIndent(rt, "\t", "")
			_, _ = w.Write(j)
			return
		}
		arr = append(arr, rt)
	}
	t.Unlock()
	if params != "" {
		_, _ = w.Write([]byte{})
		return
	}
	// Just send out the JSON version of 'arr'
	// j, _ := json.MarshalIndent(arr, "\t", "")
	// _, _ = w.Write(j)
	renderJSON(w, arr)
}

func (t *Transport) handleShowAliases(w http.ResponseWriter, r *http.Request) {
	var arr []AliasType
	t.Lock()
	for _, d := range t.Aliases {
		arr = append(arr, d)
	}
	t.Unlock()
	// Just send out the JSON version of 'arr'
	// j, _ := json.MarshalIndent(arr, "\t", "")
	// _, _ = w.Write(j)
	renderJSON(w, arr)
}

// renderJSON преобразует 'v' в формат JSON и записывает результат, в виде ответа, в w.
func renderJSON(w http.ResponseWriter, v interface{}) {
	js, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(js)
}

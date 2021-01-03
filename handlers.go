package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"net/url"
	"time"

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

type LineOfDisplay struct {
	Alias,
	Login,
	IP string
	Size float64
	IDIP,
	IDLogin int32
	HourSize [24]float64
}

type KeyForLineOfTraffic struct {
	Login   string
	IDLogin int
	Hour    int
}

type TrafficOneHoursLoginsType map[KeyForLineOfTraffic]float64

type Dysplaydata struct {
	ArrayDysplay []LineOfDisplay
	TitleReport  string
	DateFrom     string
	DateTo       string
}

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
	indextmpl, err := template.ParseFiles(cfg.AssetsPath+"/index.html", cfg.AssetsPath+"/header.html", cfg.AssetsPath+"/news.html", cfg.AssetsPath+"/footer.html")
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

	indextmpl, err := template.ParseFiles(cfg.AssetsPath+"/index.html", cfg.AssetsPath+"/header.html", cfg.AssetsPath+"/menu.html", cfg.AssetsPath+"/right.html", cfg.AssetsPath+"/footer.html")
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
	var dateFrom, dateTo string
	u, err := url.Parse(r.URL.String())
	if err != nil {
		log.Error(err)
	}
	m, _ := url.ParseQuery(u.RawQuery)

	// Checking the availability of data from the URL. To show today if there is no data.
	if len(m["date_from"]) > 0 {
		dateFrom = m["date_from"][0]
	} else {
		dateFrom = time.Now().Format("2006-01-02")
	}
	if len(m["date_to"]) > 0 {
		dateTo = m["date_to"][0]
	} else {
		dateTo = time.Now().Add(24 * time.Hour).Format("2006-01-02")
	}
	if dateFrom == "" {
		dateFrom = time.Now().Format("2006-01-02")
	}
	if dateTo == "" {
		dateTo = time.Now().Add(24 * time.Hour).Format("2006-01-02")
	}

	log.Debug("dateFrom=", dateFrom)
	log.Debug("dateTo=", dateTo)

	rows, err := data.db.Query(fmt.Sprint(`
	SELECT 
	COALESCE(scsq_log.name, scsq_ip.name,"N/A"), 
	COALESCE(scsq_ip.name, scsq_log.name,"N/A"), 
	SUM(scsq_traf.sizeinbytes) AS 's', 
	COALESCE(sa.name, scsq_log.name, scsq_ip.name,"N/A"), 
	scsq_log.id, 
	scsq_ip.id 
	FROM (SELECT scsq_traffic.id, scsq_traffic.login, scsq_traffic.ipaddress 
	  FROM scsq_traffic
		LEFT OUTER JOIN (SELECT scsq_logins.id, name FROM scsq_logins WHERE id IN ("")) AS tmplogin 
		  ON tmplogin.id=scsq_traffic.login
		  LEFT OUTER JOIN (SELECT scsq_ipaddress.id, name FROM scsq_ipaddress WHERE id IN ("")) AS tmpipaddress 
			ON tmpipaddress.id=scsq_traffic.ipaddress
	  WHERE date>UNIX_TIMESTAMP(STR_TO_DATE('`, dateFrom, `', '%Y-%m-%d')) 
      AND date<UNIX_TIMESTAMP(STR_TO_DATE('`, dateTo, `', '%Y-%m-%d'))  
      AND tmplogin.id is NULL 
      AND tmpipaddress.id IS NULL
    ) AS tmp
    INNER JOIN scsq_traffic as scsq_traf on scsq_traf.id=tmp.id
    LEFT OUTER JOIN scsq_alias sa ON sa.tableid=tmp.login and sa.typeid=0
    INNER JOIN scsq_logins as scsq_log on scsq_log.id=tmp.login
    INNER JOIN scsq_ipaddress as scsq_ip on scsq_ip.id=tmp.ipaddress
	GROUP BY scsq_log.name,
	scsq_ip.name
    	ORDER BY s DESC;`))
	if err != nil {
		log.Error(err)
	}
	defer rows.Close()
	var size int
	ReportData := []LineOfDisplay{}
	for rows.Next() {
		line := LineOfDisplay{}
		err := rows.Scan(&line.Login, &line.IP, &size, &line.Alias, &line.IDLogin, &line.IDIP)
		if err != nil {
			log.Error(err)
			continue
		}
		line.Size = math.Round((float64(size)/(1024.*1024.))*100) / 100
		ReportData = append(ReportData, line)
	}

	rows, err = data.db.Query(fmt.Sprint(`
	SELECT  login, nofriends.name, sum(sizeinbytes),FROM_UNIXTIME(date,'%k') d
	FROM scsq_quicktraffic LEFT JOIN (SELECT id, name FROM scsq_logins) AS nofriends 
		ON scsq_quicktraffic.login=nofriends.id  
		LEFT OUTER JOIN (SELECT id FROM scsq_logins WHERE id IN ("")) AS tmplogin 
		  ON tmplogin.id=scsq_quicktraffic.login
		LEFT OUTER JOIN (SELECT id FROM scsq_ipaddress WHERE id IN ("")) AS tmpipaddress 
		ON tmpipaddress.id=scsq_quicktraffic.ipaddress
		WHERE date>UNIX_TIMESTAMP(STR_TO_DATE('`, dateFrom, `', '%Y-%m-%d'))  
	  AND date<UNIX_TIMESTAMP(STR_TO_DATE('`, dateTo, `', '%Y-%m-%d'))
	  AND tmplogin.id is  NULL 
	  AND tmpipaddress.id is  NULL
	  AND site NOT IN ("")
	  AND par=1
	GROUP BY login, d, nofriends.name
	ORDER BY nofriends.name;`))
	if err != nil {
		log.Error(err)
	}
	defer rows.Close()

	TrafficByHoursLogins := TrafficOneHoursLoginsType{}
	var (
		idLogin, sizeHour, hour int
		login                   string
	)
	for rows.Next() {
		err := rows.Scan(&idLogin, &login, &sizeHour, &hour)
		if err != nil {
			log.Error(err)
			continue
		}
		key := &KeyForLineOfTraffic{
			Login:   login,
			IDLogin: idLogin,
			Hour:    hour,
		}
		TrafficByHoursLogins[*key] = math.Round((float64(sizeHour)/(1024.*1024.))*100) / 100
	}

	for index := range ReportData {
		for jIndex := range ReportData[index].HourSize {
			key := &KeyForLineOfTraffic{
				Login:   ReportData[index].Login,
				IDLogin: int(ReportData[index].IDLogin),
				Hour:    jIndex,
			}
			ReportData[index].HourSize[jIndex] = TrafficByHoursLogins[*key]
		}

	}

	Dysplaydata := &Dysplaydata{
		ArrayDysplay: ReportData,
		TitleReport:  "Отчёт по трафику пользователей с логинами и IP-адресами",
		DateFrom:     dateFrom,
		DateTo:       dateTo,
	}

	// Starting template processing to display the page in the browser
	indextmpl, err := template.ParseFiles(cfg.AssetsPath+"/flow.html", cfg.AssetsPath+"/header.html", cfg.AssetsPath+"/menu.html", cfg.AssetsPath+"/report.html", cfg.AssetsPath+"/footer.html")
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	err = indextmpl.ExecuteTemplate(w, "flow", Dysplaydata)
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

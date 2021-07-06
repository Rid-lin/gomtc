package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"

	"git.vegner.org/vsvegner/gomtc/internal/app/model"
	v "git.vegner.org/vsvegner/gomtc/internal/app/validation"
)

func (ds *DevicesType) parseLeasePrintAsValue(b bytes.Buffer) {
	var d model.DeviceType
	var addedTo string
	inputStr := b.String()
	arr := strings.Split(inputStr, ";")
	// For Debug
	_ = saveStrToFile(".config/str.temp", inputStr)
	_ = saveArrToFile(".config/arr.temp", arr)
	for _, lineItem := range arr {
		switch {
		case v.IsParametr(lineItem, ".id"):
			d.Id = v.ParseParamertToStr(lineItem)
		case v.IsParametr(lineItem, "active-address"):
			d.ActiveAddress = v.ParseParamertToStr(lineItem)
		case v.IsParametr(lineItem, "address"):
			d.Address = v.ParseParamertToStr(lineItem)
		case v.IsParametr(lineItem, "allow-dual-stack-queue"):
			d.AllowDualStackQueue = v.ParseParamertToStr(lineItem)
		case v.IsParametr(lineItem, "client-id"):
			addedTo = "client-id"
			d.ClientId = v.ParseParamertToStr(lineItem)
		case v.IsParametr(lineItem, "disabled"):
			d.DisabledL = v.ParseParamertToStr(lineItem)
		case v.IsParametr(lineItem, "insert-queue-before"):
			d.InsertQueueBefore = v.ParseParamertToStr(lineItem)
		case v.IsParametr(lineItem, "radius"):
			d.Radius = v.ParseParamertToStr(lineItem)
		case v.IsParametr(lineItem, "active-client-id"):
			d.ActiveClientId = v.ParseParamertToStr(lineItem)
		case v.IsParametr(lineItem, "address-lists"):
			addedTo = "address-lists"
			d.AddressLists = v.ParseParamertToStr(lineItem)
		case v.IsParametr(lineItem, "always-broadcast"):
			d.AlwaysBroadcast = v.ParseParamertToStr(lineItem)
		case v.IsParametr(lineItem, "dynamic"):
			d.Dynamic = v.ParseParamertToStr(lineItem)
		case v.IsParametr(lineItem, "last-seen"):
			d.LastSeen = v.ParseParamertToStr(lineItem)
		case v.IsParametr(lineItem, "rate-limit"):
			d.RateLimit = v.ParseParamertToStr(lineItem)
		case v.IsParametr(lineItem, "use-src-mac"):
			d.UseSrcMac = v.ParseParamertToStr(lineItem)
		case v.IsParametr(lineItem, "active-mac-address"):
			d.ActiveMacAddress = v.ParseParamertToStr(lineItem)
		case v.IsParametr(lineItem, "agent-circuit-id"):
			d.AgentCircuitId = v.ParseParamertToStr(lineItem)
		case v.IsParametr(lineItem, "block-access"):
			d.BlockAccess = v.ParseParamertToStr(lineItem)
		case v.IsParametr(lineItem, "dhcp-option"):
			d.DhcpOption = v.ParseParamertToStr(lineItem)
		case v.IsParametr(lineItem, "expires-after"):
			d.ExpiresAfter = v.ParseParamertToStr(lineItem)
		case v.IsParametr(lineItem, "lease-time"):
			d.LeaseTime = v.ParseParamertToStr(lineItem)
		case v.IsParametr(lineItem, "server"):
			d.Server = v.ParseParamertToStr(lineItem)
		case v.IsParametr(lineItem, "active-server"):
			d.ActiveServer = v.ParseParamertToStr(lineItem)
		case v.IsParametr(lineItem, "agent-remote-id"):
			d.AgentRemoteId = v.ParseParamertToStr(lineItem)
		// case v.IsParametr(lineItem, "blocked"):
		// 	d.Blocked = v.ParseParamertToStr(lineItem)
		case v.IsParametr(lineItem, "dhcp-option-set"):
			d.DhcpOptionSet = v.ParseParamertToStr(lineItem)
		case v.IsParametr(lineItem, "host-name"):
			d.HostName = v.ParseParamertToStr(lineItem)
		case v.IsParametr(lineItem, "mac-address"):
			d.MacAddress = v.ParseParamertToStr(lineItem)
		case v.IsParametr(lineItem, "src-mac-address"):
			d.SrcMacAddress = v.ParseParamertToStr(lineItem)
		case v.IsComment(lineItem, "comment"):
			d.Comment = v.ParseParamertToComment(lineItem)
		case v.IsParametr(lineItem, "status"):
			d.Status = v.ParseParamertToStr(lineItem)
			*ds = append(*ds, d)
			d = model.DeviceType{}
		case addedTo == "address-lists":
			d.AddressLists = d.AddressLists + "," + lineItem
		}
	}
}

func saveArrToFile(nameFile string, arr []string) error {
	f, _ := os.Create(nameFile)
	defer f.Close()
	w := bufio.NewWriter(f)
	for index := 0; index < len(arr)-1; index++ {
		fmt.Fprintln(w, arr[index])
	}
	w.Flush()
	return nil
}

func saveStrToFile(nameFile, str string) error {
	f, _ := os.Create(nameFile)
	defer f.Close()
	_, _ = f.WriteString(str)
	return nil
}

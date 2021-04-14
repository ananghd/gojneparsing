package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type TrackingDetail struct {
	AWBNo        string
	Services     string
	ShipmentDate string
	Origin       string
	Destination  string
}

type Shipper struct {
	Shipper   string
	Consignee string
}

type FormatedDate struct {
	CreatedAt string `json:"createdAt"`
}

type LogHistory struct {
	CreatedAt   string       `json:"createdAt"`
	Description string       `json:"description"`
	Formatted   FormatedDate `json:"formatted"`
}

type DataTracking struct {
	ReceivedBy string       `json:"receivedBy"`
	Histories  []LogHistory `json:"histories"`
}

type ResponseStatus struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Data    DataTracking `json:"data"`
}

func getLocalTimezone() {
	t := time.Now()
	zone, offset := t.Zone()
	fmt.Println(zone, offset)
}

func main() {
	var URL = "https://gist.githubusercontent.com/nubors/eecf5b8dc838d4e6cc9de9f7b5db236f/raw/d34e1823906d3ab36ccc2e687fcafedf3eacfac9/jne-awb.html"
	res, err := http.Get(URL)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	rows := make([]LogHistory, 0)
	var rowData []string
	loc, _ := time.LoadLocation("Asia/Jakarta")

	monthTranslate := strings.NewReplacer(
		"January", "Januari",
		"February", "Februari",
		"March", "Maret",
		"April", "April",
		"May", "Mei",
		"June", "Juni",
		"July", "Juli",
		"August", "Agustus",
		"September", "September",
		"October", "Oktober",
		"November", "November",
		"December", "Desember")

	var deliverdTemplate = "DELIVERED TO"
	var isDelivered = false
	var deliveredBy = ""
	//var deliveredDate = ""

	doc.Find(".tracking tbody").Children().Each(func(i int, sel *goquery.Selection) {
		if i < 3 {
			return
		}
		sel.Find("td").Each(func(indexth int, tablecell *goquery.Selection) {
			rowData = append(rowData, tablecell.Text())
		})
		createdAt, _ := time.ParseInLocation("02-01-2006 15:04", rowData[0], loc)

		row := new(LogHistory)
		row.CreatedAt = createdAt.Format(time.RFC3339)
		row.Description = rowData[1]
		row.Formatted.CreatedAt = monthTranslate.Replace(createdAt.Format("02 January 2006, 15:04 MST"))
		rows = append(rows, *row)
		rowData = nil

		// Delivered status checking
		if isDelivered {
			return
		}
		isDelivered = strings.HasPrefix(row.Description, deliverdTemplate)
		if isDelivered {
			//deliveredDate = row.CreatedAt
			var delivery = strings.Split(row.Description, "|")
			var deliveryName = strings.Split(delivery[0], "[")
			deliveredBy = deliveryName[1]
		}
	})

	response := new(ResponseStatus)
	response.Code = "060101"
	response.Message = "Delivery tracking detail fetched successfully"
	response.Data.ReceivedBy = deliveredBy
	response.Data.Histories = rows
	b, err := json.MarshalIndent(response, "", " ")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(b))
}

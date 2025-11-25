package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/oschwald/geoip2-golang"
	"github.com/ringsaturn/tzf"
)

type Response struct {
	IP       string `json:"ip"`
	Timezone string `json:"timezone"`
	City     string `json:"city"`
	Country  string `json:"country"`
}

var hits int = 0

func main() {
	// 1. Load GeoIP DB
	db, err := geoip2.Open("./dbip-city-lite.mmdb")
	if err != nil {
		log.Fatal("DB Error: ", err)
	}
	defer db.Close()

	// 2. Load Timezone Finder (Embeds compressed polygon data)
	tzFinder, err := tzf.NewDefaultFinder()
	if err != nil {
		log.Fatal("TZF Error: ", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hits++
		log.Printf("[%d] Request at %s", hits, time.Now().Format(time.DateTime))

		// Get IP
		ipStr := r.Header.Get("X-Real-IP")
		if ipStr == "" {
			ipStr = r.Header.Get("X-Forwarded-For")
		}
		if ipStr == "" {
			ipStr, _, _ = net.SplitHostPort(r.RemoteAddr)
		}
		if strings.Contains(ipStr, ",") {
			ipStr = strings.TrimSpace(strings.Split(ipStr, ",")[0])
		}

		ip := net.ParseIP(ipStr)
		if ip == nil {
			http.Error(w, "Invalid IP", 400)
			return
		}

		// Lookup GeoIP
		record, err := db.City(ip)
		if err != nil {
			http.Error(w, "Lookup failed", 500)
			return
		}

		// Derive Timezone from Lat/Lon using TZF
		// DB-IP Lite lacks the string, but has the coordinates.
		tz := record.Location.TimeZone
		if tz == "" && record.Location.Latitude != 0 {
			tz = tzFinder.GetTimezoneName(record.Location.Longitude, record.Location.Latitude)
		}

		resp := Response{
			IP:       ip.String(),
			Timezone: tz,
			City:     record.City.Names["en"],
			Country:  record.Country.Names["en"],
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	log.Println("Listening on :7515")
	log.Fatal(http.ListenAndServe(":7515", nil))
}

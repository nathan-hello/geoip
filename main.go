package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/oschwald/geoip2-golang"
)

type Response struct {
	IP       string `json:"ip"`
	Timezone string `json:"timezone"`
	City     string `json:"city"`
	Country  string `json:"country"`
}

func main() {
	// Point to the DB-IP file we downloaded
	db, err := geoip2.Open("./dbip-city-lite.mmdb")
	if err != nil {
		log.Fatal("Could not open database: ", err)
	}
	defer db.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Get IP from Header (if behind proxy) or RemoteAddr
		ipStr := r.Header.Get("X-Forwarded-For")
		if ipStr == "" {
			ipStr, _, _ = net.SplitHostPort(r.RemoteAddr)
		}
		
		// Handle comma-separated headers
		if strings.Contains(ipStr, ",") {
			ipStr = strings.TrimSpace(strings.Split(ipStr, ",")[0])
		}

		ip := net.ParseIP(ipStr)
		if ip == nil {
			http.Error(w, "Invalid IP", 400)
			return
		}

		// Lookup
		record, err := db.City(ip)
		if err != nil {
			http.Error(w, "Lookup failed", 500)
			return
		}

		resp := Response{
			IP:       ip.String(),
			Timezone: record.Location.TimeZone,
			City:     record.City.Names["en"],
			Country:  record.Country.Names["en"],
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

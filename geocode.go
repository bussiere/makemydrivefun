package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/codingsince1985/geo-golang/openstreetmap"
	"github.com/schollz/jsonstore"
)

var cities *jsonstore.JSONStore

func init() {
	var err error
	cities, err = jsonstore.Open("static/data/cities.json")
	if err != nil {
		panic(err)
	}
	openstreetmap.Geocoder()
}

func Geocode(s string) (p Point, err error) {
	s = strings.ToLower(s)
	err = cities.Get(s, &p)
	if err == nil {
		return
	}
	location, err := openstreetmap.Geocoder().Geocode(s)
	if err != nil {
		return
	}
	if location == nil {
		err = errors.New("'" + s + "' is not a recognized address")
		return
	}
	p = Point{
		Longitude: location.Lng,
		Latitude:  location.Lat,
	}
	cities.Set(s, p)
	jsonstore.Save(cities, "static/data/cities.json")
	return
}

func LocationFromIP(ip string) (location string, err error) {
	type ResultJSON struct {
		IP          string  `json:"ip"`
		CountryCode string  `json:"country_code"`
		CountryName string  `json:"country_name"`
		RegionCode  string  `json:"region_code"`
		RegionName  string  `json:"region_name"`
		City        string  `json:"city"`
		ZipCode     string  `json:"zip_code"`
		TimeZone    string  `json:"time_zone"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
		MetroCode   int     `json:"metro_code"`
	}
	resp, err := http.Get(geoipServer + "/json/" + ip)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var result ResultJSON
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return
	}
	location = fmt.Sprintf("%s, %s, %s", result.City, result.RegionName, result.CountryName)
	if len(location) < 5 {
		location = ""
	}
	return
}

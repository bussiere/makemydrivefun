package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
)

type Point struct {
	Longitude float64
	Latitude  float64
}

var basicCities map[string]Point

func init() {
	b, err := ioutil.ReadFile("./static/data/basic_cities.json")
	if err != nil {
		panic(err)
	}
	type CityData struct {
		Address string
		Lat     float64
		Lon     float64
	}
	var basicCityData map[string]CityData
	err = json.Unmarshal(b, &basicCityData)
	basicCities = make(map[string]Point)
	for k, v := range basicCityData {
		basicCities[k] = Point{
			Longitude: v.Lon,
			Latitude:  v.Lat,
		}
	}
}

const kmtomiles = float64(0.621371192)
const earthRadius = float64(6371)

// Distance returns the distance between two points in miles
func (pointA Point) DistanceAsCrowFlies(pointB Point) (distance float64) {
	var deltaLat = (pointB.Latitude - pointA.Latitude) * (math.Pi / 180)
	var deltaLon = (pointB.Longitude - pointA.Longitude) * (math.Pi / 180)

	var a = math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(pointA.Latitude*(math.Pi/180))*math.Cos(pointB.Latitude*(math.Pi/180))*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	var cp = 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance = earthRadius * cp
	return
}

// NearestCity returns the nearest city to that point
func (pointA Point) NearestCityWithin(miles float64) string {
	closest := 100000.0
	closestCity := ""
	for cityName, pointB := range basicCities {
		distance := pointA.DistanceAsCrowFlies(pointB)
		if distance <= miles && distance <= closest {
			closestCity = cityName
			closest = distance
		}
	}
	return closestCity
}

// DrivingRouteTo determines the route from point A to point B
func (pointA Point) DrivingRouteTo(pointB Point) (route OSRMResponse, err error) {
	url := fmt.Sprintf(osrmServer+"/route/v1/driving/%2.10f,%2.10f;%2.10f,%2.10f?geometries=geojson&steps=true", pointA.Longitude, pointA.Latitude, pointB.Longitude, pointB.Latitude)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&route)
	return
}

// OSRMResponse is the response to
// http://192.168.0.17:5000/route/v1/driving/-122.6765,45.5231;-113.4909,54.5444?geometries=geojson
type OSRMResponse struct {
	Code   string `json:"code"`
	Routes []struct {
		Geometry struct {
			Coordinates [][]float64 `json:"coordinates"`
			Type        string      `json:"type"`
		} `json:"geometry"`
		Legs []struct {
			Steps []struct {
				Intersections []struct {
					Out      float64   `json:"out"`
					Entry    []bool    `json:"entry"`
					Bearings []float64 `json:"bearings"`
					Location []float64 `json:"location"`
				} `json:"intersections"`
				Geometry struct {
					Coordinates [][]float64 `json:"coordinates"`
					Type        string      `json:"type"`
				} `json:"geometry"`
				Mode     string  `json:"mode"`
				Duration float64 `json:"duration"`
				Maneuver struct {
					BearingAfter  float64   `json:"bearing_after"`
					Location      []float64 `json:"location"`
					BearingBefore float64   `json:"bearing_before"`
					Type          string    `json:"type"`
				} `json:"maneuver"`
				Weight       float64 `json:"weight"`
				Distance     float64 `json:"distance"`
				Name         string  `json:"name"`
				Ref          string  `json:"ref,omitempty"`
				Destinations string  `json:"destinations,omitempty"`
			} `json:"steps"`
			Distance float64 `json:"distance"`
			Duration float64 `json:"duration"`
			Summary  string  `json:"summary"`
			Weight   float64 `json:"weight"`
		} `json:"legs"`
		Distance   float64 `json:"distance"`
		Duration   float64 `json:"duration"`
		WeightName string  `json:"weight_name"`
		Weight     float64 `json:"weight"`
	} `json:"routes"`
	Waypoints []struct {
		Hint     string    `json:"hint"`
		Name     string    `json:"name"`
		Location []float64 `json:"location"`
	} `json:"waypoints"`
}

// Duration returns the duration in minutes
func (r OSRMResponse) Duration() float64 {
	return r.Routes[0].Duration * 0.000277778 * 60
}

// Distance returns the distance in miles
func (r OSRMResponse) Distance() float64 {
	return r.Routes[0].Distance * 0.000621371
}

// Steps returns all the steps
func (r OSRMResponse) Steps() (route []Point) {
	route = make([]Point, 10000)
	i := 0
	for _, step := range r.Routes[0].Legs[0].Steps {
		for _, intersection := range step.Intersections {
			route[i] = Point{
				Longitude: intersection.Location[0],
				Latitude:  intersection.Location[1],
			}
			i++
		}
	}
	return route[:i]
}

// MajorSteps returns all the major steps
func (r OSRMResponse) MajorSteps() (route []Point) {
	route = make([]Point, len(r.Routes[0].Geometry.Coordinates))
	for i, coord := range r.Routes[0].Geometry.Coordinates {
		route[i] = Point{
			Longitude: coord[0],
			Latitude:  coord[1],
		}
	}
	return
}

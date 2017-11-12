package main

import (
	"encoding/json"
	"io/ioutil"
)

type GeoJSONPoints struct {
	Type     string                 `json:"type"`
	Features []GeoJSONPointsFeature `json:"features"`
}

type GeoJSONPointsFeature struct {
	Type       string `json:"type"`
	Properties struct {
		MarkerColor          string `json:"marker-color"`
		MarkerSize           string `json:"marker-size"`
		MarkerSymbol         string `json:"marker-symbol"`
		Desc                 string `json:"desc"`
		URL                  string `json:"url"`
		Title                string `json:"title"`
		Address              string `json:"address"`
		Fulltitle            string `json:"fulltitle"`
		Rating               string `json:"rating"`
		ID                   string `json:"id"`
		Distance             string `json:"distance"`
		DistanceFromLastCity string `json:"distance-from-last-city"`
		CityThatIsClose      string `json:"city-that-is-close"`
	} `json:"properties"`
	Geometry struct {
		Type        string    `json:"type"`
		Coordinates []float64 `json:"coordinates"` // Lon and Lat
	} `json:"geometry"`
}

func (g GeoJSONPointsFeature) String() string {
	b, _ := json.MarshalIndent(g, " ", " ")
	return string(b)
}

func LoadGeoJSONFile(filename string) (g GeoJSONPoints, err error) {
	var b []byte
	b, err = ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	err = json.Unmarshal(b, &g)
	return
}

func GeoJSONLineStringFeature(points []Point) string {
	type GeoJSONLineString struct {
		Type       string `json:"type"`
		Properties struct {
		} `json:"properties"`
		Geometry struct {
			Type        string      `json:"type"`
			Coordinates [][]float64 `json:"coordinates"`
		} `json:"geometry"`
	}

	var g GeoJSONLineString
	g.Type = "Feature"
	g.Geometry.Type = "LineString"
	g.Geometry.Coordinates = make([][]float64, len(points))
	for i, point := range points {
		g.Geometry.Coordinates[i] = []float64{point.Longitude, point.Latitude}
	}
	b, _ := json.MarshalIndent(g, " ", " ")
	return string(b)
}

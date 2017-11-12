package main

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"path"
	"strconv"
	"strings"

	humanize "github.com/dustin/go-humanize"
	"github.com/schollz/progressbar"
)

var allFeatures GeoJSONPoints

type RoadTrip struct {
	OriginName             string
	DestinationName        string
	Origin                 Point
	Destination            Point
	TotalCrowDistance      float64
	TotalDrivingDistance   string
	TotalDrivingTime       string
	MainRoute              OSRMResponse
	NearbyFeatures         []GeoJSONPointsFeature
	MaximumMinutesOffRoute float64
}

func (r RoadTrip) Name() string {
	s := strings.Replace(fmt.Sprintf("route%2.9f%2.9f%2.9f%2.9f", r.Origin.Longitude, r.Origin.Latitude, r.Destination.Longitude, r.Destination.Latitude), ".", "", -1)
	h := fnv.New32a()
	h.Write([]byte(s))
	return randomAlliterateCombo(int64(h.Sum32()))
}

func (r RoadTrip) Save() error {
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path.Join("cache", r.Name()+".json"), b, 0644)
}

func Load(origin, destination Point) (r *RoadTrip, err error) {
	r = new(RoadTrip)
	r.Origin = origin
	r.Destination = destination
	return LoadFromName(r.Name())
}

func LoadFromName(s string) (r *RoadTrip, err error) {
	b, err := ioutil.ReadFile(path.Join("cache", s+".json"))
	if err != nil {
		return
	}
	err = json.Unmarshal(b, &r)
	return
}

func init() {
	var err error
	allFeatures, err = LoadGeoJSONFile("static/data/features.geojson")
	if err != nil {
		panic(err)
	}
}
func New(origin, destination Point) (r *RoadTrip, err error) {
	r = new(RoadTrip)
	r.Origin = origin
	r.Destination = destination
	r.OriginName = r.Origin.NearestCityWithin(1000)
	r.DestinationName = r.Destination.NearestCityWithin(1000)
	r.TotalCrowDistance = origin.DistanceAsCrowFlies(destination)
	r.MaximumMinutesOffRoute = 20.0
	if err != nil {
		return
	}

	r.MainRoute, err = r.Origin.DrivingRouteTo(r.Destination)
	if err != nil {
		return
	}
	r.TotalDrivingDistance = humanize.Comma(int64(r.MainRoute.Distance())) + " miles"
	if r.MainRoute.Duration() < 60 {
		r.TotalDrivingTime = fmt.Sprintf("%d minutes", int(r.MainRoute.Duration()))
	} else {
		r.TotalDrivingTime = fmt.Sprintf("%d hours", int(r.MainRoute.Duration()/60))
	}
	return
}

func (r *RoadTrip) FindNearbyFeatures() {
	// filter only features between origin and destination
	possibleNearbyFeature := make(map[int]struct{})
	for featureI, feature := range allFeatures.Features {
		featurePoint := Point{
			Longitude: feature.Geometry.Coordinates[0],
			Latitude:  feature.Geometry.Coordinates[1],
		}
		if r.Origin.DistanceAsCrowFlies(featurePoint) < r.TotalCrowDistance && r.Destination.DistanceAsCrowFlies(featurePoint) < r.TotalCrowDistance {
			if r.Origin.DistanceAsCrowFlies(featurePoint) < r.TotalCrowDistance/2 || r.Destination.DistanceAsCrowFlies(featurePoint) < r.TotalCrowDistance/2 {
				possibleNearbyFeature[featureI] = struct{}{}
			}
		}
	}

	// classify cities along route
	log.Println("\nDetermining city names")
	haveCity := make(map[string]struct{})
	stepCityNames := make([]string, len(r.MainRoute.Steps()))
	distanceFromLastCity := make([]float64, len(r.MainRoute.Steps()))
	stepCityNames[0] = r.Origin.NearestCityWithin(500)
	haveCity[stepCityNames[0]] = struct{}{}
	distanceFromLastCity[0] = 0.0
	lastPoint := r.Origin
	p := progressbar.New(len(r.MainRoute.Steps()))
	for stepI, step := range r.MainRoute.Steps() {
		p.Add(1)
		if stepI == 0 {
			continue
		}
		routePoint := Point{
			Longitude: step.Longitude,
			Latitude:  step.Latitude,
		}
		distanceFromLastCity[stepI] = distanceFromLastCity[stepI-1] + lastPoint.DistanceAsCrowFlies(routePoint)
		lastPoint = routePoint
		nearestCityName := routePoint.NearestCityWithin(3)
		if nearestCityName == "" {
			stepCityNames[stepI] = stepCityNames[stepI-1]
		} else {
			if _, ok := haveCity[nearestCityName]; ok {
				stepCityNames[stepI] = stepCityNames[stepI-1]
			} else {
				stepCityNames[stepI] = nearestCityName
				haveCity[nearestCityName] = struct{}{}
				distanceFromLastCity[stepI] = 0
			}
		}
	}

	featureFirstAppears := make(map[int]float64)
	stepOfFeature := make(map[int]int)
	p = progressbar.New(len(possibleNearbyFeature))
	for featureI := range possibleNearbyFeature {
		p.Add(1)
		featurePoint := Point{
			Longitude: allFeatures.Features[featureI].Geometry.Coordinates[0],
			Latitude:  allFeatures.Features[featureI].Geometry.Coordinates[1],
		}
		closestFeature := -1
		closestDistance := 1000000.0
		closestCumulativeDistance := 0.0
		cumulativeDistance := 0.0
		lastPoint := r.Origin
		bestRoutePoint := r.Origin
		bestStepToGetOff := 0
		for stepI, step := range r.MainRoute.Steps() {
			routePoint := Point{
				Longitude: step.Longitude,
				Latitude:  step.Latitude,
			}
			cumulativeDistance += lastPoint.DistanceAsCrowFlies(routePoint)
			lastPoint = routePoint

			distanceEstimate := routePoint.DistanceAsCrowFlies(featurePoint)
			// skip everything over 20 miles off the route
			if distanceEstimate > 20 {
				continue
			}

			if distanceEstimate < closestDistance {
				closestDistance = distanceEstimate
				closestFeature = featureI
				closestCumulativeDistance = cumulativeDistance
				bestRoutePoint = routePoint
				bestStepToGetOff = stepI
			}
		}
		// find which feature is closest
		if closestFeature > -1 {
			// skip things that are MaximumMinutesOffRoute minutes away
			route, _ := bestRoutePoint.DrivingRouteTo(featurePoint)
			if route.Duration() < r.MaximumMinutesOffRoute {
				featureFirstAppears[closestFeature] = closestCumulativeDistance
				stepOfFeature[closestFeature] = bestStepToGetOff
			}
		}
	}

	log.Println("sorting nearby features")
	r.NearbyFeatures = make([]GeoJSONPointsFeature, len(featureFirstAppears))
	haveCity = make(map[string]struct{})
	for i, featureI := range SortMapKeysByValue(featureFirstAppears) {
		r.NearbyFeatures[i] = allFeatures.Features[featureI]
		r.NearbyFeatures[i].Properties.ID = strconv.Itoa(i)
		cityName := stepCityNames[stepOfFeature[featureI]]
		if _, ok := haveCity[cityName]; !ok {
			r.NearbyFeatures[i].Properties.CityThatIsClose = cityName
			haveCity[cityName] = struct{}{}
		}
		r.NearbyFeatures[i].Properties.DistanceFromLastCity = fmt.Sprintf("%2.1f miles", distanceFromLastCity[stepOfFeature[featureI]])
		r.NearbyFeatures[i].Properties.Distance = fmt.Sprintf("%2.1f miles", featureFirstAppears[featureI])
	}
}

func (r RoadTrip) JSONString() (geojson string) {
	allFeatures := make([]string, len(r.NearbyFeatures)+1)
	allFeatures[0] = GeoJSONLineStringFeature(r.MainRoute.Steps())
	for i, feature := range r.NearbyFeatures {
		allFeatures[i+1] = feature.String()
	}
	return fmt.Sprintf(`{"type": "FeatureCollection","features": [%s]}`, strings.Join(allFeatures, ","))
}

func (r RoadTrip) Dump(filename string) error {
	s := r.JSONString()
	return ioutil.WriteFile(filename, []byte(s), 0644)
}

package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

var (
	osrmServer, geoipServer string
)

func main() {
	flag.StringVar(&osrmServer, "osrm", "http://osrm.makemydrive.fun", "address of OSRM server")
	flag.StringVar(&geoipServer, "geoip", "http://geoip.makemydrive.fun", "address of GeoIP server")
	flag.Parse()
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	store := sessions.NewCookieStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./static")
	router.GET("/", handleIndex)
	router.GET("/favicon.ico", func(c *gin.Context) {
		c.Redirect(302, "/static/img/meta/favicon.ico")
	})
	router.GET("/map/:mapid", handleMap)
	router.GET("/directions/:mapid", handleDirections)
	router.POST("/", handlePost)
	fmt.Println("Serving at http://localhost:5003/")
	router.Run(":5003")
}

type MainView struct {
	Error            string
	Previous         map[string]string
	Features         []GeoJSONPointsFeature
	GeoJSON          template.JS
	Latitude         float64
	Longitude        float64
	MapID            string
	Route            RoadTrip
	Acknowledge      string
	RandomAttraction GeoJSONPointsFeature
	CurrentLocation  string
}

func handleIndex(c *gin.Context) {
	serverName := strings.Replace(osrmServer, "http://", "", -1)
	if !strings.Contains(serverName, ":") {
		serverName += ":80"
	}
	_, err := net.DialTimeout("tcp", serverName, 5*time.Second)
	clientIP, errIP := GetClientIPHelper(c.Request)
	var loc string
	if errIP == nil {
		log.Println(clientIP)
		loc, _ = LocationFromIP(clientIP)
	} else {
		log.Println(errIP.Error())
	}
	acknowledge := strings.Title(adjectives[rand.Intn(len(adjectives)-1)])
	if len(acknowledge) < 2 {
		acknowledge = "Go"
	}
	view := MainView{
		Acknowledge:     acknowledge,
		CurrentLocation: loc,
	}
	if err == nil {
		c.HTML(http.StatusOK, "main", view)
	} else {
		log.Println(err)
		c.HTML(http.StatusOK, "down", view)
	}
}

func handleMap(c *gin.Context) {
	mapid := c.Param("mapid")
	r, err := LoadFromName(mapid)
	if err != nil {
		c.HTML(http.StatusOK, "main", MainView{
			Error:       err.Error(),
			Acknowledge: "Go",
		})
	}
	c.HTML(http.StatusOK, "map", MainView{
		Route:     *r,
		Features:  r.NearbyFeatures,
		GeoJSON:   template.JS(r.JSONString()),
		Latitude:  r.Origin.Latitude,
		Longitude: r.Origin.Longitude,
		MapID:     r.Name(),
	})
}

func handleDirections(c *gin.Context) {
	mapid := c.Param("mapid")
	r, err := LoadFromName(mapid)
	if err != nil {
		c.HTML(http.StatusOK, "main", MainView{
			Error:       err.Error(),
			Acknowledge: "Go",
		})
	}
	c.HTML(http.StatusOK, "directions", MainView{
		Route:            *r,
		Features:         r.NearbyFeatures,
		GeoJSON:          template.JS(r.JSONString()),
		Latitude:         r.Origin.Latitude,
		Longitude:        r.Origin.Longitude,
		MapID:            r.Name(),
		RandomAttraction: r.NearbyFeatures[rand.Intn(len(r.NearbyFeatures)-1)],
	})
}

func handlePost(c *gin.Context) {
	type FormInput struct {
		Origin      string `form:"origin" json:"origin" binding:"required"`
		Destination string `form:"destination" json:"destination" binding:"required"`
		Est         string
	}
	var form FormInput
	if err := c.ShouldBind(&form); err == nil {
		origin, err := Geocode(form.Origin)
		if err != nil {
			c.HTML(http.StatusOK, "main", MainView{
				Error:       err.Error(),
				Acknowledge: "Go",
			})
			return
		}
		destination, err := Geocode(form.Destination)
		if err != nil {
			c.HTML(http.StatusOK, "main", MainView{
				Error:       err.Error(),
				Acknowledge: "Go",
			})
			return
		}
		var r *RoadTrip
		r, err = Load(origin, destination)
		if err != nil {
			r, err = New(origin, destination)
			if err != nil {
				c.HTML(http.StatusOK, "main", MainView{
					Error:       err.Error(),
					Acknowledge: "Go",
				})
				return
			}
			r.FindNearbyFeatures()
			err = r.Save()
			if err != nil {
				c.HTML(http.StatusOK, "main", MainView{
					Error:       err.Error(),
					Acknowledge: "Go",
				})
				return
			}
		} else {
			log.Println("loaded")
		}
		c.Redirect(http.StatusMovedPermanently, "/directions/"+r.Name())

	} else {
		c.HTML(http.StatusOK, "main", MainView{
			Error:       err.Error(),
			Acknowledge: "Go",
		})
	}
}

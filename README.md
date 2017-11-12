# Make My Drive Fun!

[![coverage](https://img.shields.io/badge/coverage-53%25-green.svg)](https://gocover.io/github.com/schollz/makemydrivefun)
[![godocs](https://godoc.org/github.com/schollz/makemydrivefun?status.svg)](https://godoc.org/github.com/schollz/makemydrivefun) 


This is the code that powers [makemydrive.fun](https://makemydrive.fun/), a web-app that finds quirky places (places that you wouldn't normally find in a travel book) along the route for a city-to-city road trip. I decided to make this when I was planning on moving and wanted to find some fun places to stop along the way of my planned route. 

How does it work? At the top-level, it will generate a driving route between two cities and then it will attempt to find any of the [17,000 novelty features](https://github.com/schollz/makemydrivefun/blob/master/static/data/features.geojson) that are within 20 minutes driving distance of the route. These are then sorted, collated, and displayed on the web page. There are three "microservices" that are used to accomplish it - the makemydrivefun server, the OSRM routing server, and a GeoIP server. 


# Quick start


First get the Go dependencies:

```
go get -u -v github.com/schollz/makemydrivefun
```

Then `cd` into the directory and build:

```
cd $GOPATH/src/github.com/schollz/makemydrivefun
go build -v
```

Then you should be able to run

```
./makemydrivefun
```

You can also run tests with `go test -cover`.

# Complete install

To install completely, and self-host everything yourself you will also the OSRM server that serves the roadmap, and a GeoIP server.

## Install OSRM server

First download [the North America `.osm.pbf`](http://download.geofabrik.de/). 
Then install the OSRM server [following these instructions](https://github.com/Project-OSRM/osrm-backend#using-docker). _Note:_ that this takes about 60GB of free disk space and you need about 50GB of *free memory* to calculate the database. Also, even with 8 cores it will take about 10 hours to compile the entire North America map.

Once installed, you can run with:

```
docker run -d -t -i -p 5000:5000 -v $(pwd):/data osrm/osrm-backend osrm-routed --algorithm mld /data/north-america-latest.osrm
```

## Install GeoIP server

This is a self-contained docker project as well. Just run:

```
docker run --restart=always -p 5006:8080 -d fiorix/freegeoip
```


## Start makemydrivefun

After everything is setup you can then run with

```
./makemydrivefun -osrm http://localhost:5000 -geoip http://localhost:5006
```

# License

AGPL

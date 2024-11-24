package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	in = []byte(`{
  "origin":{
    "location":{
      "latLng":{
      }
    }
  },
  "destination":{
    "location":{
      "latLng":{
      }
    }
  },
  "mode": "driving",
  "routingPreference": "TRAFFIC_AWARE",
	"extraComputations": ["TRAFFIC_ON_POLYLINE"],
  "computeAlternativeRoutes": true,
  "routeModifiers": {
    "avoidTolls": false,
    "avoidHighways": false,
    "avoidFerries": false
  },
  "languageCode": "en-US",
  "units": "METRIC"
}`)

	apiKeyFlag  = flag.String("api_key", "", "API key")
	origin      = flag.String("source", "", "source, as a lat,long")
	destination = flag.String("destination", "", "destination, as a lat,long")
	travelMode  = flag.String("mode", "TRANSIT", "travel mode, see https://developers.google.com/maps/documentation/routes/vehicles")
)

type Location struct {
	Location struct {
		LatLng struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"latLng"`
	} `json:"location"`
}
type RouteModifiers struct {
	AvoidTolls    bool `json:"avoidTolls"`
	AvoidHighways bool `json:"avoidHighways"`
	AvoidFerries  bool `json:"avoidFerries"`
}
type Request struct {
	Origin      Location `json:"origin"`
	Destination Location `json:"destination"`

	DepartureTime            string         `json:"departureTime"`
	TravelMode               string         `json:"travelMode"`
	RoutingPreference        string         `json:"routingPreference,omitempty"`
	ExtraComputations        []string       `json:"extraComputations"`
	ComputeAlternativeRoutes bool           `json:"computeAlternativeRoutes"`
	RouteModifiers           RouteModifiers `json:"routeModifiers"`
	LanguageCode             string         `json:"languageCode"`
	Units                    string         `json:"units"`
}

const (
	api = "https://routes.googleapis.com/directions/v2:computeRoutes"
)

func parseLatLong(in string) (loc Location, err error) {
	c := strings.Split(in, ",")
	if len(c) != 2 {
		return loc, fmt.Errorf("input %v does not have two comma separated bits, %v", in, c)
	}
	lat, err := strconv.ParseFloat(c[0], 64)
	if err != nil {
		return loc, err

	}
	long, err := strconv.ParseFloat(c[1], 64)
	if err != nil {
		return loc, err
	}
	loc.Location.LatLng.Latitude = lat
	loc.Location.LatLng.Longitude = long
	return loc, err
}
func main() {
	flag.Parse()
	ctx := context.Background()
	client := &http.Client{}
	f := Request{}
	err := json.Unmarshal(in, &f)
	if err != nil {
		log.Fatal(err)
	}
	start := time.Now().Add(time.Minute)
	if err != nil {
		log.Fatal(err)
	}
	t := start
	for ; t.Before(start.Add(24 * time.Hour)); t = t.Add(time.Minute * 10) {
		f.DepartureTime = t.UTC().Format("2006-01-02T15:04:05Z")
		f.TravelMode = *travelMode
		if f.TravelMode != "DRIVE" {
			f.ExtraComputations = []string{}
			f.RoutingPreference = ""
		}
		star, err := parseLatLong(*origin)
		if err != nil {
			log.Fatal(err)
		}
		f.Origin = star
		end, err := parseLatLong(*destination)
		if err != nil {
			log.Fatal(err)
		}
		f.Destination = end

		buf := new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(f); err != nil {
			log.Fatal(err)
		}
		req, err := http.NewRequestWithContext(ctx, "POST", api, buf)
		if err != nil {
			log.Fatal(err)
		}

		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("X-Goog-Api-Key", *apiKeyFlag)
		req.Header.Add("X-Goog-FieldMask", "routes.duration,routes.travelAdvisory,routes.polyline.encodedPolyline")
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		recv := new(Msg)
		if err := json.Unmarshal(body, recv); err != nil {
			log.Fatal(err)
		}

		if len(recv.Routes) == 0 {
			log.Fatal(recv, string(body))
		}
		duration := time.Duration(recv.Routes[0].Duration)
		for _, r := range recv.Routes {
			d := time.Duration(r.Duration)
			if d.Seconds() < duration.Seconds() {
				duration = d
			}

		}
		fmt.Println(t, duration.Seconds()/60.)

	}
}

type Msg struct {
	Routes []Route
}

type Route struct {
	Duration Duration
}

type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		*d = Duration(time.Duration(value))
		return nil
	case string:
		tmp, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = Duration(tmp)
		return nil
	default:
		return errors.New("invalid duration")
	}
}

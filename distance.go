package distance

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Bowbaq/pool"
)

const (
	base_url   = "https://maps.googleapis.com/maps/api/directions/json?"
	workers    = 10
	op_per_sec = 8
)

const (
	Walking   = "walking"
	Bicycling = "bicycling"
	Transit   = "transit"
	Driving   = "driving"
)

type DirectionsAPI struct {
	api_key string
	pool    *pool.Pool
}

func NewDirectionsAPI(api_key string) *DirectionsAPI {
	api := DirectionsAPI{
		api_key: api_key,
	}
	api.pool = pool.NewRateLimitedPool(workers, op_per_sec, op_per_sec, api.get_distance)

	return &api
}

func (api *DirectionsAPI) GetDistance(trip Trip) (uint64, error) {
	job := pool.NewJob(trip)

	api.pool.Submit(job)

	return extract_distance(job.Result())
}

func (api *DirectionsAPI) GetDistances(trips []Trip) map[Trip]uint64 {
	jobs := make(map[Trip]pool.Job)
	for _, trip := range trips {
		job := pool.NewJob(trip)
		jobs[trip] = job

		api.pool.Submit(job)
	}

	result := make(map[Trip]uint64)
	for trip, job := range jobs {
		if distance, err := extract_distance(job.Result()); err == nil {
			result[trip] = distance
		}
	}

	return result
}

func (api *DirectionsAPI) get_distance(id uint, payload interface{}) interface{} {
	req := payload.(Trip)

	q := url.Values{}
	q.Add("key", api.api_key)
	q.Add("origin", req.From.String())
	q.Add("destination", req.To.String())
	q.Add("mode", req.Mode)

	resp, err := http.Get(base_url + q.Encode())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var directions directions_response
	if err := json.NewDecoder(resp.Body).Decode(&directions); err != nil {
		return err
	}

	distance, err := directions.distance()
	if err != nil {
		return err
	}

	return distance
}

func extract_distance(result interface{}) (uint64, error) {
	switch result.(type) {
	case uint64:
		return result.(uint64), nil
	case error:
		return 0, result.(error)
	default:
		return 0, errors.New("unexpected result from pool")
	}
}

type Trip struct {
	From, To Coord
	Mode     string
}

type Coord struct {
	Lat float64
	Lng float64
}

func (coord Coord) String() string {
	return fmt.Sprintf("%.8f,%.8f", coord.Lat, coord.Lng)
}

type directions_response struct {
	Routes []struct {
		Legs []struct {
			Distance struct {
				Value uint64
			}
		}
	}
}

func (directions directions_response) distance() (uint64, error) {
	if len(directions.Routes) > 0 {
		route := directions.Routes[0]
		if len(route.Legs) > 0 {
			return route.Legs[0].Distance.Value, nil
		} else {
			return 0, errors.New("no legs in route")
		}
	} else {
		return 0, errors.New("no route")
	}
}

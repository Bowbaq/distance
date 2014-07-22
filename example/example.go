package main

import (
	"fmt"

	"github.com/Bowbaq/distance"
)

func main() {
	api := distance.NewDirectionsAPI("<you google directions api key>")
	distance, err := api.GetDistance(
		distance.Coord{40.71117416, -74.00016545},
		distance.Coord{40.68382604, -73.97632328},
		distance.Bicycling,
	)

	fmt.Println(distance, err)
}
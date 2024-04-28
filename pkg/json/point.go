package json

type Point struct {
	// field names matched to GPX spec
	Name        string  `json:"name"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Description string  `json:"desc"`
	Symbol      string  `json:"sym"`
}

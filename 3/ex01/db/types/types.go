package types

type Place struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Address  string `json:"address"`
	Phone    string `json:"phone"`
	Location struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"location"`
}

type SearchResult struct {
	Hits struct {
		Hits []struct {
			Source Place `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

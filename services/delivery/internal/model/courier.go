package model

type Courier struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Status    string  `json:"status"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

const (
	CourierAvailable = "AVAILABLE"
	CourierBusy      = "BUSY"
)

package model

type Restaurant struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Address   string `json:"address"`
	IsOpen    bool   `json:"is_open"`
	OpenFrom  string `json:"open_from"`
	OpenTo    string `json:"open_to"`
}

type MenuItem struct {
	ID           string  `json:"id"`
	RestaurantID string  `json:"restaurant_id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Price        float64 `json:"price"`
	Available    bool    `json:"available"`
}

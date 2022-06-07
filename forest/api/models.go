package api

type Country struct {
	Id               string   `json:"id"`
	Name             string   `json:"name"`
	Emoji            string   `json:"emoji"`
	AlternativeNames []string `json:"alternative_names"`
	CurrencyCode     int      `json:"currency_code"`
}

type Location struct {
	Country
	Id               string   `json:"id"`
	Name             string   `json:"name"`
	Latitude         float64  `json:"latitude"`
	Longitude        float64  `json:"longitude"`
	AlternativeNames []string `json:"alternative_names"`
}

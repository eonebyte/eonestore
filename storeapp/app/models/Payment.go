package models

type Payment struct {
	Amount int       `json:"amount"`
	Items  []ItemOld `json:"items"`
}

type ItemOld struct {
	Name  string `json:"name"`
	Price int    `json:"price"`
}

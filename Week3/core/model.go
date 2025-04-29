package core

type Tender struct {
	Id               string  `json:"id"`
	Date             string  `json:"date"`
	Title            string  `json:"title"`
	Description      string  `json:"description"`
	Place            string  `json:"place"`
	Awarded_value    float64 `json:"awarded_value"`
	Awarded_currency string  `json:"awarded_currency"`
	Purchaser_name   string  `json:"purchaser_name"`
	Supplier_name    string  `json:"supplier_name"`
}

type FinalResponse struct {
	CountryName string
	Tenders     []Tender
}

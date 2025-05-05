package tenders

type LicitacionPurchaser struct {
	Name string `json:"name"`
}

type AwardedSupplier struct {
	Suppliers []Suppliers `json:"suppliers"`
}
type Suppliers struct {
	Name string `json:"name"`
}

type ReqLicitation struct {
	Id               string              `json:"id"`
	Title            string              `json:"title"`
	Date             string              `json:"date"`
	Description      string              `json:"description"`
	Place            string              `json:"place"`
	Awarded_value    string              `json:"awarded_value"`
	Awarded_currency string              `json:"awarded_currency"`
	Purcharser       LicitacionPurchaser `json:"purchaser"`
	Awarded          []AwardedSupplier   `json:"awarded"`
}

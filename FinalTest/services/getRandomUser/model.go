package get_random_user

type APILocation struct {
	Country string `json:"country"`
}

type APIRegistered struct {
	Age uint `json:"age"`
}

type APIUser struct {
	Gender   string      `json:"gender"`
	Location APILocation `json:"location"`

	Registered APIRegistered `json:"registered"`
}

type APIResponse struct {
	Results []APIUser `json:"results"`
	// Info    map[string]any `json:"info"`
}

type GenderDistribution struct {
	Male   float64 `json:"male"`
	Female float64 `json:"female"`
}

type TopCountry struct {
	Country string `json:"country"`
	Count   uint   `json:"count"`
}

const BASE_URL = "https://randomuser.me/api/"
const RESULTS_LIMIT = 5000

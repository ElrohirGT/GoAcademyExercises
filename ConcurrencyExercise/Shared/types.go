package shared

type APIUserName struct {
	First string `json:"first"`
	Last  string `json:"last"`
}

type APIStreetLocation struct {
	Number int    `json:"number"`
	Name   string `json:"name"`
}
type APILocation struct {
	Street  APIStreetLocation `json:"street"`
	City    string            `json:"city"`
	State   string            `json:"state"`
	Country string            `json:"country"`
}

type APIUser struct {
	Gender   string      `json:"gender"`
	Name     APIUserName `json:"name"`
	Location APILocation `json:"location"`
	Email    string      `json:"email"`
	Phone    string      `json:"phone"`
}

type APIResponse struct {
	Results []APIUser      `json:"results"`
	Info    map[string]any `json:"info"`
}

type APIError struct {
	Error string `json:"error"`
}

package api

type RegisterRequest struct {
	Ip string `json:"ip"`
}

type RegisterResponse struct {
	Ips map[string]string `json:"ips"`
}

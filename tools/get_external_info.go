package tools

type Data struct {
	Dat struct {
		Host     string `json:"host"`
		Account  string `json:"account"`
		Password string `json:"password"`
	} `json:"dat"`
}

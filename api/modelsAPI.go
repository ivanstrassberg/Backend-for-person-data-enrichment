package api

type PersonReq struct {
	Name       string `json:"name"`
	Surname    string `json:"surname"`
	Patronymic string `json:"patronymic"`
}

type PersonEnriched struct {
	PersonReq
	Age         int
	Gender      string
	Nationality string
}

type AgeResp struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Count int    `json:"count"`
}

type GenderResp struct {
	Name        string  `json:"name"`
	Gender      string  `json:"gender"`
	Probability float64 `json:"probability"`
	Count       int     `json:"count"`
}

type CountryRespMap struct {
	CountryID   string  `json:"country_id"`
	Probability float64 `json:"probability"`
}

type NationalityResp struct {
	Name    string           `json:"name"`
	Country []CountryRespMap `json:"country"`
	Count   int              `json:"count"`
}

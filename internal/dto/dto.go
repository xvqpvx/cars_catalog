package dto

type Filters struct {
	Mark  string
	Model string
	Year  string
	Page  string
	Limit string
}

type AddCarsDto struct {
	Mark   string `json:"mark"`
	Model  string `json:"model"`
	Year   int    `json:"year,omitempty"`
	RegNum string `json:"regNum"`
	Owner  People
}

type UpdateCarDto struct {
	Mark   string `json:"mark,omitempty"`
	Model  string `json:"model,omitempty"`
	Year   string `json:"year,omitempty"`
	RegNum string `json:"regNum,omitempty"`
}

type GetFilteredCarsDto struct {
	CarId int
	Mark  string
	Model string
	Year  string
}

type People struct {
	Name       string `json:"name"`
	Surname    string `json:"surname"`
	Patronymic string `json:"patronymic,omitempty"`
}

type RegNumsRequest struct {
	RegNums []string `json:"regNums"`
}

type Cursors struct {
	Prev string `json:"prev,omitempty"`
	Next string `json:"next,omitempty"`
}

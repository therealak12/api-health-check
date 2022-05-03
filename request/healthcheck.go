package request

type CreateHealthcheck struct {
	IntervalSeconds int               `json:"IntervalSeconds"`
	Url             string            `json:"url"`
	HttpMethod      string            `json:"httpMethod"`
	Headers         map[string]string `json:"headers"`
	Body            string            `json:"body"`
}

type DeleteHealthcheck struct {
	ID int `param:"id" validate:"required,gt=0"`
}

type ToggleHealthcheck struct {
	ID int `param:"id" validate:"required,gt=0"`
}

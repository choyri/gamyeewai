package gamyeewai

type requestData struct {
	Message   string      `json:"message"`
	URL       string      `json:"url"`
	Query     string      `json:"query"`
	Form      string      `json:"form"`
	Body      string      `json:"body"`
	Token     string      `json:"token"`
	ClientIP  string      `json:"client_ip"`
	RawData   string      `json:"raw_data"`
	Timestamp interface{} `json:"timestamp"`
	Trace     string      `json:"trace"`
}

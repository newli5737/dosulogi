package domain

type Cookie struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type CookieExport struct {
	Cookies []Cookie `json:"cookies"`
}

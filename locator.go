package locate

type Locator interface {
	Locate(query string) []Result
}

type Result interface {
	Name() string
}

package catalog

type Lookup interface {
	Get(query string) []interface{}
}

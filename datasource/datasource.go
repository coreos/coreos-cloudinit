package datasource

type Datasource interface {
	ConfigRoot() string
	FetchMetadata() ([]byte, error)
	FetchUserdata() ([]byte, error)
	Type() string
}

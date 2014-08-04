package datasource

const (
	Ec2ApiVersion       = "2009-04-04"
	OpenstackApiVersion = "2012-08-10"
)

type Datasource interface {
	IsAvailable() bool
	AvailabilityChanges() bool
	ConfigRoot() string
	FetchMetadata() ([]byte, error)
	FetchUserdata() ([]byte, error)
	Type() string
}

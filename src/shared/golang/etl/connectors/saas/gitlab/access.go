package gitlab

type AccessLevel int

const (
	NoAccess         AccessLevel = 0
	GuestAccess                  = 10
	ReporterAccess               = 20
	DeveloperAccess              = 30
	MaintainerAccess             = 40
	OwnerAccess                  = 50
)

func (lvl AccessLevel) ToString() string {
	switch lvl {
	case NoAccess:
		return "None"
	case GuestAccess:
		return "Guest"
	case ReporterAccess:
		return "Reporter"
	case DeveloperAccess:
		return "Developer"
	case MaintainerAccess:
		return "Maintainer"
	case OwnerAccess:
		return "Owner"
	}
	return "None"
}

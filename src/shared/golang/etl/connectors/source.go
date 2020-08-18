package connectors

type EtlCommandInfo struct {
	Command    string
	Parameters interface{}
	RawData    string
}

type EtlSourceInfo struct {
	Commands []*EtlCommandInfo
}

func CreateSourceInfo() *EtlSourceInfo {
	return &EtlSourceInfo{
		Commands: []*EtlCommandInfo{},
	}
}

func (s *EtlSourceInfo) AddCommand(c *EtlCommandInfo) {
	s.Commands = append(s.Commands, c)
}

func (s *EtlSourceInfo) MergeWith(other *EtlSourceInfo) {
	s.Commands = append(s.Commands, other.Commands...)
}

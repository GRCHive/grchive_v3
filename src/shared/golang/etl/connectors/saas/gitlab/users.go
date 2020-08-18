package gitlab

type EtlGitlabConnectorUser struct {
	opts *EtlGitlabOptions
}

func createGitlabConnectorUser(opts *EtlGitlabOptions) (*EtlGitlabConnectorUser, error) {
	return &EtlGitlabConnectorUser{
		opts: opts,
	}, nil
}

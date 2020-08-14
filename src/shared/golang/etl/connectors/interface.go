package connectors

import (
	"errors"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
)

var ErrEtlFnNotImpelemented = errors.New("Not implemented.")

type EtlConnectorUserInterface interface {
	GetUserListing() ([]*types.EtlUser, *EtlSourceInfo, error)
}

type EtlConnectorInterface interface {
	GetUserInterface() (EtlConnectorUserInterface, error)
}

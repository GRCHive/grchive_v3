package source_test

import (
	"github.com/onsi/gomega"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"testing"
)

func TestCreateSourceInfo(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	source := connectors.CreateSourceInfo()
	g.Expect(len(source.Commands)).To(gomega.Equal(0))
}

func TestAddCommand(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	cmd := connectors.EtlCommandInfo{}
	cmd2 := connectors.EtlCommandInfo{}

	source := connectors.CreateSourceInfo()

	source.AddCommand(&cmd)
	g.Expect(len(source.Commands)).To(gomega.Equal(1))
	g.Expect(source.Commands[0]).To(gomega.Equal(&cmd))

	source.AddCommand(&cmd2)
	g.Expect(len(source.Commands)).To(gomega.Equal(2))
	g.Expect(source.Commands[1]).To(gomega.Equal(&cmd2))
}

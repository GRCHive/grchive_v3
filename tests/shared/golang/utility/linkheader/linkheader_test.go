package linkheader

import (
	"github.com/onsi/gomega"
	"testing"
)

func TestParseLinkHeaderFromHttpResponse(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	for _, test := range []struct {
		Header  string
		RefLink Links
	}{
		{
			Header: `<https://www.google.com>; rel="next"; title="test title"`,
			RefLink: Links{
				AllLinks: []Link{
					Link{
						Uri: "https://www.google.com",
						Rel: "next",
					},
				},
				RelMap: map[string]*Link{
					"next": &Link{
						Uri: "https://www.google.com",
						Rel: "next",
					},
				},
			},
		},
		{
			Header: `<https://www.google.com>; rel="next",<https://bing.com>; rel="self"`,
			RefLink: Links{
				AllLinks: []Link{
					Link{
						Uri: "https://www.google.com",
						Rel: "next",
					},
					Link{
						Uri: "https://bing.com",
						Rel: "self",
					},
				},
				RelMap: map[string]*Link{
					"next": &Link{
						Uri: "https://www.google.com",
						Rel: "next",
					},
					"self": &Link{
						Uri: "https://bing.com",
						Rel: "self",
					},
				},
			},
		},
	} {
		cmp := ParseLinkHeader(test.Header)
		g.Expect(cmp).To(gomega.Equal(test.RefLink))
	}
}

package linkheader

import (
	"net/http"
	"regexp"
	"strings"
)

var re = regexp.MustCompile(`<(.*)>(; .*=".*")*`)
var paramRe = regexp.MustCompile(`; (.*?)="(.*?)"`)

type Link struct {
	Uri string
	Rel string
}

type Links struct {
	AllLinks []Link
	RelMap   map[string]*Link
}

func (l Links) FindLinkWithRel(rel string) *Link {
	return l.RelMap[rel]
}

func ParseLinkHeader(linkHeader string) Links {
	retLinks := Links{
		RelMap: map[string]*Link{},
	}

	if linkHeader != "" {
		allLinks := []Link{}

		splitLink := strings.Split(linkHeader, ",")
		for _, single := range splitLink {
			match := re.FindAllStringSubmatch(single, -1)

			uri := match[0][1]
			params := match[0][2:]

			newLink := Link{
				Uri: uri,
			}

			for _, p := range params {
				paramMatch := paramRe.FindAllStringSubmatch(p, -1)
				for _, param := range paramMatch {
					if param[1] == "rel" {
						newLink.Rel = param[2]
					}
				}
			}
			allLinks = append(allLinks, newLink)
		}

		retLinks.AllLinks = allLinks
		for idx, link := range allLinks {
			if link.Rel == "" {
				continue
			}
			retLinks.RelMap[link.Rel] = &allLinks[idx]
		}
	}

	return retLinks
}

func ParseLinkHeaderFromHttpResponse(resp *http.Response) Links {
	header := resp.Header
	linkHeader := header.Get("Link")
	return ParseLinkHeader(linkHeader)
}

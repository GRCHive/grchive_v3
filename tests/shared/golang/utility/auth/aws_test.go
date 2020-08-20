package auth_utility

import (
	"github.com/onsi/gomega"
	"gitlab.com/grchive/grchive-v3/shared/test_utility"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestCanonicalAwsHttpMethod(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	for _, test := range []struct {
		Input  string
		Output string
	}{
		{
			"",
			"GET",
		},
		{
			"GET",
			"GET",
		},
		{
			"get",
			"GET",
		},

		{
			"POST",
			"POST",
		},
		{
			"PUT",
			"PUT",
		},
		{
			"DELETE",
			"DELETE",
		},
	} {
		cmp := canonicalAwsHttpMethod(test.Input)
		g.Expect(cmp).To(gomega.Equal(test.Output))
	}
}

func TestCanonicalAwsPath(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	for _, test := range []struct {
		Input  string
		Output string
	}{
		{
			Input:  "/documents and settings/",
			Output: "/documents%2520and%2520settings/",
		},
		{
			Input:  "",
			Output: "/",
		},
	} {
		cmp := canonicalAwsPath(test.Input)
		g.Expect(cmp).To(gomega.Equal(test.Output))
	}
}

func TestCanonicalAwsQueryString(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	for _, test := range []struct {
		Input  url.Values
		Output string
	}{
		{
			Input: url.Values{
				"Action":  []string{"ListUsers"},
				"Version": []string{"2010-05-08"},
			},
			Output: "Action=ListUsers&Version=2010-05-08",
		},
	} {
		cmp := canonicalAwsQueryString(test.Input)
		g.Expect(cmp).To(gomega.Equal(test.Output))
	}
}

func TestCanonicalAwsHeaders(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	for _, test := range []struct {
		Input  http.Header
		Output string
	}{
		{
			Input: http.Header{
				"Content-Type": []string{"application/x-www-form-urlencoded; charset=utf-8"},
				"Host":         []string{"iam.amazonaws.com"},
				"X-Amz-Date":   []string{"20150830T123600Z"},
			},
			Output: `content-type:application/x-www-form-urlencoded; charset=utf-8
host:iam.amazonaws.com
x-amz-date:20150830T123600Z

content-type;host;x-amz-date`,
		},
	} {
		cmp := canonicalAwsHeaders(test.Input)
		g.Expect(cmp).To(gomega.Equal(test.Output))
	}
}

func TestCanonicalAwsSignedHeaders(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	for _, test := range []struct {
		Input  http.Header
		Output string
	}{
		{
			Input: http.Header{
				"Content-Type": []string{"application/x-www-form-urlencoded; charset=utf-8"},
				"Host":         []string{"iam.amazonaws.com"},
				"X-Amz-Date":   []string{"20150830T123600Z"},
			},
			Output: `content-type;host;x-amz-date`,
		},
	} {
		cmp := canonicalAwsSignedHeaders(test.Input)
		g.Expect(cmp).To(gomega.Equal(test.Output))
	}
}

func TestCanonicalAwsPayload(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	for _, test := range []struct {
		Input  string
		Output string
	}{
		{
			Input:  "",
			Output: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			Input:  "asdf",
			Output: "f0e4c2f76c58916ec258f246851bea091d14d4247a2fc3e18694461b1816e13b",
		},
	} {
		cmp, err := canonicalAwsPayload(ioutil.NopCloser(strings.NewReader(test.Input)))
		g.Expect(err).To(gomega.BeNil())
		g.Expect(cmp).To(gomega.Equal(test.Output))
	}
}

func mustCreateUrl(inputUrl string) *url.URL {
	u, _ := url.Parse(inputUrl)
	return u
}

func TestCreateCanonicalRequest(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	for _, test := range []struct {
		Input  http.Request
		Output string
	}{
		{
			Input: http.Request{
				Method: "GET",
				URL:    mustCreateUrl("http://iam.amazonaws.com/?Action=ListUsers&Version=2010-05-08"),
				Header: http.Header{
					"Content-Type": []string{"application/x-www-form-urlencoded; charset=utf-8"},
					"Host":         []string{"iam.amazonaws.com"},
					"X-Amz-Date":   []string{"20150830T123600Z"},
				},
			},
			Output: "f536975d06c0309214f805bb90ccff089219ecd68b2577efef23edd43b7e1a59",
		},
	} {

		tripper := awsRoundTripper{
			clock: test_utility.FixedClock{time.Date(2000, 10, 12, 13, 14, 15, 16, time.UTC)},
		}
		cmp, err := tripper.createCanonicalRequest(&test.Input)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(cmp).To(gomega.Equal(test.Output))
	}
}

func TestCreateStringToSign(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	for _, test := range []struct {
		Input  http.Request
		Time   time.Time
		Output string
	}{
		{
			Input: http.Request{
				Method: "GET",
				URL:    mustCreateUrl("http://iam.amazonaws.com/?Action=ListUsers&Version=2010-05-08"),
				Header: http.Header{
					"Content-Type": []string{"application/x-www-form-urlencoded; charset=utf-8"},
					"Host":         []string{"iam.amazonaws.com"},
					"X-Amz-Date":   []string{"20150830T123600Z"},
				},
			},
			Time: time.Date(2015, 8, 30, 12, 36, 0, 0, time.UTC),
			Output: `AWS4-HMAC-SHA256
20150830T123600Z
20150830/us-east-1/iam/aws4_request
f536975d06c0309214f805bb90ccff089219ecd68b2577efef23edd43b7e1a59`,
		},
	} {

		tripper := awsRoundTripper{
			clock: test_utility.FixedClock{test.Time},
		}
		cmp, err := tripper.createStringToSign(test.Time, &test.Input)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(cmp).To(gomega.Equal(test.Output))
	}
}

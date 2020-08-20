package auth_utility

import (
	"github.com/onsi/gomega"
	"gitlab.com/grchive/grchive-v3/shared/test_utility"
	"gitlab.com/grchive/grchive-v3/shared/utility/time"
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

func TestCanonicalAwsTime(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	for _, test := range []struct {
		Input  time.Time
		Output string
	}{
		{
			Input:  time.Date(2010, 5, 10, 12, 00, 00, 00, time.UTC),
			Output: "20100510T120000Z",
		},
		{
			Input:  time.Date(2020, 12, 1, 12, 35, 30, 00, time.UTC),
			Output: "20201201T123530Z",
		},
	} {
		cmp := canonicalAwsTime(test.Input)
		g.Expect(cmp).To(gomega.Equal(test.Output))
	}
}

func TestCanonicalAwsDate(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	for _, test := range []struct {
		Input  time.Time
		Output string
	}{
		{
			Input:  time.Date(2010, 5, 10, 12, 00, 00, 00, time.UTC),
			Output: "20100510",
		},
		{
			Input:  time.Date(2020, 12, 1, 12, 35, 30, 00, time.UTC),
			Output: "20201201",
		},
	} {
		cmp := canonicalAwsDate(test.Input)
		g.Expect(cmp).To(gomega.Equal(test.Output))
	}
}

func TestCanonicalAwsService(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	for _, test := range []struct {
		Input  string
		Output string
	}{
		{
			Input:  "iam.amazonaws.com",
			Output: "iam",
		},
	} {
		cmp := canonicalAwsService(test.Input)
		g.Expect(cmp).To(gomega.Equal(test.Output))
	}
}

func TestCanonicalAwsScope(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	for _, test := range []struct {
		Time   time.Time
		Host   string
		Output string
	}{
		{
			Time:   time.Date(2010, 5, 10, 12, 00, 00, 00, time.UTC),
			Host:   "iam.amazonaws.com",
			Output: "20100510/us-east-1/iam/aws4_request",
		},
	} {
		cmp := canonicalAwsScope(test.Time, test.Host)
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

func TestCreateAwsSignature(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	for _, test := range []struct {
		Tripper awsRoundTripper
		Request *http.Request
		Output  string
	}{
		{
			Tripper: awsRoundTripper{
				clock:     test_utility.FixedClock{time.Date(2015, 8, 30, 12, 36, 0, 0, time.UTC)},
				keyId:     "ABCDEFGH",
				keySecret: "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY",
			},
			Request: &http.Request{
				Method: "GET",
				URL:    mustCreateUrl("http://iam.amazonaws.com/?Action=ListUsers&Version=2010-05-08"),
				Header: http.Header{
					"Content-Type": []string{"application/x-www-form-urlencoded; charset=utf-8"},
					"Host":         []string{"iam.amazonaws.com"},
					"X-Amz-Date":   []string{"20150830T123600Z"},
				},
			},
			Output: "5d672d79c15b13162d9279b0855cfba6789a8edb4c82c400e06b5924a6f2b5d7",
		},
	} {
		cmp, err := test.Tripper.createAwsSignature(test.Tripper.clock.Now(), test.Request)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(cmp).To(gomega.Equal(test.Output))
	}
}

func TestAddAwsAuthorizationHeaders(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	for _, test := range []struct {
		Tripper awsRoundTripper
		Request *http.Request
		Headers map[string][]string
	}{
		{
			Tripper: awsRoundTripper{
				clock:     test_utility.FixedClock{time.Date(2015, 8, 30, 12, 36, 0, 0, time.UTC)},
				keyId:     "ABCDEFGH",
				keySecret: "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY",
			},
			Request: &http.Request{
				Method: "GET",
				URL:    mustCreateUrl("http://iam.amazonaws.com/?Action=ListUsers&Version=2010-05-08"),
				Header: http.Header{
					"Content-Type": []string{"application/x-www-form-urlencoded; charset=utf-8"},
				},
			},
			Headers: map[string][]string{
				"Content-Type":  []string{"application/x-www-form-urlencoded; charset=utf-8"},
				"Host":          []string{"iam.amazonaws.com"},
				"X-Amz-Date":    []string{"20150830T123600Z"},
				"Authorization": []string{"AWS4-HMAC-SHA256 Credential=ABCDEFGH/20150830/us-east-1/iam/aws4_request, SignedHeaders=content-type;host;x-amz-date, Signature=5d672d79c15b13162d9279b0855cfba6789a8edb4c82c400e06b5924a6f2b5d7"},
			},
		},
	} {
		err := test.Tripper.addAwsAuthorizationHeaders(test.Request)
		g.Expect(err).To(gomega.BeNil())
		for h, v := range test.Headers {
			values, ok := test.Request.Header[h]
			g.Expect(ok).To(gomega.BeTrue())
			g.Expect(values).To(gomega.Equal(v))
		}
	}
}

func TestCreateAWSHttpClient(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	for _, test := range []struct {
		Clock     time_utility.Clock
		KeyId     string
		KeySecret string
	}{
		{
			Clock:     test_utility.FixedClock{time.Date(2015, 8, 30, 12, 36, 0, 0, time.UTC)},
			KeyId:     "12345",
			KeySecret: "ABCDEFGH",
		},
	} {
		client, ok := CreateAWSHttpClient(test.Clock, test.KeyId, test.KeySecret).(*http.Client)
		g.Expect(ok).To(gomega.BeTrue())
		g.Expect(client).NotTo(gomega.BeNil())

		t, ok := client.Transport.(*awsRoundTripper)
		g.Expect(ok).To(gomega.BeTrue())

		g.Expect(t.clock.Now()).To(gomega.BeTemporally("==", test.Clock.Now()))
		g.Expect(t.keyId).To(gomega.Equal(test.KeyId))
		g.Expect(t.keySecret).To(gomega.Equal(test.KeySecret))
	}
}

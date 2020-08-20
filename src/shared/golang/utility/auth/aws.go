package auth_utility

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/utility/crypto"
	"gitlab.com/grchive/grchive-v3/shared/utility/http"
	"gitlab.com/grchive/grchive-v3/shared/utility/time"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const awsSignDateFormat = "20060102"
const awsSignTimeFormat = awsSignDateFormat + "T150405Z"
const awsHashAlg = "AWS4-HMAC-SHA256"

type awsRoundTripper struct {
	clock     time_utility.Clock
	keyId     string
	keySecret string
}

func canonicalAwsHttpMethod(method string) string {
	if method == "" {
		return "GET"
	} else {
		return strings.ToUpper(method)
	}
}

func canonicalAwsPath(path string) string {
	// We're going to assume here that our request path is in the correct (normalized, if necessary) form.
	// So the only thing we need to do URI encode the path (AWS docs says to do it twice).
	if path == "" {
		return "/"
	} else {
		splitPath := strings.Split(path, "/")
		for idx, p := range splitPath {
			splitPath[idx] = url.PathEscape(url.PathEscape(p))
		}
		return strings.Join(splitPath, "/")
	}
}

func canonicalAwsQueryString(query url.Values) string {
	return query.Encode()
}

// This function does the canonical headers as well as the signed headers.
func canonicalAwsHeaders(header http.Header) string {
	// Use the following pseudocode to create the canonical form of the AWS headers.
	// CanonicalHeaders =
	//		CanonicalHeadersEntry0 + CanonicalHeadersEntry1 + ... + CanonicalHeadersEntryN
	//		CanonicalHeadersEntry =
	//		Lowercase(HeaderName) + ':' + Trimall(HeaderValue) + '\n'
	canonicalHeaders := http.Header{}
	sortedHeaderNames := []string{}

	for k, v := range header {
		nm := strings.ToLower(k)

		sortedHeaderNames = append(sortedHeaderNames, nm)
		canonicalHeaders[nm] = make([]string, len(v))
		for i, val := range v {
			canonicalHeaders[nm][i] = strings.TrimSpace(val)
		}
		sort.Strings(canonicalHeaders[nm])
	}

	sort.Strings(sortedHeaderNames)

	ret := strings.Builder{}
	for _, header := range sortedHeaderNames {
		for _, v := range canonicalHeaders[header] {
			entry := fmt.Sprintf("%s:%s\n", header, v)
			ret.WriteString(entry)
		}
	}
	ret.WriteString("\n")
	ret.WriteString(strings.Join(sortedHeaderNames, ";"))
	return ret.String()
}

func canonicalAwsSignedHeaders(header http.Header) string {
	sortedHeaderNames := []string{}

	for k, _ := range header {
		nm := strings.ToLower(k)
		sortedHeaderNames = append(sortedHeaderNames, nm)
	}

	sort.Strings(sortedHeaderNames)
	return strings.Join(sortedHeaderNames, ";")
}

func canonicalAwsPayload(body io.ReadCloser) (string, error) {
	payload, err := ioutil.ReadAll(body)
	if err != nil {
		return "", err
	}

	hashedPayload := sha256.Sum256(payload)
	return hex.EncodeToString(hashedPayload[:]), nil
}

func canonicalAwsTime(tm time.Time) string {
	return tm.UTC().Format(awsSignTimeFormat)
}

func canonicalAwsDate(tm time.Time) string {
	return tm.UTC().Format(awsSignDateFormat)
}

func canonicalAwsService(host string) string {
	// I don't know if this is correct in every situation.
	return strings.Split(host, ".")[0]
}

func canonicalAwsScope(nw time.Time, host string) string {
	return fmt.Sprintf(
		"%s/us-east-1/%s/aws4_request",
		canonicalAwsDate(nw),
		canonicalAwsService(host),
	)
}

func (t *awsRoundTripper) createCanonicalRequest(req *http.Request) (string, error) {
	canonicalRequest := strings.Builder{}

	// Canonical Request Pseudocode:
	// CanonicalRequest =
	// 	HTTPRequestMethod + '\n' +
	// 	CanonicalURI + '\n' +
	// 	CanonicalQueryString + '\n' +
	// 	CanonicalHeaders + '\n' +
	// 	SignedHeaders + '\n' +
	// 	Lowercase(HexEncode(Hash(RequestPayload)))

	// 1. HTTP Request Method
	canonicalRequest.WriteString(canonicalAwsHttpMethod(req.Method))
	canonicalRequest.WriteString("\n")

	// 2. URI.
	canonicalRequest.WriteString(canonicalAwsPath(req.URL.Path))
	canonicalRequest.WriteString("\n")

	// 3. Query String
	canonicalRequest.WriteString(canonicalAwsQueryString(req.URL.Query()))
	canonicalRequest.WriteString("\n")

	// 4/5. Canonical  + Signed Headers
	canonicalRequest.WriteString(canonicalAwsHeaders(req.Header))
	canonicalRequest.WriteString("\n")

	// 6. Payload
	var canonicalPayload string
	var err error
	if req.Body != nil {
		body, err := req.GetBody()
		if err != nil {
			return "", err
		}
		defer body.Close()

		canonicalPayload, err = canonicalAwsPayload(body)
	} else {
		canonicalPayload, err = canonicalAwsPayload(ioutil.NopCloser(strings.NewReader("")))
	}

	if err != nil {
		return "", err
	}
	canonicalRequest.WriteString(canonicalPayload)

	hashedRequest := sha256.Sum256([]byte(canonicalRequest.String()))
	return hex.EncodeToString(hashedRequest[:]), nil
}

func (t *awsRoundTripper) createStringToSign(nw time.Time, req *http.Request) (string, error) {
	// StringToSign =
	// 	Algorithm + \n +
	// 	RequestDateTime + \n +
	// 	CredentialScope + \n +
	// 	HashedCanonicalRequest
	toSign := strings.Builder{}
	// Hardcode SHA256 for now.
	toSign.WriteString(awsHashAlg)
	toSign.WriteString("\n")

	toSign.WriteString(canonicalAwsTime(nw))
	toSign.WriteString("\n")

	toSign.WriteString(canonicalAwsScope(nw, req.URL.Host))
	toSign.WriteString("\n")

	canonicalRequest, err := t.createCanonicalRequest(req)
	if err != nil {
		return "", err
	}
	toSign.WriteString(canonicalRequest)
	return toSign.String(), nil
}

func (t *awsRoundTripper) createAwsSignature(nw time.Time, req *http.Request) (string, error) {
	// To create the signing key:
	// kSecret = your secret access key
	// kDate = HMAC("AWS4" + kSecret, Date)
	// kRegion = HMAC(kDate, Region)
	// kService = HMAC(kRegion, Service)
	// kSigning = HMAC(kService, "aws4_request")
	//
	// NOTE: HMAC(key ,data) in the above pseudocode.
	kSecret := t.keySecret

	kDate := crypto_utility.Sha256HMAC([]byte("AWS4"+kSecret), []byte(canonicalAwsDate(nw)))
	kRegion := crypto_utility.Sha256HMAC(kDate, []byte("us-east-1"))
	kService := crypto_utility.Sha256HMAC(kRegion, []byte(canonicalAwsService(req.URL.Host)))
	kSigning := crypto_utility.Sha256HMAC(kService, []byte("aws4_request"))

	// To compute the signature:
	// signature = HexEncode(HMAC(derived signing key, string to sign))
	signString, err := t.createStringToSign(nw, req)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(crypto_utility.Sha256HMAC(kSigning, []byte(signString))), nil
}

func (t *awsRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	nw := t.clock.Now()
	req.Header.Add("Host", req.URL.Host)
	req.Header.Add("X-Amz-Date", canonicalAwsTime(nw))

	sig, err := t.createAwsSignature(nw, req)
	if err != nil {
		return nil, err
	}

	authHeader := fmt.Sprintf(
		"%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		awsHashAlg,
		t.keyId,
		canonicalAwsScope(nw, req.URL.Host),
		canonicalAwsSignedHeaders(req.Header),
		sig,
	)

	req.Header.Add("Authorization", authHeader)

	return http.DefaultTransport.RoundTrip(req)
}

func CreateAWSHttpClient(clock time_utility.Clock, keyId string, keySecret string) http_utility.HttpClient {
	return &http.Client{
		Transport: &awsRoundTripper{
			clock:     clock,
			keyId:     keyId,
			keySecret: keySecret,
		},
	}
}

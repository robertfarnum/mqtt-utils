package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iot"
)

const (
	port                  = 443
	protocol              = "wss"
	path                  = "/mqtt"
	endpointType          = "iot:Data-ATS"
	subTimeout            = 2000
	pubTimeout            = 2000
	qos1                  = 1
	defaultMQTTBufferSize = 1024
)

func newSession(profile string, region string) (*session.Session, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            aws.Config{Region: aws.String(region)},
		Profile:           profile,
	})
	if err != nil {
		return nil, err
	}

	return sess, nil
}

// GetEndpoint retries the IoT endpoint
func getEndpoint(sess *session.Session) (string, error) {
	svc := iot.New(sess)

	input := iot.DescribeEndpointInput{
		EndpointType: aws.String(endpointType),
	}

	output, err := svc.DescribeEndpoint(&input)
	if err != nil {
		fmt.Printf("Could not get iotdata endpoint: %v", err)
		return "", err
	}

	return *output.EndpointAddress, nil
}

func GetWebsocketUrl(profile string, region string) (*url.URL, error) {
	sess, err := newSession(profile, region)
	if err != nil {
		return nil, err
	}

	endpoint, err := getEndpoint(sess)
	if err != nil {
		return nil, err
	}

	creds, err := sess.Config.Credentials.Get()
	if err != nil {
		return nil, err
	}

	// according to docs, time must be within 5min of actual time (or at least according to AWS servers)
	now := time.Now().UTC()
	dateLong := now.Format("20060102T150405Z")
	dateShort := dateLong[:8]
	serviceName := "iotdevicegateway"
	scope := fmt.Sprintf("%s/%s/%s/aws4_request", dateShort, region, serviceName)
	alg := "AWS4-HMAC-SHA256"
	q := [][2]string{
		{"X-Amz-Algorithm", alg},
		{"X-Amz-Credential", creds.AccessKeyID + "/" + scope},
		{"X-Amz-Date", dateLong},
		{"X-Amz-SignedHeaders", "host"},
	}

	query := awsQueryParams(q)
	signKey := awsSignKey(creds.SecretAccessKey, dateShort, *sess.Config.Region, serviceName)
	stringToSign := awsSignString(creds.AccessKeyID, creds.SecretAccessKey, query, endpoint, dateLong, alg, scope)
	signature := fmt.Sprintf("%x", awsHmac(signKey, []byte(stringToSign)))
	wsURLStr := fmt.Sprintf("%s://%s%s?%s&X-Amz-Signature=%s", protocol, endpoint, path, query, signature)

	if creds.SessionToken != "" {
		wsURLStr = fmt.Sprintf("%s&X-Amz-Security-Token=%s", wsURLStr, url.QueryEscape(creds.SessionToken))
	}

	wsURL, err := url.Parse(wsURLStr)
	if err != nil {
		return nil, err
	}

	return wsURL, nil
}

func awsQueryParams(q [][2]string) string {
	var buff bytes.Buffer
	var i int
	for _, param := range q {
		if i != 0 {
			buff.WriteRune('&')
		}
		i++
		buff.WriteString(param[0])
		buff.WriteRune('=')
		buff.WriteString(url.QueryEscape(param[1]))
	}
	return buff.String()
}

func awsSignString(accessKey string, secretKey string, query string, host string, dateLongStr string, alg string, scopeStr string) string {
	emptyStringHash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	req := strings.Join([]string{
		"GET",
		"/mqtt",
		query,
		"host:" + host,
		"", // separator
		"host",
		emptyStringHash,
	}, "\n")
	return strings.Join([]string{
		alg,
		dateLongStr,
		scopeStr,
		awsSha(req),
	}, "\n")
}

func awsHmac(key []byte, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func awsSignKey(secretKey string, dateShort string, region string, serviceName string) []byte {
	h := awsHmac([]byte("AWS4"+secretKey), []byte(dateShort))
	h = awsHmac(h, []byte(region))
	h = awsHmac(h, []byte(serviceName))
	h = awsHmac(h, []byte("aws4_request"))
	return h
}

func awsSha(in string) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s", in)
	return fmt.Sprintf("%x", h.Sum(nil))
}

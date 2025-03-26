package hz_client

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/ctfloyd/hazelmere-commons/pkg/hz_api"
	jsoniter "github.com/json-iterator/go"
	"io"
	"math"
	"net/http"
	"time"
)

var ErrHttpClient = errors.New("generic http client error")

type HttpClientConfig struct {
	Host           string
	TimeoutMs      int
	Retries        int
	RetryWaitMs    int
	RetryMaxWaitMs int
}
type HttpClient struct {
	config      HttpClientConfig
	client      *http.Client
	errorMap    map[string]error
	errorLogger func(string)
}

func NewHttpClient(config HttpClientConfig, errorLogger func(string)) *HttpClient {
	httpClient := http.Client{
		Timeout: time.Duration(config.TimeoutMs) * time.Millisecond,
	}

	return &HttpClient{
		config:      config,
		client:      &httpClient,
		errorMap:    make(map[string]error),
		errorLogger: errorLogger,
	}
}

func (hc *HttpClient) AddErrorMappings(errors map[string]error) {
	for k, v := range errors {
		hc.errorMap[k] = v
	}
}

func (hc *HttpClient) GetHost() string {
	return hc.config.Host
}

func (hc *HttpClient) GetV1Url() string {
	return fmt.Sprintf("%s/%s", hc.config.Host, "v1")
}

func (hc *HttpClient) Get(url string, response any) error {
	return hc.GetWithHeaders(url, nil, response)
}

func (hc *HttpClient) GetWithHeaders(url string, headers map[string]string, response any) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return errors.Join(err, ErrHttpClient)
	}

	req.Header.Set("Content-Type", "application/json")
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	return hc.doRequest(req, response)
}

func (hc *HttpClient) Post(url string, body any, response any) error {
	return hc.PostWithHeaders(url, nil, body, response)
}

func (hc *HttpClient) PostWithHeaders(url string, headers map[string]string, body any, response any) error {
	bodyBytes, err := jsoniter.Marshal(body)
	if err != nil {
		return errors.Join(ErrHttpClient, err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return errors.Join(err, ErrHttpClient)
	}

	req.Header.Set("Content-Type", "application/json")
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	return hc.doRequest(req, response)
}

func (hc *HttpClient) Patch(url string, body any, response any) error {
	return hc.PatchWithHeaders(url, nil, body, response)
}

func (hc *HttpClient) PatchWithHeaders(url string, headers map[string]string, body any, response any) error {
	bodyBytes, err := jsoniter.Marshal(body)
	if err != nil {
		return errors.Join(ErrHttpClient, err)
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return errors.Join(err, ErrHttpClient)
	}

	req.Header.Set("Content-Type", "application/json")
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	return hc.doRequest(req, response)
}

func (hc *HttpClient) doRequest(request *http.Request, response any) error {
	attempt := 0
	for attempt < hc.config.Retries+1 {
		res, err := hc.client.Do(request)
		if err != nil {
			return errors.Join(err, ErrHttpClient)
		}

		if res.StatusCode >= 200 && res.StatusCode <= 299 {
			return hc.parseSuccessResponse(res, response)
		}

		if res.StatusCode <= 499 {
			return hc.handleNonRetryableErrorResponse(res)
		}

		time.Sleep(time.Duration(hc.computeWaitMs(attempt)) * time.Millisecond)
		attempt += 1
	}
	return errors.New("maximum retry attempts exceeded")
}

func (hc *HttpClient) parseSuccessResponse(res *http.Response, response any) error {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			hc.errorLogger(err.Error() + "\n")
		}
	}(res.Body)

	responseBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return errors.Join(err, ErrHttpClient)
	}

	err = jsoniter.Unmarshal(responseBytes, &response)
	if err != nil {
		return errors.Join(err, ErrHttpClient)
	}

	return nil
}

func (hc *HttpClient) handleNonRetryableErrorResponse(res *http.Response) error {
	errorResponse, err := hc.parseErrorResponse(res)
	if err != nil {
		return errors.Join(ErrHttpClient, err)
	}

	if value, ok := hc.errorMap[errorResponse.Code]; ok {
		return errors.Join(value, errors.New(errorResponse.Message))
	} else {
		return errors.Join(ErrHttpClient, errors.New(fmt.Sprintf("[%s] - %s", errorResponse.Code, errorResponse.Message)))
	}
}

func (hc *HttpClient) parseErrorResponse(res *http.Response) (hz_api.ErrorResponse, error) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			hc.errorLogger(err.Error())
		}
	}(res.Body)

	responseBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return hz_api.ErrorResponse{}, errors.Join(err, ErrHttpClient)
	}

	var errorResponse hz_api.ErrorResponse
	err = jsoniter.Unmarshal(responseBytes, &errorResponse)
	if err != nil {
		return hz_api.ErrorResponse{}, errors.Join(err, ErrHttpClient)
	}

	return errorResponse, nil
}

func (hc *HttpClient) computeWaitMs(attempt int) int {
	return int(math.Min(float64(attempt*hc.config.RetryWaitMs), float64(hc.config.RetryMaxWaitMs)))
}

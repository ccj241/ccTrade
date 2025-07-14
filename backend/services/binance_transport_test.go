package services

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestBinanceTransport_AddApiAgentCode(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		url          string
		contentType  string
		body         string
		apiAgentCode string
		expectInURL  bool
		expectInBody bool
	}{
		{
			name:         "POST with form data should add to body",
			method:       "POST",
			url:          "https://api.binance.com/api/v3/order",
			contentType:  "application/x-www-form-urlencoded",
			body:         "symbol=BTCUSDT&side=BUY",
			apiAgentCode: "JW9QZKMK",
			expectInURL:  false,
			expectInBody: true,
		},
		{
			name:         "POST with JSON should add to URL",
			method:       "POST",
			url:          "https://api.binance.com/api/v3/order",
			contentType:  "application/json",
			body:         `{"symbol":"BTCUSDT"}`,
			apiAgentCode: "JW9QZKMK",
			expectInURL:  true,
			expectInBody: false,
		},
		{
			name:         "POST without content type should add to URL",
			method:       "POST",
			url:          "https://api.binance.com/api/v3/order",
			contentType:  "",
			body:         "",
			apiAgentCode: "JW9QZKMK",
			expectInURL:  true,
			expectInBody: false,
		},
		{
			name:         "GET should not add apiAgentCode by default",
			method:       "GET",
			url:          "https://api.binance.com/api/v3/account",
			contentType:  "",
			body:         "",
			apiAgentCode: "JW9QZKMK",
			expectInURL:  false,
			expectInBody: false,
		},
		{
			name:         "GET to order endpoint should add apiAgentCode",
			method:       "GET",
			url:          "https://api.binance.com/api/v3/order?symbol=BTCUSDT",
			contentType:  "",
			body:         "",
			apiAgentCode: "JW9QZKMK",
			expectInURL:  true,
			expectInBody: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建一个捕获请求的RoundTripper
			var capturedReq *http.Request
			captureTransport := &captureRoundTripper{
				handler: func(req *http.Request) (*http.Response, error) {
					capturedReq = req.Clone(req.Context())
					// 保存body内容
					if req.Body != nil {
						bodyBytes, _ := ioutil.ReadAll(req.Body)
						capturedReq.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
					}
					return &http.Response{
						StatusCode: 200,
						Body:       ioutil.NopCloser(strings.NewReader("{}")),
						Header:     make(http.Header),
					}, nil
				},
			}

			// 创建BinanceTransport
			transport := &BinanceTransport{
				apiAgentCode:  tt.apiAgentCode,
				baseTransport: captureTransport,
				isFutures:     strings.Contains(tt.url, "fapi.binance.com"),
			}

			// 创建请求
			var bodyReader *bytes.Reader
			if tt.body != "" {
				bodyReader = bytes.NewReader([]byte(tt.body))
			} else {
				bodyReader = bytes.NewReader([]byte{})
			}

			req, err := http.NewRequest(tt.method, tt.url, bodyReader)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			// 执行请求
			_, err = transport.RoundTrip(req)
			if err != nil {
				t.Fatalf("RoundTrip failed: %v", err)
			}

			// 验证结果
			if capturedReq == nil {
				t.Fatal("No request was captured")
			}

			// 检查URL中的apiAgentCode
			urlParams := capturedReq.URL.Query()
			hasInURL := urlParams.Get("apiAgentCode") == tt.apiAgentCode

			// 检查Body中的apiAgentCode
			var hasInBody bool
			if capturedReq.Body != nil {
				bodyBytes, _ := ioutil.ReadAll(capturedReq.Body)
				bodyStr := string(bodyBytes)
				if tt.contentType == "application/x-www-form-urlencoded" {
					values, _ := url.ParseQuery(bodyStr)
					hasInBody = values.Get("apiAgentCode") == tt.apiAgentCode
				}
			}

			// 验证期望
			if tt.expectInURL && !hasInURL {
				t.Errorf("Expected apiAgentCode in URL, but not found. URL: %s", capturedReq.URL.String())
			}
			if !tt.expectInURL && hasInURL {
				t.Errorf("Did not expect apiAgentCode in URL, but found. URL: %s", capturedReq.URL.String())
			}
			if tt.expectInBody && !hasInBody {
				bodyBytes, _ := ioutil.ReadAll(capturedReq.Body)
				t.Errorf("Expected apiAgentCode in body, but not found. Body: %s", string(bodyBytes))
			}
			if !tt.expectInBody && hasInBody {
				t.Errorf("Did not expect apiAgentCode in body, but found")
			}
		})
	}
}

// captureRoundTripper 用于捕获HTTP请求
type captureRoundTripper struct {
	handler func(*http.Request) (*http.Response, error)
}

func (c *captureRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return c.handler(req)
}

func TestBinanceTransport_PreserveOriginalRequest(t *testing.T) {
	// 测试确保原始请求不被修改
	transport := NewBinanceTransport("TEST123", false)
	transport.baseTransport = &captureRoundTripper{
		handler: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader("{}")),
				Header:     make(http.Header),
			}, nil
		},
	}

	originalBody := "symbol=BTCUSDT&side=BUY&quantity=1"
	req, _ := http.NewRequest("POST", "https://api.binance.com/api/v3/order", strings.NewReader(originalBody))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 保存原始URL
	originalURL := req.URL.String()

	// 执行请求
	transport.RoundTrip(req)

	// 验证原始请求未被修改
	if req.URL.String() != originalURL {
		t.Errorf("Original request URL was modified. Original: %s, Current: %s", originalURL, req.URL.String())
	}

	// 读取body验证是否可以重新读取
	bodyBytes, _ := ioutil.ReadAll(req.Body)
	if string(bodyBytes) != originalBody {
		t.Errorf("Original request body was not preserved. Expected: %s, Got: %s", originalBody, string(bodyBytes))
	}
}
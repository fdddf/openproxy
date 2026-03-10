package pkg

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
	xproxy "golang.org/x/net/proxy"
)

func CreateHTTPRequest(method, url string, body io.Reader, headers map[string]string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return req, nil
}

func SendHTTPRequest(proxy string, req *http.Request) (*http.Response, error) {
	var transport *http.Transport
	if proxy != "" {
		proxyUrl, err := url.Parse(proxy)
		if err != nil {
			return nil, fmt.Errorf("无效的代理URL: %w", err)
		}

		// Check if it's a SOCKS5 proxy by looking at the scheme
		if proxyUrl.Scheme == "socks5" || proxyUrl.Scheme == "socks5h" {
			// Create a SOCKS5 proxy dialer using the x/net/proxy package
			socksDialer, err := xproxy.SOCKS5("tcp", proxyUrl.Host, nil, xproxy.Direct)
			if err != nil {
				return nil, fmt.Errorf("创建SOCKS5代理失败: %w", err)
			}

			// Create transport with custom dialer for SOCKS5
			transport = &http.Transport{
				Dial: socksDialer.Dial,
			}
		} else {
			// Use the standard HTTP proxy
			transport = &http.Transport{
				Proxy: http.ProxyURL(proxyUrl),
			}
		}
	} else {
		// 使用默认的Transport配置
		transport = http.DefaultTransport.(*http.Transport).Clone()
	}

	client := &http.Client{
		Transport: transport,
	}
	return client.Do(req)
}

func BuildDecompressReader(encoding string, resp *http.Response) (io.Reader, func()) {
	body := resp.Body

	switch encoding {
	case "gzip":
		gr, err := gzip.NewReader(body)
		if err != nil {
			return body, nil
		}
		return gr, func() { gr.Close() }

	case "deflate":
		dr, err := zlib.NewReader(body)
		if err != nil {
			return body, nil
		}
		return dr, func() { dr.Close() }

	case "br":
		br := brotli.NewReader(body)
		return br, nil // brotli reader 无需 close

	case "zstd":
		zr, err := zstd.NewReader(body)
		if err != nil {
			return body, nil
		}
		return zr, func() { zr.Close() }

	case "identity", "":
		// 无压缩
		return body, nil

	default:
		// 未知编码 — 当 identity 处理
		return body, nil
	}
}

// DecompressData decompresses data based on content encoding header
func DecompressData(contentEncoding string, data []byte) ([]byte, error) {
	if contentEncoding == "" || contentEncoding == "identity" {
		return data, nil
	}

	reader := BuildBytesReader(data)
	switch contentEncoding {
	case "gzip":
		gr, err := gzip.NewReader(reader)
		if err != nil {
			return data, nil
		}
		defer gr.Close()
		return io.ReadAll(gr)

	case "deflate":
		dr, err := zlib.NewReader(reader)
		if err != nil {
			return data, nil
		}
		defer dr.Close()
		return io.ReadAll(dr)

	case "br":
		br := brotli.NewReader(reader)
		return io.ReadAll(br)

	case "zstd":
		zr, err := zstd.NewReader(reader)
		if err != nil {
			return data, nil
		}
		defer zr.Close()
		return io.ReadAll(zr)

	default:
		return data, nil
	}
}

// BuildBytesReader creates an io.Reader from byte slice
func BuildBytesReader(data []byte) io.Reader {
	return io.NopCloser(bytes.NewReader(data))
}

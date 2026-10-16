package httpclient

import (
	"compress/gzip"
	"compress/zlib"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/andybalholm/brotli"
	"github.com/imroc/req/v3"
	"github.com/klauspost/compress/zstd"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/pkg/osutil"
	"golang.org/x/net/idna"
	"golang.org/x/sync/singleflight"
)

const (
	maxPublicURLLength = 8192
	maxPublicRedirects = 5
)

var (
	publicDownloadGroup    singleflight.Group
	getPublicClient        = sync.OnceValues(newPublicClient)
	errPublicBodyTooLarge  = errors.New("public origin response exceeds download byte limit")
	publicReservedPrefixes = [...]netip.Prefix{
		netip.MustParsePrefix("0.0.0.0/8"),
		netip.MustParsePrefix("10.0.0.0/8"),
		netip.MustParsePrefix("100.64.0.0/10"),
		netip.MustParsePrefix("127.0.0.0/8"),
		netip.MustParsePrefix("169.254.0.0/16"),
		netip.MustParsePrefix("172.16.0.0/12"),
		netip.MustParsePrefix("192.0.0.0/24"),
		netip.MustParsePrefix("192.0.2.0/24"),
		netip.MustParsePrefix("192.88.99.0/24"),
		netip.MustParsePrefix("192.168.0.0/16"),
		netip.MustParsePrefix("198.18.0.0/15"),
		netip.MustParsePrefix("198.51.100.0/24"),
		netip.MustParsePrefix("203.0.113.0/24"),
		netip.MustParsePrefix("224.0.0.0/4"),
		netip.MustParsePrefix("240.0.0.0/4"),
		netip.MustParsePrefix("2001::/23"),
		netip.MustParsePrefix("2001:db8::/32"),
		netip.MustParsePrefix("2002::/16"),
		netip.MustParsePrefix("3fff::/20"),
	}
	publicIPv6Space = netip.MustParsePrefix("2000::/3")
)

func forbiddenPublicURLCharacters(value string) bool {
	for _, r := range value {
		if r == '\\' || unicode.IsControl(r) {
			return true
		}
	}
	return false
}

func parsePublicURL(rawURL string) (*url.URL, error) {
	if rawURL == "" || len(rawURL) > maxPublicURLLength || !utf8.ValidString(rawURL) || forbiddenPublicURLCharacters(rawURL) {
		return nil, errors.New("invalid public origin URL")
	}
	decoded, err := url.PathUnescape(rawURL)
	if err != nil || forbiddenPublicURLCharacters(decoded) {
		return nil, errors.New("invalid public origin URL encoding")
	}
	u, err := url.Parse(rawURL)
	if err != nil || u.Opaque != "" || u.Host == "" || u.User != nil || strings.Contains(rawURL, "#") {
		return nil, errors.New("public origin must be an absolute URL without credentials or fragment")
	}
	defaultPort := ""
	switch u.Scheme {
	case "http":
		defaultPort = "80"
	case "https":
		defaultPort = "443"
	default:
		return nil, errors.New("public origin requires HTTP or HTTPS")
	}
	port := u.Port()
	if strings.HasSuffix(u.Host, ":") || port != "" && port != defaultPort {
		return nil, errors.New("public origin requires the scheme's default port")
	}
	host := u.Hostname()
	if host == "" || strings.ContainsAny(host, "%:") && !strings.HasPrefix(u.Host, "[") {
		return nil, errors.New("invalid public origin hostname")
	}
	if ip, err := netip.ParseAddr(host); err == nil {
		if !isPublicAddress(ip) {
			return nil, errors.New("public origin address is not publicly routable")
		}
		host = ip.String()
	} else {
		if strings.HasPrefix(u.Host, "[") {
			return nil, errors.New("public origin brackets require an IP literal")
		}
		host, err = idna.Lookup.ToASCII(host)
		if err != nil || !validPublicHostname(host) {
			return nil, errors.New("invalid public origin hostname")
		}
	}
	if port != "" {
		u.Host = net.JoinHostPort(host, port)
	} else if strings.Contains(host, ":") {
		u.Host = "[" + host + "]"
	} else {
		u.Host = host
	}
	return u, nil
}

func validPublicHostname(host string) bool {
	host = strings.TrimSuffix(host, ".")
	if host == "" || len(host) > 253 {
		return false
	}
	for label := range strings.SplitSeq(host, ".") {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, c := range label {
			if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-') {
				return false
			}
		}
	}
	return true
}

func isPublicAddress(ip netip.Addr) bool {
	if !ip.IsValid() || ip.Zone() != "" || ip.Is4In6() || !ip.IsGlobalUnicast() || ip.IsPrivate() {
		return false
	}
	if ip.Is6() && !publicIPv6Space.Contains(ip) {
		return false
	}
	for _, prefix := range publicReservedPrefixes {
		if prefix.Contains(ip) {
			return false
		}
	}
	return true
}

func resolvePublicHost(ctx context.Context, host string) ([]netip.Addr, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if ip, err := netip.ParseAddr(host); err == nil {
		if !isPublicAddress(ip) {
			return nil, errors.New("public origin address is not publicly routable")
		}
		return []netip.Addr{ip}, nil
	}
	addresses, err := net.DefaultResolver.LookupNetIP(ctx, "ip", host)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, errors.New("public origin DNS lookup failed")
	}
	if len(addresses) == 0 {
		return nil, errors.New("public origin DNS lookup returned no addresses")
	}
	for _, ip := range addresses {
		if !isPublicAddress(ip) {
			return nil, errors.New("public origin DNS contains a nonpublic address")
		}
	}
	return addresses, nil
}

func validatePublicURL(ctx context.Context, rawURL string) (*url.URL, error) {
	u, err := parsePublicURL(rawURL)
	if err != nil {
		return nil, err
	}
	if _, err := resolvePublicHost(ctx, u.Hostname()); err != nil {
		return nil, err
	}
	return u, nil
}

// ValidatePublicURL validates both the URL and every address in its current DNS answer.
func ValidatePublicURL(ctx context.Context, rawURL string) error {
	_, err := validatePublicURL(ctx, rawURL)
	return err
}

func newPublicClient() (*req.Client, error) {
	client := req.C().
		ImpersonateChrome().
		SetLogger(nil).
		DisableAutoReadResponse().
		DisableAutoDecode().
		DisableAutoDecompress().
		SetCommonHeader("Accept-Encoding", "identity").
		SetProxy(nil)
	client.SetRedirectPolicy(func(request *http.Request, via []*http.Request) error {
		if len(via) > maxPublicRedirects {
			return errors.New("public origin redirect limit exceeded")
		}
		u, err := validatePublicURL(request.Context(), request.URL.String())
		if err != nil {
			return err
		}
		request.URL = u
		return nil
	})
	dialer := &net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}
	if proxyURL := runtimecfg.Get().HttpClient.Proxy; proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		if err != nil || proxy.Hostname() == "" || proxy.Fragment != "" || proxy.RawQuery != "" {
			return nil, errors.New("invalid configured HTTP client proxy")
		}
		switch proxy.Scheme {
		case "http", "https", "socks5", "socks5h":
		default:
			return nil, errors.New("unsupported configured HTTP client proxy scheme")
		}
		// Only administrator configuration selects a proxy. Its remote DNS and
		// target routing are trusted; the proxy must also enforce public-only egress.
		client.SetProxy(http.ProxyURL(proxy)).SetDial(dialer.DialContext)
		return client, nil
	}
	client.SetDial(func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil || port != "80" && port != "443" {
			return nil, errors.New("invalid public origin dial address")
		}
		addresses, err := resolvePublicHost(ctx, host)
		if err != nil {
			return nil, err
		}
		for _, ip := range addresses {
			// Dial the checked literal, never the hostname, to prevent DNS rebinding.
			conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
			if err == nil {
				return conn, nil
			}
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
		}
		return nil, errors.New("public origin connection failed")
	})
	return client, nil
}

// PublicFetcher downloads only public HTTP origins with a mandatory byte limit.
type PublicFetcher struct {
	MaxBytes int64
}

type publicLimitedWriter struct {
	writer    io.Writer
	remaining int64
}

func (w *publicLimitedWriter) Write(p []byte) (int, error) {
	if int64(len(p)) > w.remaining {
		return 0, errPublicBodyTooLarge
	}
	n, err := w.writer.Write(p)
	w.remaining -= int64(n)
	return n, err
}

func publicResponseBody(response *http.Response) (io.ReadCloser, error) {
	switch strings.ToLower(strings.TrimSpace(strings.Join(response.Header.Values("Content-Encoding"), ","))) {
	case "", "identity":
		return io.NopCloser(response.Body), nil
	case "gzip":
		return gzip.NewReader(response.Body)
	case "deflate":
		return zlib.NewReader(response.Body)
	case "br":
		return io.NopCloser(brotli.NewReader(response.Body)), nil
	case "zstd":
		decoder, err := zstd.NewReader(response.Body,
			zstd.WithDecoderConcurrency(1), zstd.WithDecoderMaxMemory(64<<20))
		if err != nil {
			return nil, err
		}
		return decoder.IOReadCloser(), nil
	default:
		return nil, errors.New("unsupported public origin content encoding")
	}
}

func (f PublicFetcher) FetchToFile(ctx context.Context, rawURL, path string) error {
	if f.MaxBytes <= 0 {
		return errors.New("public origin download byte limit must be positive")
	}
	u, err := validatePublicURL(ctx, rawURL)
	if err != nil {
		return err
	}
	client, err := getPublicClient()
	if err != nil {
		return err
	}
	response, err := client.R().SetContext(ctx).Get(u.String())
	if err != nil {
		if response != nil && response.Response != nil && response.Body != nil {
			response.Body.Close()
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		// HTTP client errors may contain signed URLs or redirect Location values.
		return errors.New("public origin request failed")
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("public origin HTTP status: %d", response.StatusCode)
	}
	if response.ContentLength > f.MaxBytes {
		return errPublicBodyTooLarge
	}
	body, err := publicResponseBody(response.Response)
	if err != nil {
		return errors.New("invalid public origin content encoding")
	}
	defer body.Close()
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(&publicLimitedWriter{writer: file, remaining: f.MaxBytes}, body)
	closeErr := file.Close()
	if copyErr != nil {
		return fmt.Errorf("copy public origin response: %w", copyErr)
	}
	return closeErr
}

func publicCacheExists(path string, maxBytes int64) (bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !info.Mode().IsRegular() {
		return false, errors.New("public origin cache is not a regular file")
	}
	if info.Size() > maxBytes {
		return false, errPublicBodyTooLarge
	}
	return true, nil
}

func openPublicCache(path string, maxBytes int64) (*osutil.File, error) {
	exists, err := publicCacheExists(path, maxBytes)
	if err != nil || !exists {
		return nil, err
	}
	file, err := osutil.OpenCache(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > maxBytes {
		file.Close()
		if err != nil {
			return nil, err
		}
		return nil, errors.New("public origin cache is invalid or exceeds download byte limit")
	}
	return file, nil
}

// DownloadPublicOriginFile validates even cache hits and never shares delegated downloads' flight group.
func DownloadPublicOriginFile(ctx context.Context, rawURL string, maxBytes int64) (*osutil.File, error) {
	if maxBytes <= 0 {
		return nil, errors.New("public origin download byte limit must be positive")
	}
	if err := ValidatePublicURL(ctx, rawURL); err != nil {
		return nil, err
	}
	cachePath := getCachePath(rawURL)
	if file, err := openPublicCache(cachePath, maxBytes); file != nil || err != nil {
		return file, err
	}
	key := cachePath + ":" + strconv.FormatInt(maxBytes, 10)
	result := publicDownloadGroup.DoChan(key, func() (any, error) {
		if exists, err := publicCacheExists(cachePath, maxBytes); exists || err != nil {
			return nil, err
		}
		return nil, downloadToFile(ctx, PublicFetcher{MaxBytes: maxBytes}, rawURL, cachePath)
	})
	select {
	case outcome := <-result:
		if outcome.Err != nil {
			return nil, outcome.Err
		}
		file, err := openPublicCache(cachePath, maxBytes)
		if file == nil && err == nil {
			err = errors.New("public origin download cache disappeared")
		}
		return file, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

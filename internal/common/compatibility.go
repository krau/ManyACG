package common

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/duke-git/lancet/v2/fileutil"
	"github.com/gookit/slog"
	"github.com/gookit/slog/handler"
	"github.com/gookit/slog/rotatefile"
	"github.com/imroc/req/v3"
	"github.com/krau/ManyACG/internal/infra/config"
	"github.com/krau/ManyACG/pkg/osutil"
	"github.com/krau/ManyACG/pkg/strutil"
	"github.com/meilisearch/meilisearch-go"
)

// just for temporary compatibility during refactoring

var Logger *slog.Logger

func init() {
	if Logger != nil {
		return
	}
	slog.DefaultChannelName = "ManyACG"
	Logger = slog.New()
	logLevel := slog.LevelByName("TRACE")
	logFilePath := "./logs/manyacg.log"
	var logBackupNum uint = 7
	var logLevels []slog.Level
	for _, level := range slog.AllLevels {
		if level <= logLevel {
			logLevels = append(logLevels, level)
		}
	}
	consoleH := handler.NewConsoleHandler(logLevels)
	fileH, err := handler.NewTimeRotateFile(
		logFilePath,
		rotatefile.EveryDay,
		handler.WithLogLevels(slog.AllLevels),
		handler.WithBackupNum(logBackupNum),
		handler.WithBuffSize(0),
	)
	if err != nil {
		panic(err)
	}
	Logger.AddHandlers(consoleH, fileH)
}

var (
	defaultClient *req.Client
	once          sync.Once
	cacheLocks    sync.Map
)

func initDefaultClient() {
	c := req.C().ImpersonateChrome().
		SetCommonRetryCount(2).
		SetTLSHandshakeTimeout(time.Second * 10).
		SetTimeout(time.Minute * 2)
	defaultClient = c
	if config.Get().Source.Proxy != "" {
		defaultClient.SetProxyURL(config.Get().Source.Proxy)
	}
}

func getCachePath(url string) string {
	return filepath.Join(config.Get().Storage.CacheDir, "req", strutil.MD5Hash(url))
}

func DownloadWithCache(ctx context.Context, url string, client *req.Client) ([]byte, error) {
	once.Do(initDefaultClient)
	if client == nil {
		client = defaultClient
	}

	cachePath := getCachePath(url)
	data, err := os.ReadFile(cachePath)
	if err == nil {
		return data, nil
	}

	lock, _ := cacheLocks.LoadOrStore(url, &sync.Mutex{})
	lock.(*sync.Mutex).Lock()
	defer func() {
		lock.(*sync.Mutex).Unlock()
		cacheLocks.Delete(url)
	}()

	resp, err := client.R().SetContext(ctx).Get(url)
	if err != nil {
		return nil, err
	}
	if resp.IsErrorState() {
		return nil, fmt.Errorf("http error: %d", resp.GetStatusCode())
	}
	data = resp.Bytes()
	osutil.MkCache(cachePath, data, time.Duration(config.Get().Storage.CacheTTL)*time.Second)
	return data, nil
}

func GetBodyReader(ctx context.Context, url string, client *req.Client) (io.ReadCloser, error) {
	once.Do(initDefaultClient)
	if client == nil {
		client = defaultClient
	}
	cachePath := getCachePath(url)
	if file, err := os.Open(cachePath); err == nil {
		return file, nil
	}

	lock, _ := cacheLocks.LoadOrStore(url, &sync.Mutex{})
	lock.(*sync.Mutex).Lock()
	defer func() {
		lock.(*sync.Mutex).Unlock()
		cacheLocks.Delete(url)
	}()

	if file, err := os.Open(cachePath); err == nil {
		return file, nil
	}
	resp, err := client.R().SetContext(ctx).Get(url)
	if err != nil {
		return nil, err
	}
	if resp.IsErrorState() {
		return nil, fmt.Errorf("http error: %d", resp.GetStatusCode())
	}
	return resp.Body, nil
}

func GetReqCachedFile(url string) ([]byte, error) {
	cachePath := getCachePath(url)
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}
	return data, nil
}

type taggerClient struct {
	Client  *req.Client
	host    string
	token   string
	timeout time.Duration
}

var TaggerClient *taggerClient

func (c *taggerClient) Health() (string, error) {
	var health struct {
		Status string `json:"status"`
	}
	resp, err := c.Client.R().Get("/health")
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(resp.Bytes(), &health); err != nil {
		return "", err
	}
	return health.Status, nil
}

type taggerPredictResponse struct {
	PredictedTags []string           `json:"predicted_tags"`
	Scores        map[string]float64 `json:"scores"`
}

func (c *taggerClient) Predict(ctx context.Context, file []byte) (*taggerPredictResponse, error) {
	resp, err := c.Client.R().SetContext(ctx).SetFileBytes("file", "image", file).Post("/predict")
	if err != nil {
		return nil, err
	}
	if resp.IsErrorState() {
		return nil, fmt.Errorf("tagger predict failed: %s", resp.Status)
	}
	var predict taggerPredictResponse
	if err := json.Unmarshal(resp.Bytes(), &predict); err != nil {
		return nil, err
	}
	return &predict, nil
}
func NewKonatagger(host, token string, timeout time.Duration) (*taggerClient, error) {
	client := req.C().
		SetCommonBearerAuthToken(token).
		SetBaseURL(host).
		SetTimeout(timeout * time.Second).
		SetUserAgent("ManyACG")
	tagerC := &taggerClient{
		Client:  client,
		host:    host,
		token:   token,
		timeout: timeout * time.Second,
	}
	if _, err := tagerC.Health(); err != nil {
		return nil, fmt.Errorf("tagger health check failed: %w", err)
	}
	return tagerC, nil
}

var fileLocks sync.Map

// 创建文件, 自动创建目录
func MkFile(path string, data []byte) error {
	lock, _ := fileLocks.LoadOrStore(path, &sync.Mutex{})
	lock.(*sync.Mutex).Lock()
	defer func() {
		lock.(*sync.Mutex).Unlock()
		fileLocks.Delete(path)
	}()

	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, os.ModePerm)
}

func FileExists(path string) bool {
	return fileutil.IsExist(path)
}

// 删除文件, 并清理空目录. 如果文件不存在则返回 nil
func PurgeFile(path string) error {
	if err := os.Remove(path); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return RemoveEmptyDirectories(filepath.Dir(path))
}

// 递归删除空目录
func RemoveEmptyDirectories(dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		err := os.Remove(dirPath)
		if err != nil {
			return err
		}
		return RemoveEmptyDirectories(filepath.Dir(dirPath))
	}
	return nil
}

// 在指定时间后删除和清理文件 (定时器)
func PurgeFileAfter(path string, td time.Duration) {
	_, err := os.Stat(path)
	if err != nil {
		return
	}
	time.AfterFunc(td, func() {
		PurgeFile(path)
	})
}

var timerMap sync.Map

func RmFileAfter(path string, td time.Duration) {
	if _, ok := timerMap.Load(path); ok {
		return
	}
	timerMap.Store(path, struct{}{})
	defer timerMap.Delete(path)

	_, err := os.Stat(path)
	if err != nil {
		return
	}
	time.AfterFunc(td, func() {
		os.Remove(path)
	})
}

func MkCache(path string, data []byte, td time.Duration) {
	if err := MkFile(path, data); err != nil {
		return
	}
	go RmFileAfter(path, td)
}

const defaultCharset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func GenerateRandomString(length int, charset ...string) string {
	var letters string
	if len(charset) > 0 {
		letters = strings.Join(charset, "")
	} else {
		letters = defaultCharset
	}
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func MD5Hash(data string) string {
	sum := md5.Sum([]byte(data))
	return hex.EncodeToString(sum[:])
}

var tagRe = regexp.MustCompile(`(?:^|[\p{Zs}\s.,!?(){}[\]<>\"\'，。！？（）：；、])#([\p{L}\d_]+)`)

func ExtractTagsFromText(text string) []string {
	matches := tagRe.FindAllStringSubmatch(text, -1)
	tags := make([]string, 0)
	for _, match := range matches {
		if len(match) > 1 {
			tags = append(tags, match[1])
		}
	}
	return tags
}

func TagRegex() *regexp.Regexp {
	return tagRe
}

// ParseTo2DArray parses a string into a 2D array using two separators.
//
// ParseTo2DArray("1,2,3;4,5,6", ",", ";") => [][]string{{"1", "2", "3"}, {"4", "5", "6"}}
//
// ParseTo2DArray("1,2,3;\"4,5,6\"", ",", ";") => [][]string{{"1", "2", "3"}, {"4,5,6"}}
func ParseStringTo2DArray(str, sep, sep2 string) [][]string {
	var result [][]string
	if str == "" {
		return result
	}

	var row []string
	var inQuote bool
	var builder strings.Builder

	for _, c := range str {
		if inQuote {
			if c == '"' || c == '\'' {
				inQuote = false
			} else {
				builder.WriteRune(c)
			}
		} else {
			if c == '"' || c == '\'' {
				inQuote = true
			} else if string(c) == sep {
				row = append(row, builder.String())
				builder.Reset()
			} else if string(c) == sep2 {
				row = append(row, builder.String())
				result = append(result, row)
				row = nil
				builder.Reset()
			} else {
				builder.WriteRune(c)
			}
		}
	}

	if builder.Len() > 0 {
		row = append(row, builder.String())
	}
	if len(row) > 0 {
		result = append(result, row)
	}

	return result
}

func EscapeMarkdown(text string) string {
	panic("deprecated")
}

var MeilisearchClient *meilisearch.ServiceManager

var Client *req.Client

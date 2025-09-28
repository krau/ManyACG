package log_test

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/krau/ManyACG/pkg/log"
)

func readLastLine(t *testing.T, filePath string) string {
	t.Helper()
	f, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("无法打开日志文件: %v", err)
	}
	defer f.Close()

	var lastLine string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lastLine = scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}
	return lastLine
}

func TestZapLogger(t *testing.T) {
	// 创建临时目录和日志文件
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	log := log.ZapLog(log.Config{
		LogFile: logFile,
	})

	// 写日志
	log.Info("服务启动", "port", 8080)
	log.Debug("调试信息", "user", "alice")
	log.Error("出错了", "err", "连接超时")

	// 确认文件存在
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Fatalf("日志文件未生成: %s", logFile)
	}

	// 读取最后一行（应为 Error）
	lastLine := readLastLine(t, logFile)

	// 检查关键字段是否存在
	if !strings.Contains(lastLine, `"msg":"出错了"`) {
		t.Errorf("日志内容缺失 msg，实际: %s", lastLine)
	}
	if !strings.Contains(lastLine, `"err":"连接超时"`) {
		t.Errorf("日志内容缺失字段 err，实际: %s", lastLine)
	}
	if !strings.Contains(lastLine, `"caller"`) {
		t.Errorf("日志未包含 caller 信息，实际: %s", lastLine)
	}
}

func TestCharmLogger(t *testing.T) {
	// 创建临时目录和日志文件
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	log := log.CharmLog(log.Config{
		LogFile: logFile,
	})

	// 写日志
	log.Info("服务启动", "port", 8080)
	log.Debug("调试信息", "user", "alice")
	log.Error("出错了", "err", "连接超时")

	// 确认文件存在
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Fatalf("日志文件未生成: %s", logFile)
	}

	// 读取最后一行（应为 Error）
	lastLine := readLastLine(t, logFile)

	// 检查关键字段是否存在
	if !strings.Contains(lastLine, `"msg":"出错了"`) {
		t.Errorf("日志内容缺失 msg，实际: %s", lastLine)
	}
	if !strings.Contains(lastLine, `"err":"连接超时"`) {
		t.Errorf("日志内容缺失字段 err，实际: %s", lastLine)
	}
	if !strings.Contains(lastLine, `"caller"`) {
		t.Errorf("日志未包含 caller 信息，实际: %s", lastLine)
	}
}

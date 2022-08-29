package my_log

import (
	"errors"
	"my_log/logger"
	"testing"

	"go.uber.org/zap"
)

func Test_Info(t *testing.T) {

	conf := make(map[string]interface{})
	conf["logName"] = "my_log"
	conf["logPath"] = "/Users/akuan/vs_workspace/ccblog2/log"

	logger.Init2(conf)
	defer logger.Sync()

	logCopy := logger.LogAppend(zap.String("requestId", "ae6371606baa38b04c0a89fa5c23a14e"))
	logCopy.Info("this is a test!", zap.Bool("yes", true))
	logger.Info("this is a test!", zap.Bool("yes", true))
	logger.Error("test error", zap.Error(errors.New("this is a test error log!")))
}

package logger

import "github.com/sirupsen/logrus"

// New создаёт настроенный logrus-логгер.
// Уровень логирования задаётся через переменную окружения LOG_LEVEL (debug, info, warn, error).
func New(level string) *logrus.Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05Z07:00",
	})

	parsed, err := logrus.ParseLevel(level)
	if err != nil {
		parsed = logrus.InfoLevel
	}
	log.SetLevel(parsed)
	return log
}

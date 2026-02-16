package pkg

import "time"

type TelegramMeta struct {
	UserID        int64
	ChatID        int64
	VideoDuration int
	MessageID     int
	Timestamp     time.Time
}

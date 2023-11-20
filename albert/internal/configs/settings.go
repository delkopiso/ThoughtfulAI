package configs

import (
	"time"
)

const AppName = "albert"
const CsrfFormFieldName = "csrf-token"
const SessionKeyUserId = "user-id"

type Settings struct {
	ApiKey                 string        `required:"true" split_words:"true"`
	CsrfKey                []byte        `required:"true" split_words:"true"`
	DatabaseUrl            string        `required:"true" split_words:"true"`
	WebAddress             string        `default:"0.0.0.0:8080" split_words:"true"`
	LoadStockFrequency     time.Duration `default:"24h" split_words:"true"`
	LoadPricesFrequency    time.Duration `default:"1m" split_words:"true"`
	PriceMaxAge            time.Duration `default:"5m" split_words:"true"`
	ShutdownWindow         time.Duration `default:"5s" split_words:"true"`
	SessionMaxLifetime     time.Duration `default:"6h" split_words:"true"`
	SessionCleanupInterval time.Duration `default:"24h" split_words:"true"`
}

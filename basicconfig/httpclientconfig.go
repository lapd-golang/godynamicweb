package basicconfig

import (
	"net/http"
	"time"
)

func ConfigHttpClient() {
	http.DefaultClient.Timeout = time.Second * 10
}

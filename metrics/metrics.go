package metrics

import (
	"context"
	"expvar"
	"github.com/paulbellamy/ratecounter"
	"log"
	"net/http"
	"time"
)

var (
	start            = time.Now()
	srv              *http.Server
	RequestCounter   = ratecounter.NewRateCounter(1 * time.Minute)
	ErrorEncountered *expvar.Int
	JoinedGuilds     *expvar.Int
)

func calculateUptime() interface{} {
	return int64(time.Since(start).Seconds())
}

func init() {
	expvar.Publish("uptime", expvar.Func(calculateUptime))
	ErrorEncountered = expvar.NewInt("errorEncountered")
	JoinedGuilds = expvar.NewInt("joinedGuilds")
	expvar.Publish("requestsPerMinute", RequestCounter)
}

func Close() {
	if srv != nil {
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		srv.Shutdown(ctx)
	}
}

func SetServer() {
	srv = &http.Server{Addr: ":8080", Handler: expvar.Handler()}

	go func() {
		if err := http.ListenAndServe(":8080", http.DefaultServeMux); err != nil {
			ErrorEncountered.Add(1)
			log.Fatal(err)
		}
	}()
}

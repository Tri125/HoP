package metrics

import (
	"context"
	"expvar"
	"log"
	"net/http"
	"time"

	"github.com/paulbellamy/ratecounter"
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
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}
}

func SetServer() {
	http.Handle("/metrics", expvar.Handler())
	srv = &http.Server{Addr: ":8080", Handler: expvar.Handler()}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			ErrorEncountered.Add(1)
			log.Fatal(err)
		}
	}()
}

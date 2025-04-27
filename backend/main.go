package main

import (
	"encoding/json"
	"flag"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func main() {
	var socketPath string
	flag.StringVar(&socketPath, "socket", "/run/guest-services/backend.sock", "Unix domain socket to listen on")
	flag.Parse()

	_ = os.RemoveAll(socketPath)

	logger.SetOutput(os.Stdout)

	logMiddleware := middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper: middleware.DefaultSkipper,
		Format: `{"time":"${time_rfc3339_nano}","id":"${id}",` +
			`"method":"${method}","uri":"${uri}",` +
			`"status":${status},"error":"${error}"` +
			`}` + "\n",
		CustomTimeFormat: "2006-01-02 15:04:05.00000",
		Output:           logger.Writer(),
	})

	logger.Infof("Starting listening on %s\n", socketPath)
	router := echo.New()
	router.HideBanner = true
	router.Use(logMiddleware)
	startURL := ""

	ln, err := listen(socketPath)
	if err != nil {
		logger.Fatal(err)
	}
	router.Listener = ln

	router.GET("/hello", hello)
	router.GET("/apps", getApps)

	logger.Fatal(router.Start(startURL))
}

func listen(path string) (net.Listener, error) {
	return net.Listen("unix", path)
}

func hello(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, HTTPMessageBody{Message: "hello"})
}

type HTTPMessageBody struct {
	Message string
}

type AppSummary struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Repository  string `json:"repository"`
	Icon        string `json:"icon,omitempty"`
}

type appFeed struct {
	Applist []struct {
		Name       string `json:"Name"`
		Overview   string `json:"Overview"`
		Repository string `json:"Repository"`
		Icon       string `json:"Icon"`
	} `json:"applist"`
}

var (
	appsCache      []AppSummary
	appsCacheTime  time.Time
	appsCacheTTL   = 10 * time.Minute
)

func getApps(c echo.Context) error {
	if time.Since(appsCacheTime) > appsCacheTTL {
		feed, err := fetchAppFeed()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		var summaries []AppSummary
		for _, app := range feed.Applist {
			summaries = append(summaries, AppSummary{
				Name:        app.Name,
				Description: app.Overview,
				Repository:  app.Repository,
				Icon:        app.Icon,
			})
		}
		appsCache = summaries
		appsCacheTime = time.Now()
	}
	return c.JSON(http.StatusOK, appsCache)
}

func fetchAppFeed() (*appFeed, error) {
	resp, err := http.Get("https://assets.ca.unraid.net/feed/applicationFeed.json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var feed appFeed
	if err := json.Unmarshal(body, &feed); err != nil {
		return nil, err
	}
	return &feed, nil
}

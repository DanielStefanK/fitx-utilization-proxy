package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/DanielStefanK/fitx-utilization-proxy/responses"
	"github.com/DanielStefanK/fitx-utilization-proxy/store"

	ginzap "github.com/gin-contrib/zap"
	"go.uber.org/zap"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.New()
	store := store.NewStore()
	logger, _ := zap.NewProduction()
	router.Use(ginzap.Ginzap(logger, time.RFC3339, true))

	logger.Info("creating new interval to fetch store info")
	ticker := time.NewTicker(6 * time.Hour)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				logger.Info("updating store info")
				store.UpdateStudios()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	router.GET("/api/utilization/:studioId", func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("studioId"), 10, 64)

		if err != nil {
			c.PureJSON(404, &responses.ErrorResponse{Message: "could not parse studio id"})
			return
		}

		if !store.StudioExists(id) {
			c.PureJSON(404, responses.ErrorResponse{
				Message: "studio with provided id could not be found",
			})
			return
		}

		logger.Info("getting utilization for studio", zap.Uint64("studioId", id), zap.String("operation", "getUtilization"))

		resp := store.Get(id)

		if resp == nil {
			c.PureJSON(404, responses.ErrorResponse{
				Message: "Could not get utilization for provided studio id",
			})
			return
		}

		c.PureJSON(200, resp)
	})

	router.GET("/api/studios", func(c *gin.Context) {

		resp := store.GetStudios()

		if resp == nil {
			c.PureJSON(404, responses.ErrorResponse{
				Message: "Could not get studio infos",
			})
			return
		}

		c.PureJSON(200, resp)
	})

	router.Use(middleware("/", "./static"))

	router.Run(":8080")
}

func middleware(urlPrefix, spaDirectory string) gin.HandlerFunc {
	directory := static.LocalFile(spaDirectory, true)
	fileserver := http.FileServer(directory)
	if urlPrefix != "" {
		fileserver = http.StripPrefix(urlPrefix, fileserver)
	}
	return func(c *gin.Context) {
		if directory.Exists(urlPrefix, c.Request.URL.Path) {
			fileserver.ServeHTTP(c.Writer, c.Request)
			c.Abort()
		} else {
			c.Request.URL.Path = "/"
			fileserver.ServeHTTP(c.Writer, c.Request)
			c.Abort()
		}
	}
}

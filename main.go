package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/DanielStefanK/fitx-utilization-proxy/responses"
	"github.com/DanielStefanK/fitx-utilization-proxy/store"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	store := store.NewStore()

	ticker := time.NewTicker(6 * time.Hour)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
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

		c.PureJSON(200, store.Get(id))
	})

	router.GET("/api/studios", func(c *gin.Context) {
		c.PureJSON(200, store.GetStudios())
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

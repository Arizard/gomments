package main

import (
	"fmt"
	"log"
	"os"

	"github.com/arizard/gomments"
	"github.com/arizard/gomments/internal"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/contrib/secure"
	"github.com/gin-gonic/gin"
)

type gommentsEnv struct {
	required map[string]string
	optional map[string]string
}

func main() {
	env := gommentsEnv{
		required: map[string]string{
			"PORT": "",
		},
		optional: map[string]string{
			"BASE_URL": "",
		},
	}

	for k := range env.required {
		v := os.Getenv(k)
		if v == "" {
			log.Fatalf("missing env variable: %s", k)
		}

		env.required[k] = v
	}

	for k := range env.optional {
		env.optional[k] = os.Getenv(k)
	}

	log.Printf("base url is %q", env.optional["BASE_URL"])

	dbx, err := internal.InitSQLiteDatabase("/home/appuser/data/gomments.db")
	if err != nil {
		log.Fatalf("getting migrated dbx: %s", err)
		return
	}

	svc := gomments.New(dbx)

	router := gin.Default()

	// Set request size limits to prevent memory exhaustion attacks
	router.MaxMultipartMemory = 1 << 20 // 1 MB

	router.SetTrustedProxies(nil)

	// Security headers middleware
	router.Use(secure.Secure(secure.Options{
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		ContentSecurityPolicy: "default-src 'self'",
	}))

	corsCfg := cors.DefaultConfig()
	corsCfg.AllowOrigins = []string{"http://localhost:1313", "https://less.coffee"}
	corsCfg.AllowMethods = []string{"GET", "POST", "OPTIONS"}
	router.Use(cors.New(corsCfg))
	router.Use(internal.NewClientIPRateLimiterMiddleware(10))

	internal.InitRoutes(router, svc, internal.InitRoutesOptions{BaseURL: env.optional["BASE_URL"]})

	if err := router.Run(fmt.Sprintf(":%s", env.required["PORT"])); err != nil {
		log.Fatalln(err.Error())
		return
	}
}

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/arizard/gomments"
	"github.com/arizard/gomments/internal"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type GommentsEnv struct {
	Required map[string]string
	Optional map[string]string
}

func main() {
	env := GommentsEnv{
		Required: map[string]string{
			"PORT": "",
		},
		Optional: map[string]string{
			"BASE_URL": "",
		},
	}

	for k := range env.Required {
		v := os.Getenv(k)
		if v == "" {
			log.Fatalf("missing env variable: %s", k)
		}

		env.Required[k] = v
	}

	for k := range env.Optional {
		env.Optional[k] = os.Getenv(k)
	}

	log.Printf("base url is %q", env.Optional["BASE_URL"])

	dbx, err := internal.InitSQLiteDatabase("/root/data/gomments.db")
	if err != nil {
		log.Fatalf("getting migrated dbx: %s", err)
		return
	}

	svc := gomments.New(dbx)

	router := gin.Default()

	router.SetTrustedProxies(nil)

	corsCfg := cors.DefaultConfig()
	corsCfg.AllowOrigins = []string{"http://localhost:1313", "https://less.coffee"}
	corsCfg.AllowMethods = []string{"GET", "POST", "OPTIONS"}
	router.Use(cors.New(corsCfg))

	internal.InitRoutes(router, svc, internal.InitRoutesOptions{BaseURL: env.Optional["BASE_URL"]})

	if err := router.Run(fmt.Sprintf(":%s", env.Required["PORT"])); err != nil {
		log.Fatalln(err.Error())
		return
	}
}

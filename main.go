package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/barsuk/sxgeo"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var gracefulStop = make(chan os.Signal)

	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	go func() {
		sig := <-gracefulStop
		fmt.Printf("Caught sig: %+v\n", sig)
		fmt.Println("I am that force which ever brings you good though wants be evil")
		os.Exit(0)
	}()

	var ip string
	var endian bool
	var setEndian int
	var dbPath string
	flag.StringVar(&ip, "ip", "", "ip address to convert")
	flag.IntVar(&setEndian, "se", 0, "set endianness")
	flag.BoolVar(&endian, "e", false, "check endianness of your system")
	flag.StringVar(&dbPath, "d", "./SxGeoCity.dat", "path to SxGeoCity.dat file")
	flag.Parse()

	if endian {
		sxgeo.DetectEndian()
		os.Exit(0)
	}

	if setEndian > 0 {
		sxgeo.SetEndian(sxgeo.BIG)
		fmt.Printf("host binary endian set to %s\n", sxgeo.Endian())
	}

	if _, err := sxgeo.ReadDBToMemory(dbPath); err != nil {
		log.Fatalf("error: cannot read database file: %v", err)
	}

	if len(ip) > 0 {
		city, err := sxgeo.GetCityFull(ip)
		if err != nil {
			fmt.Printf("error: %v", err)
			os.Exit(1)
		}

		enc, err := json.Marshal(city)
		if err != nil {
			fmt.Printf("error: %v", err)
			os.Exit(1)
		}

		fmt.Printf("%s\n", enc)
		os.Exit(0)
	}
	r := gin.New()
	r.Use(setHeaders)
	r.OPTIONS("/", optionsHandler)
	r.GET("/", sxgeoHandler)
	erro := r.Run(fmt.Sprintf(":%d", 8080))
	if erro != nil {
		log.Fatal("is there any angel? Gin don't like them...")
	}
}

func sxgeoHandler(c *gin.Context) {
	ip := c.Query("ip")
	if len(ip) < 4 {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "give me an IP, please"})
		return
	}
	fmt.Printf("IP: %s\n", ip)
	city, err := sxgeo.GetCityFull(ip)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusAccepted, city)
}

func optionsHandler(c *gin.Context) {
	c.JSON(http.StatusAccepted, "")
}

func setHeaders(c *gin.Context) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Header("Access-Control-Allow-Origin", setAccessHost(c.Request.Header.Get("Origin")))
	c.Header("Access-Control-Allow-Credentials", "true")
	c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	c.Header("Access-Control-Max-Age", "600")
	c.Header("Access-Control-Allow-Headers", "origin, content-type")
	c.Header("Connection", "keep-alive")
}

// CORS hosts
func setAccessHost(origin string) string {
	for _, v := range accessHosts {
		if v == "origin" {
			return origin
		}
		if v == origin {
			return v
		}
	}
	return ""
}

var accessHosts = []string{"origin",}

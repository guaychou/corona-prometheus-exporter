package main

import (
	"errors"
	"flag"
	"net/http"
	"os"
	"time"

	cga "github.com/guaychou/corona-api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

func main() {
	countryPtr := flag.String("country", "", "Country name you want to get COVID19 status")
	portPtr := flag.String("listen.port",":10198", "Port listen address")
	flag.Parse()
	if *countryPtr=="" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	err:=checkCountry(countryPtr)
	if err!=nil{
		log.Fatal(err)
	}
	var (
		confirmed = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "confirmed_corona_"+*countryPtr,
			Help: "Current total confirmed corona in "+*countryPtr,
		})
		recovered = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "recovered_corona_"+*countryPtr,
			Help: "Current total recovered corona in "+*countryPtr,
		})
		recoveryRate =prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "recovery_rate_corona_"+*countryPtr,
			Help: "Current recovery rate in "+*countryPtr,
		})
		death = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "death_corona_"+*countryPtr,
			Help: "Current total death people in "+*countryPtr,
		})
		deathRate =prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "death_rate_corona_"+*countryPtr,
			Help: "Current fatality rate in "+*countryPtr,
		})
	)
	prometheus.MustRegister(confirmed)
	prometheus.MustRegister(recovered)
	prometheus.MustRegister(death)
	prometheus.MustRegister(deathRate)
	prometheus.MustRegister(recoveryRate)
	go func() {
		log.Info("Scrapping corona status in "+*countryPtr)
		for {
			result:=get(countryPtr)
			log.Info("Scrapped . . . ")
			setPrometheusValue(confirmed, recovered, death,deathRate,recoveryRate, result)
			time.Sleep(5 * time.Second)
		}
	}()
	http.Handle("/metrics", promhttp.Handler())
	log.Println("Web Server started. Listening on port "+*portPtr)
	log.Fatal(http.ListenAndServe(*portPtr, nil))
}

func checkCountry(countryPtr *string) error {
	result:=get(countryPtr)
	if result.Recovered.Value==-1 || result.Deaths.Value==-1 || result.Confirmed.Value==-1{
		return errors.New("Country not found")
	}
	return nil
}

func get(country *string)cga.CurrentCoronaStatus{
		result:=cga.GetCorona(*country)
		return result
}

func setPrometheusValue(confirmed prometheus.Gauge, recovered prometheus.Gauge, death prometheus.Gauge,deathRate prometheus.Gauge,recoveryRate prometheus.Gauge, result cga.CurrentCoronaStatus){
	confirmed.Set(float64(result.Confirmed.Value))
	recovered.Set(float64(result.Recovered.Value))
	death.Set(float64(result.Deaths.Value))
	deathRate.Set(result.CaseFatalityRate)
	recoveryRate.Set(result.CaseRecoveryRate)
}
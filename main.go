package main

import (
	"errors"
	"flag"
	"net/http"
	"os"
	"strings"
	"time"

	//"os"
	//"time"

	cga "github.com/guaychou/corona-api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)



var (
	confirmed = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "total_confirmed_corona",
		Help: "Current total confirmed corona",
	},
		[]string{"country"},
	)
	recovered = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "total_recovered_corona",
		Help: "Current total recovered corona",
	},
		[]string{"country"},
	)
	recoveryRate = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "recovery_rate_corona",
		Help: "Current recovery rate",
	},
		[]string{"country"},
	)
	death = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "total_death_corona",
		Help: "Current total death people",
	},
		[]string{"country"},
	)
	deathRate = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "death_rate_corona",
		Help: "Current fatality rate",
	},
		[]string{"country"},
	)
)

func init(){
	prometheus.MustRegister(confirmed)
	prometheus.MustRegister(recovered)
	prometheus.MustRegister(death)
	prometheus.MustRegister(deathRate)
	prometheus.MustRegister(recoveryRate)
}
func main() {
	countryPtr := flag.String("country", "", "Country name you want to get COVID19 status")
	addressPtr := flag.String("listen.address",":10198", "Port listen address")
	updateIntervalPtr := flag.Duration("update.interval",5 , "Update interval in minutes")
	flag.Parse()
	if *countryPtr=="" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	countrySplit:=strings.Split(*countryPtr,",")

	for _,value :=range(countrySplit){
		err:=checkCountry(value)
		if err!=nil{
			log.Fatal(err)
		}
	}

	for _,value :=range(countrySplit){
		go func(value string) {
			log.Info("Scrapping corona status in "+value)
			for {
				result:=get(value)
				log.Info("Country: "+value+" has been scrapped . . .")
				confirmed.WithLabelValues(value).Set(float64(result.Confirmed.Value))
				death.WithLabelValues(value).Set(float64((result.Deaths.Value)))
				recovered.WithLabelValues(value).Set(float64(result.Recovered.Value))
				recoveryRate.WithLabelValues(value).Set(result.CaseRecoveryRate)
				deathRate.WithLabelValues(value).Set(result.CaseFatalityRate)
				time.Sleep(*updateIntervalPtr * time.Minute)
			}

		}(value)
	}
	http.Handle("/metrics", promhttp.Handler())
	log.Println("Web Server started. Listening on address "+*addressPtr)
	log.Fatal(http.ListenAndServe(*addressPtr, nil))
}

func checkCountry(country string) error {
	result:=get(country)
	if result.Recovered.Value==-1 || result.Deaths.Value==-1 || result.Confirmed.Value==-1{
		return errors.New("Country "+country+" not found")
	}
	return nil
}

func get(country string)cga.CurrentCoronaStatus{
	result:=cga.GetCorona(country)
	return result
}
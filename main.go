package main

import (
	"errors"
	"flag"
	cga "github.com/guaychou/corona-api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strings"
	"time"
)
type Country struct {
	Confirmed prometheus.Gauge
	Recovered prometheus.Gauge
	Death prometheus.Gauge
	DeathRate prometheus.Gauge
	RecoveryRate prometheus.Gauge
}

type CountryMetrics struct {
	Confirmed prometheus.GaugeOpts
	Recovered prometheus.GaugeOpts
	Death prometheus.GaugeOpts
	DeathRate prometheus.GaugeOpts
	RecoveryRate prometheus.GaugeOpts
}

var MetricsInterface map[string] *CountryMetrics
var CountryInterface map[string] *Country


func main() {
	countryPtr := flag.String("country", "", "Country name you want to get COVID19 status.\nSeparate with comma ',' to use multiple country")
	addressPtr := flag.String("listen.address",":10198", "Port listen address")
	flag.Parse()
	if *countryPtr=="" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	MetricsInterface = make(map[string] *CountryMetrics)
	CountryInterface = make(map[string] *Country)
	countrySplit:=strings.Split(*countryPtr,",")
	for _,value :=range(countrySplit){
		err:=checkCountry(value)
		if err!=nil{
			log.Fatal(err)
		}
		setGaugeOpts(value)
		setGauge(value)
		registerGauge(value)
	}

	for key:=range CountryInterface{
		go func(key string) {
			log.Info("Scrapping corona status in "+key)
			for {
				result:=get(key)
				log.Info("Country: "+key+" has been scrapped . . .")
				CountryInterface[key].Confirmed.Set(float64(result.Confirmed.Value))
				CountryInterface[key].Death.Set(float64(result.Deaths.Value))
				CountryInterface[key].Recovered.Set(float64(result.Recovered.Value))
				CountryInterface[key].DeathRate.Set(result.CaseFatalityRate)
				CountryInterface[key].RecoveryRate.Set(result.CaseRecoveryRate)
				time.Sleep(5 * time.Second)
			}
		}(key)
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

func setGaugeOpts(param string){
	metrics:=new(CountryMetrics)
	metrics.Confirmed.Name="confirmed_corona_"+param
	metrics.Death.Name="death_corona_"+param
	metrics.Recovered.Name="recovered_corona_"+param
	metrics.DeathRate.Name="death_rate_corona_"+param
	metrics.RecoveryRate.Name="recovery_rate_corona_"+param
	MetricsInterface[param]=metrics
}

func setGauge(param string){
	country:=new(Country)
	country.Recovered=prometheus.NewGauge(MetricsInterface[param].Recovered)
	country.Death=prometheus.NewGauge(MetricsInterface[param].Death)
	country.Confirmed=prometheus.NewGauge(MetricsInterface[param].Confirmed)
	country.DeathRate=prometheus.NewGauge(MetricsInterface[param].DeathRate)
	country.RecoveryRate=prometheus.NewGauge(MetricsInterface[param].RecoveryRate)
	CountryInterface[param]=country
}

func registerGauge(param string){
	prometheus.MustRegister(CountryInterface[param].Confirmed)
	prometheus.MustRegister(CountryInterface[param].Recovered)
	prometheus.MustRegister(CountryInterface[param].Death)
	prometheus.MustRegister(CountryInterface[param].RecoveryRate)
	prometheus.MustRegister(CountryInterface[param].DeathRate)
}
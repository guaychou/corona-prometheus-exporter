# corona-prometheus-exporter
This is corona exporter for prometheus

Metrics | Description
------------- | -------------
confirmed_corona_\<CountryName>  | Current total confirmed positive COVID19
death_corona_\<CountryName>  | Current total death people in specified country
recovered_corona_\<CountryName> | Current total recovered people in specified country
recovery_rate_corona_\<CountryName> | Current recovery rate COVID19 in specified country
death_rate_corona_\<CountryName> | Current case fatality rate COVID19 in specified country

### How To Run
```cassandraql
$ ./corona-exporter --country=indonesia
```

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	hostRegexp = regexp.MustCompile("^[A-Za-z0-9.-]+$")

	// Our metrics
	awairScoreDesc = prometheus.NewDesc(
		"awair_awair_score", "Awair score.", []string{}, nil,
	)
	dewPointDesc = prometheus.NewDesc(
		"awair_dew_point_celsius", "Dew point.", []string{}, nil,
	)
	tempDesc = prometheus.NewDesc(
		"awair_temp_celsius", "Temperature.", []string{}, nil,
	)
	humidDesc = prometheus.NewDesc(
		"awair_relative_humidity", "Relative humidity.", []string{}, nil,
	)
	absHumidDesc = prometheus.NewDesc(
		"awair_absolute_humidity_grams_per_cubic_meter", "Absolute humidity.", []string{}, nil,
	)
	cO2Desc = prometheus.NewDesc(
		"awair_co2_parts_per_million", "CO2.", []string{}, nil,
	)
	cO2EstDesc = prometheus.NewDesc(
		"awair_co2_est_parts_per_million", "(Estimated?) CO2; unclear how this metric differs from CO2.", []string{}, nil,
	)
	vOCDesc = prometheus.NewDesc(
		"awair_voc_parts_per_billion", "VOC (Volatile organic compounds).", []string{}, nil,
	)
	vOCBaselineDesc = prometheus.NewDesc(
		"awair_voc_baseline", "Unknown, possibly unused?", []string{}, nil,
	)
	vOCH2RawDesc = prometheus.NewDesc(
		"awair_voc_h2_raw", "Unknown, possibly dihydrogen ppb?", []string{}, nil,
	)
	vOCEthanolRawDesc = prometheus.NewDesc(
		"awair_voc_ethanol_raw", "Unknown, possibly ethanol ppb?", []string{}, nil,
	)
	pM25Desc = prometheus.NewDesc(
		"awair_pm25_micrograms_per_cubic_meter", "Particulate matter (fine-dust).", []string{}, nil,
	)
	pM10EstDesc = prometheus.NewDesc(
		"awair_pm10_est_micrograms_per_cubic_meter", "Likely estimated particulate matter (big particles).", []string{}, nil,
	)

	// App-specific metrics
	collectTimeDesc = prometheus.NewDesc(
		"awair_scrape_duration_seconds", "Amount of time spent scraping metrics.", []string{}, nil,
	)
	errorsDesc = prometheus.NewDesc(
		"awair_scrape_errors", "How many errors occured during the scrape event.", []string{}, nil,
	)
)

// AwairCollector collects metrics from an Awair instance. Implements the Collector interface.
type AwairCollector struct {
	start time.Time
	url   *url.URL
}

// Describe is implemented by DescribeByCollect
func (ac AwairCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(ac, ch)
}

// Collect actually collects those sweet sweet metrics.
func (ac AwairCollector) Collect(ch chan<- prometheus.Metric) {
	errors := 0
	resp, err := ac.SampleApi()
	if err != nil {
		log.Printf("Error scraping metrics: %s\n", err.Error())
		errors++
	} else {
		ch <- prometheus.MustNewConstMetric(awairScoreDesc, prometheus.GaugeValue, float64(resp.Score))
		ch <- prometheus.MustNewConstMetric(dewPointDesc, prometheus.GaugeValue, float64(resp.DewPoint))
		ch <- prometheus.MustNewConstMetric(tempDesc, prometheus.GaugeValue, float64(resp.Temp))
		ch <- prometheus.MustNewConstMetric(humidDesc, prometheus.GaugeValue, float64(resp.Humid))
		ch <- prometheus.MustNewConstMetric(absHumidDesc, prometheus.GaugeValue, float64(resp.AbsHumid))
		ch <- prometheus.MustNewConstMetric(cO2Desc, prometheus.GaugeValue, float64(resp.CO2))
		ch <- prometheus.MustNewConstMetric(cO2EstDesc, prometheus.GaugeValue, float64(resp.CO2Est))
		ch <- prometheus.MustNewConstMetric(vOCDesc, prometheus.GaugeValue, float64(resp.VOC))
		ch <- prometheus.MustNewConstMetric(vOCBaselineDesc, prometheus.GaugeValue, float64(resp.VOCBaseline))
		ch <- prometheus.MustNewConstMetric(vOCH2RawDesc, prometheus.GaugeValue, float64(resp.VOCH2Raw))
		ch <- prometheus.MustNewConstMetric(vOCEthanolRawDesc, prometheus.GaugeValue, float64(resp.VOCEthanolRaw))
		ch <- prometheus.MustNewConstMetric(pM25Desc, prometheus.GaugeValue, float64(resp.PM25))
		ch <- prometheus.MustNewConstMetric(pM10EstDesc, prometheus.GaugeValue, float64(resp.PM10Est))
	}
	// App-specific metrics
	ch <- prometheus.MustNewConstMetric(collectTimeDesc, prometheus.GaugeValue, time.Since(ac.start).Seconds())
	ch <- prometheus.MustNewConstMetric(errorsDesc, prometheus.GaugeValue, float64(errors))

}

func (ac AwairCollector) SampleApi() (*AwairAirDataResponse, error) {
	resp, err := http.Get(ac.url.String())
	if err != nil {
		return nil, fmt.Errorf("Unable to poll url %s: %w", ac.url.String(), err)
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	var awairResponse AwairAirDataResponse
	if err := dec.Decode(&awairResponse); err != nil {
		return nil, fmt.Errorf("Invalid response: %w", err)
	}

	return &awairResponse, nil
}

type AwairAirDataResponse struct {
	// ISO 8601, example: "2020-08-09T05:35:28.034Z"
	Timestamp string  `json:"timestamp"`
	Score     float64 `json:"score"`
	DewPoint  float64 `json:"dew_point"`
	Temp      float64 `json:"temp"`
	Humid     float64 `json:"humid"`
	AbsHumid  float64 `json:"abs_humid"`
	CO2       float64 `json:"co2"`
	CO2Est    float64 `json:"co2_est"`
	VOC       float64 `json:"voc"`
	// Example value: 2352254740, unsure why so high
	VOCBaseline   float64 `json:"voc_baseline"`
	VOCH2Raw      float64 `json:"voc_h2_raw"`
	VOCEthanolRaw float64 `json:"voc_ethanol_raw"`
	PM25          float64 `json:"pm25"`
	PM10Est       float64 `json:"pm10_est"`
}

func AwairHandler() http.Handler {
	return http.HandlerFunc(func(rsp http.ResponseWriter, req *http.Request) {
		// Parse host address
		h := req.URL.Query().Get("host")
		if h == "" {
			rsp.WriteHeader(http.StatusBadRequest)
			_, err := rsp.Write([]byte("host query parameter is required"))
			if err != nil {
				log.Printf("Error writing response: %s", err.Error())
			}
			return
		}
		if hostRegexp.FindString(h) == "" {
			rsp.WriteHeader(http.StatusBadRequest)
			_, err := rsp.Write([]byte(fmt.Sprintf("host query parameter does not match valid hostname: %s", hostRegexp.String())))
			if err != nil {
				log.Printf("Error writing response: %s", err.Error())
			}
			return
		}
		u := &url.URL{
			Scheme: "http",
			Host:   h,
			Path:   "air-data/latest",
		}

		// TODO probably pretty wasteful to create these dynamically
		reg := prometheus.NewRegistry()
		ac := &AwairCollector{url: u, start: time.Now()}
		reg.MustRegister(ac)

		promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP(rsp, req)
	})
}

var addr = flag.String("listen-address", ":8123", "The address to listen on for HTTP requests.")

func main() {
	flag.Parse()
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/awair", AwairHandler())
	log.Printf("Listening for http connections on %s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

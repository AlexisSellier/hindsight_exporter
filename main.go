package main
import (
	"flag"
	"os"
	"bufio"
	"strings"
	"net/http"
	"fmt"
	"log"
	"strconv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)


type hindsight struct {
	filename string
	metrics  map[string]hindsightMetric
}

type hindsightMetric struct {
	desc    *prometheus.Desc
	valType prometheus.ValueType
	index   int
}

var (
	listenAddress		= flag.String("web.listen-address", ":9121", "Address to listen on for web interface and telemetry.")
	metricPath		= flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	hindsightStatsPath	= flag.String("hindsight", "/hindsight.tsv", "hindsight.tsv file location.")
)


func newHindsightExporter(file string) *hindsight {
	e := hindsight{
		filename: file,
		metrics: map[string]hindsightMetric{
			"hindsight_injected_message_count": {
				desc:    prometheus.NewDesc("hindsight_injected_message_count", "Number of injected messages", []string{"plugin"}, nil),
				valType: prometheus.GaugeValue,
				index: 1,
			},
			"hindsight_injected_message_bytes": {
				desc:    prometheus.NewDesc("hindsight_injected_message_bytes", "Number of injected bytes", []string{"plugin"}, nil),
				valType: prometheus.GaugeValue,
				index: 2,
			},
			"hindsight_process_message_count": {
				desc:    prometheus.NewDesc("hindsight_process_message_count", "Number of processed messages", []string{"plugin"}, nil),
				valType: prometheus.GaugeValue,
				index:3,
			},
			"hindsight_process_message_failures": {
				desc:    prometheus.NewDesc("hindsight_process_message_failures", "Number of processed failure messages", []string{"plugin"}, nil),
				valType: prometheus.GaugeValue,
				index: 4,
			},
		},
	}
	return &e
}


func (h *hindsight) Describe(ch chan<- *prometheus.Desc) {
	for _, i := range h.metrics {
		ch <- i.desc
	}
}

func (h *hindsight) Collect(ch chan<- prometheus.Metric) {
	fmt.Println("DONCHED");
	h.fetchHindsightStatistcs(ch)
}

func parseFloatOrZero(s string) float64 {
	res, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Printf("Cannot parse %s\n", s)
		return 0.0
	}
	return res
}

func main() {
	flag.Parse()
	e := newHindsightExporter(*hindsightStatsPath)
	prometheus.MustRegister(e)
	http.Handle(*metricPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
		       <head><title>Warp10 exporter</title></head>
		       <body>
		       <h1>Warp10 exporter</h1>
		       <p><a href='` + *metricPath + `'>Metrics</a></p>
		       </body>
		       </html>`))
	})
	log.Printf("providing metrics at %s%s", *listenAddress, *metricPath)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
func (h *hindsight) fetchHindsightStatistcs(ch chan<- prometheus.Metric) {
	file, err := os.Open(h.filename)
	defer file.Close()
	if err == nil {
		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines) 
		scanner.Scan()
		scanner.Text()
		for scanner.Scan() {
			values := strings.Fields(scanner.Text())
			for _, e := range h.metrics {
				ch <- prometheus.MustNewConstMetric(e.desc, e.valType, parseFloatOrZero(values[e.index]), values[0])
			}
		}
	} else {
		log.Printf("Cannot output %s hindsight statistics", h.filename)
	}
}

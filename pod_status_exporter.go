package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os/exec"
	"strings"
)

type PodCollector struct {

	PodMetric *prometheus.Desc
}

func (collector *PodCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.PodMetric
}


var vLable = []string{}

var vValue = []string{}

var constLabel = prometheus.Labels{"component": "pod"}


func newPodCollector() *PodCollector {
	var rm = make(map[string]string)
	rm = getPodStatus()
	if _, ok := rm["msg"]; ok {
		log.Error("command execute failed：", rm["msg"])
	} else {

		for k := range rm {

			vLable = append(vLable, k)
		}
	}

	return &PodCollector{
		PodMetric: prometheus.NewDesc("pod_status",
			"Show pod status", vLable,
			constLabel),
	}
}

func getPodStatus() (m map[string]string) {
	result,err:= exec.Command("bash", "-c", "/Users/wanghaoyang/.rd/bin/kubectl get pod -A|awk 'NR!=1 {print $2\":\"$4}'").Output()
	if err != nil {
		log.Error("result: ", string(result))
		m = make(map[string]string)
		m["msg"] = "failure"
		return m
	} else if len(result) == 0 {
		log.Errorf("command exec failed because result is nil")
		m = make(map[string]string)
		m["msg"] = "return nil"
		return m
	}
	ret := strings.TrimSpace(string(result))
	tt := strings.Split(ret, "\n")

	var nMap = make(map[string]string)
	for i:=0; i<len(tt);i++{
		if strings.Contains(tt[i], ":") {
			nKey := strings.Split(tt[i], ":")[0]
			nKey = strings.ReplaceAll(nKey, "-", "_")
			nValue := strings.Split(tt[i], ":")[1]
			nMap[nKey] = nValue
		}
	}
	//fmt.Println("打印map: ")
	//for key, value := range nMap{
	//	fmt.Println(key)
	//	fmt.Println(value)
	//}
	return nMap
}

func (collector *PodCollector) Collect(ch chan<- prometheus.Metric) {
	var metricValue float64
	var rm = make(map[string]string)
	rm = getPodStatus()
	if _, ok := rm["msg"]; ok {
		log.Error("command exec failed")
		metricValue = 5
		ch <- prometheus.MustNewConstMetric(collector.PodMetric, prometheus.CounterValue, metricValue)
	} else {
		//vValue = vValue[0:0]

		for _, v := range rm {

			vValue = append(vValue, v)

			if v == "Completed" {
				metricValue++
			}
		}
		ch <- prometheus.MustNewConstMetric(collector.PodMetric, prometheus.CounterValue, metricValue, vValue...)
	}
}

func main() {
	pod := newPodCollector()
	prometheus.MustRegister(pod)

	http.Handle("/metrics", promhttp.Handler())

	log.Info("begin to server on port 8080")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
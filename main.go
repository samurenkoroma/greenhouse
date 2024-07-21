package main

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"math"
	"net/http"
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}
var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func main() {

	// Create a non-global registry.
	reg := prometheus.NewRegistry()

	// Create new metrics and register them using the custom registry.
	m := NewMetrics(reg)
	var broker = "my.dom"
	var port = 1883

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID("tiun_client[go]")
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		var sensors []Sensor
		json.Unmarshal(msg.Payload(), &sensors)
		for _, v := range sensors {

			m.sensors.With(prometheus.Labels{"name": v.Name, "type": v.Type}).Set(math.Round(v.Value*10) / 10)
		}
	})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	go sub(client)
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	http.HandleFunc("/api", apiHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func apiHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("{\"status\": \"ok\"}"))
}

func sub(client mqtt.Client) {
	topic := "tiun/base"
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Printf("Subscribed to topic: %s \n", topic)
}

type Sensor struct {
	Name  string  `json:"name"`
	Type  string  `json:"type"`
	Value float64 `json:"value"`
}
type metrics struct {
	sensors *prometheus.GaugeVec
}

func NewMetrics(reg prometheus.Registerer) *metrics {
	m := &metrics{
		sensors: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "sensors",
				Help: "Текущая показания",
			},
			[]string{"name", "type"},
		)}
	reg.MustRegister(m.sensors)
	return m
}

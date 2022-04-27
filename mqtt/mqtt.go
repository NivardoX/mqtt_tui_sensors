package mqtt

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"os"
)

type Handler struct {
	client mqtt.Client
	topic  string
}

func getEnv(key string, fallback string) string {
	var env = os.Getenv(key)
	if env == "" {
		env = fallback
	}
	return env
}
func NewMqttHandler(topic string, onMessage func(data sensorData)) *Handler {
	var broker = getEnv("MQTT_BROKER_URL", "localhost")
	var port = getEnv("MQTT_BROKER_PORT", "1883")
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%s", broker, port))
	opts.SetDefaultPublishHandler((*Handler)(nil).getMessagePubHandler(topic, onMessage))
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	handler := Handler{
		client: client,
		topic:  topic}
	return &handler
}

func (c *Handler) getMessagePubHandler(topic string, onMessage func(data sensorData)) mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		if msg.Topic() == topic {
			data := &sensorData{}
			err := json.Unmarshal(msg.Payload(), data)
			if err != nil {
				return
			}
			onMessage(*data)

		}
	}

}

// sub subscribes to the topic
func (c *Handler) sub() {
	token := c.client.Subscribe(c.topic, 0, nil)
	token.Wait()
}

// pub publishes a message to the topic
func (c *Handler) pub(data sensorData) {
	dataAsJson, _ := json.Marshal(data)
	token := c.client.Publish(c.topic, 0, false, string(dataAsJson))
	token.Wait()
}

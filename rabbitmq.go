package cfutil

import (
	"errors"
	"fmt"
	cfenv "github.com/cloudfoundry-community/go-cfenv"
	"github.com/mitchellh/mapstructure"
	"github.com/streadway/amqp"
	"log"
)

func RabbitMQAdminURI(serviceName string) (string, error) {
	appEnv, _ := Current()
	service := &cfenv.Service{}
	err := errors.New("")
	if serviceName != "" {
		service, err = serviceByName(appEnv, serviceName)
	} else {
		service, err = firstMatchingService(appEnv, "amqp")
	}
	if err != nil {
		return "", err
	}
	protocols := map[string]interface{}{}
	management := map[string]string{}

	if service.Credentials["protocols"] != nil {
		err := mapstructure.Decode(service.Credentials["protocols"], &protocols)
		if err != nil {
			return "", fmt.Errorf("Error decoding protocols section")
		}
		if protocols["management"] != nil {
			err = mapstructure.Decode(protocols["management"], &management)
			if err != nil {
				return "", fmt.Errorf("Error decoding management section")
			}
			return management["uri"], nil
		}
		return "", fmt.Errorf("No mangement section defined")
	}
	return "", fmt.Errorf("No protocols section defined")
}

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	tag     string
	done    chan error
}

func NewConsumer(serviceName, exchange, exchangeType, queueName, key, ctag string, handlerFunc ConsumerHandlerFunc) (*Consumer, error) {
	connectString := ""
	var err error
	appEnv, _ := Current()
	if serviceName != "" {
		connectString, err = serviceURIByName(appEnv, serviceName)
	} else {
		connectString, err = firstMatchingServiceURI(appEnv, "amqp")
	}

	c := &Consumer{
		conn:    nil,
		channel: nil,
		tag:     ctag,
		done:    make(chan error),
	}

	log.Printf("dialing %q", connectString)
	c.conn, err = amqp.Dial(connectString)
	if err != nil {
		return nil, fmt.Errorf("Dial: %s", err)
	}

	go func() {
		fmt.Printf("closing: %s", <-c.conn.NotifyClose(make(chan *amqp.Error)))
	}()

	log.Printf("got Connection, getting Channel")
	c.channel, err = c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("Channel: %s", err)
	}

	err = c.channel.Qos(2, 0, false)
	if err != nil {
		log.Printf("error setting Qos: %s", err)
	}

	log.Printf("got Channel, declaring Exchange (%q)", exchange)
	if err = c.channel.ExchangeDeclare(
		exchange,     // name of the exchange
		exchangeType, // type
		true,         // durable
		false,        // delete when complete
		false,        // internal
		false,        // noWait
		nil,          // arguments
	); err != nil {
		return nil, fmt.Errorf("Exchange Declare: %s", err)
	}

	log.Printf("declared Exchange, declaring Queue %q", queueName)
	queue, err := c.channel.QueueDeclare(
		queueName, // name of the queue
		true,      // durable
		false,     // delete when usused
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Queue Declare: %s", err)
	}

	log.Printf("declared Queue (%q %d messages, %d consumers), binding to Exchange (key %q)",
		queue.Name, queue.Messages, queue.Consumers, key)

	if err = c.channel.QueueBind(
		queue.Name, // name of the queue
		key,        // bindingKey
		exchange,   // sourceExchange
		false,      // noWait
		nil,        // arguments
	); err != nil {
		return nil, fmt.Errorf("Queue Bind: %s", err)
	}

	log.Printf("Queue bound to Exchange, starting Consume (consumer tag %q)", c.tag)
	deliveries, err := c.channel.Consume(
		queue.Name, // name
		c.tag,      // consumerTag,
		false,      // noAck
		false,      // exclusive
		false,      // noLocal
		false,      // noWait
		nil,        // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Queue Consume: %s", err)
	}

	go handle(handlerFunc, deliveries, c.done)

	return c, nil
}

func (c *Consumer) Shutdown() error {
	// will close() the deliveries channel
	if err := c.channel.Cancel(c.tag, true); err != nil {
		return fmt.Errorf("Consumer cancel failed: %s", err)
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("AMQP connection close error: %s", err)
	}

	defer log.Printf("AMQP shutdown OK")

	// wait for handle() to exit
	return <-c.done
}

type ConsumerHandlerFunc func(delivery amqp.Delivery) error

func DummyConsumerHandler(d amqp.Delivery) error {
	log.Printf(
		"got %dB delivery: [%v] %q",
		len(d.Body),
		d.DeliveryTag,
		d.Body,
	)
	d.Ack(false)
	return nil
}

func handle(handler ConsumerHandlerFunc, deliveries <-chan amqp.Delivery, done chan error) {
	for d := range deliveries {
		handler(d)
	}
	log.Printf("handle: deliveries channel closed")
	done <- nil
}

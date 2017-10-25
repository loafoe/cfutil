package cfutil

import (
	"context"
	"errors"
	"fmt"
	"time"

	cfenv "github.com/cloudfoundry-community/go-cfenv"
	"github.com/mitchellh/mapstructure"
	"github.com/streadway/amqp"
)

func RabbitMQAdminURI(serviceName string) (string, error) {
	if IsLocal() {
		return "http://guest:guest@localhost:15672", nil
	}

	appEnv, _ := Current()
	var service *cfenv.Service
	var err error
	if serviceName != "" {
		service, err = serviceByName(appEnv, serviceName)
	} else {
		service, err = firstMatchingService(appEnv, "amqp")
	}
	if err != nil {
		return "", err
	}
	if service == nil {
		return "", errors.New("RabbitMQ serice not found")
	}
	protocols := map[string]interface{}{}
	management := map[string]interface{}{}

	if service.Credentials["protocols"] != nil {
		err := mapstructure.Decode(service.Credentials["protocols"], &protocols)
		if err != nil {
			return "", fmt.Errorf("Error decoding protocols section: %s", err.Error())
		}
		if protocols["management"] != nil {
			err = mapstructure.Decode(protocols["management"], &management)
			if err != nil {
				return "", fmt.Errorf("Error decoding management section: %s", err.Error())
			}
			if management["uri"] != nil {
				str, ok := management["uri"].(string)
				if !ok {
					return "", fmt.Errorf("Management URI not a string")
				}
				return str, nil
			}
			return "", fmt.Errorf("Management URI not found")
		}
		return "", fmt.Errorf("No mangement section defined")
	}
	return "", fmt.Errorf("No protocols section defined")
}

type Consumer struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	handlerFunc  ConsumerHandlerFunc
	done         chan error
	consumerTag  string // Name that consumer identifies itself to the server with
	uri          string // uri of the rabbitmq server
	exchange     string // exchange that we will bind to
	exchangeType string // topic, direct, etc...
	bindingKey   string // routing key that we are using
	queueName    string // queue name
	log          Logger
}

type Producer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	done    chan error
	log     Logger
}

type ProducerConfig struct {
	ServiceName  string
	Exchange     string
	ExchangeType string
	log          Logger
}

type ConsumerConfig struct {
	ServiceName  string
	Exchange     string
	ExchangeType string
	QueueName    string
	RoutingKey   string
	CTag         string
	HandlerFunc  ConsumerHandlerFunc
	log          Logger
}

func (p *Producer) Publish(exchange, routingKey string, msg amqp.Publishing) error {
	if err := p.channel.Publish(
		exchange,   // publish to an exchange
		routingKey, // routing to 0 or more queues
		false,      // mandatory
		false,      // immediate
		msg,
	); err != nil {
		return fmt.Errorf("Exchange Publish: %s", err)
	}
	return nil
}

func (p *Producer) Close() {
	p.conn.Close()
}

func NewProducer(config ProducerConfig) (*Producer, error) {
	var connectString string
	var err error
	appEnv, _ := Current()
	if config.ServiceName != "" {
		connectString, _ = serviceURIByName(appEnv, config.ServiceName)
	} else {
		connectString, _ = firstMatchingServiceURI(appEnv, "amqp")
	}

	p := &Producer{
		conn:    nil,
		channel: nil,
		done:    make(chan error),
		log:     config.log,
	}
	if p.log == nil {
		p.log = defaultLogger
	}

	p.conn, err = amqp.Dial(connectString)
	if err != nil {
		return nil, fmt.Errorf("Dial: %s", err)
	}
	p.channel, err = p.conn.Channel()
	if err != nil {
		p.conn.Close()
		return nil, fmt.Errorf("Channel: %s", err)
	}
	if err = p.channel.ExchangeDeclare(
		config.Exchange,     // name
		config.ExchangeType, // type
		true,                // durable
		false,               // auto-deleted
		false,               // internal
		false,               // noWait
		nil,                 // arguments
	); err != nil {
		p.conn.Close()
		return nil, fmt.Errorf("Exchange Declare: %s", err)
	}
	return p, nil
}

func NewConsumer(config ConsumerConfig) (*Consumer, error) {
	var connectString string
	var err error
	appEnv, _ := Current()
	if config.ServiceName != "" {
		connectString, err = serviceURIByName(appEnv, config.ServiceName)
	} else {
		connectString, err = firstMatchingServiceURI(appEnv, "amqp")
	}

	c := &Consumer{
		conn:         nil,
		channel:      nil,
		uri:          connectString,
		handlerFunc:  config.HandlerFunc,
		consumerTag:  config.CTag,
		exchange:     config.Exchange,
		exchangeType: config.ExchangeType,
		bindingKey:   config.RoutingKey,
		queueName:    config.QueueName,
		done:         make(chan error),
		log:          config.log,
	}
	if c.log == nil {
		c.log = defaultLogger
	}
	return c, err
}

func (c *Consumer) Start() error {
	if err := c.Connect(); err != nil {
		return err
	}

	deliveries, err := c.AnnounceQueue(c.queueName, c.bindingKey)
	if err != nil {
		return err
	}
	go c.Handle(deliveries, c.handlerFunc, 1, c.queueName, c.bindingKey)
	return nil
}

// ReConnect is called in places where NotifyClose() channel is called
// wait 30 seconds before trying to reconnect. Any shorter amount of time
// will  likely destroy the error log while waiting for servers to come
// back online. This requires two parameters which is just to satisfy
// the AccounceQueue call and allows greater flexability
func (c *Consumer) ReConnect(queueName, bindingKey string) (<-chan amqp.Delivery, error) {
	time.Sleep(30 * time.Second)

	if err := c.Connect(); err != nil {
		c.log.Info(context.TODO(), "Could not connect in reconnect call: %v", err.Error())
	}

	deliveries, err := c.AnnounceQueue(queueName, bindingKey)
	if err != nil {
		return deliveries, errors.New("Couldn't connect")
	}

	return deliveries, nil
}

// Connect to RabbitMQ server
func (c *Consumer) Connect() error {

	var err error

	c.conn, err = amqp.Dial(c.uri)
	if err != nil {
		return fmt.Errorf("Dial: %s", err)
	}

	go func() {
		// Waits here for the channel to be closed
		c.log.Info(context.TODO(), "closing: %s", <-c.conn.NotifyClose(make(chan *amqp.Error)))
		// Let Handle know it's not time to reconnect
		c.done <- errors.New("Channel Closed")
	}()

	c.log.Info(context.TODO(), "got Connection, getting Channel")
	c.channel, err = c.conn.Channel()
	if err != nil {
		return fmt.Errorf("Channel: %s", err)
	}

	c.log.Info(context.TODO(), "got Channel, declaring Exchange (%q)", c.exchange)
	if err = c.channel.ExchangeDeclare(
		c.exchange,     // name of the exchange
		c.exchangeType, // type
		true,           // durable
		false,          // delete when complete
		false,          // internal
		false,          // noWait
		nil,            // arguments
	); err != nil {
		return fmt.Errorf("Exchange Declare: %s", err)
	}

	return nil
}

// AnnounceQueue sets the queue that will be listened to for this
// connection...
func (c *Consumer) AnnounceQueue(queueName, bindingKey string) (<-chan amqp.Delivery, error) {
	c.log.Info(context.TODO(), "declared Exchange, declaring Queue %q", queueName)
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

	c.log.Info(context.TODO(), "declared Queue (%q %d messages, %d consumers), binding to Exchange (key %q)",
		queue.Name, queue.Messages, queue.Consumers, bindingKey)

	// Qos determines the amount of messages that the queue will pass to you before
	// it waits for you to ack them. This will slow down queue consumption but
	// give you more certainty that all messages are being processed. As load increases
	// I would reccomend upping the about of Threads and Processors the go process
	// uses before changing this although you will eventually need to reach some
	// balance between threads, procs, and Qos.
	err = c.channel.Qos(1, 0, false)
	if err != nil {
		return nil, fmt.Errorf("Error setting qos: %s", err)
	}

	if err = c.channel.QueueBind(
		queue.Name, // name of the queue
		bindingKey, // bindingKey
		c.exchange, // sourceExchange
		false,      // noWait
		nil,        // arguments
	); err != nil {
		return nil, fmt.Errorf("Queue Bind: %s", err)
	}

	c.log.Info(context.TODO(), "Queue bound to Exchange, starting Consume (consumer tag %q)", c.consumerTag)
	deliveries, err := c.channel.Consume(
		queue.Name,    // name
		c.consumerTag, // consumerTag,
		false,         // noAck
		false,         // exclusive
		false,         // noLocal
		false,         // noWait
		nil,           // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Queue Consume: %s", err)
	}

	return deliveries, nil
}

// Handle has all the logic to make sure your program keeps running
// d should be a delievey channel as created when you call AnnounceQueue
// fn should be a function that handles the processing of deliveries
// this should be the last thing called in main as code under it will
// become unreachable unless put int a goroutine. The q and rk params
// are redundent but allow you to have multiple queue listeners in main
// without them you would be tied into only using one queue per connection
func (c *Consumer) Handle(
	d <-chan amqp.Delivery,
	fn ConsumerHandlerFunc,
	threads int,
	queue string,
	routingKey string) {

	var err error

	for {
		for i := 0; i < threads; i++ {
			go fn(d)
		}

		// Go into reconnect loop when
		// c.done is passed non nil values
		if <-c.done != nil {
			d, err = c.ReConnect(queue, routingKey)
			if err != nil {
				// Very likely chance of failing
				// should not cause worker to terminate
				c.log.Error(context.TODO(), "Reconnecting Error: %s", err)
			}
		}
		c.log.Info(context.TODO(), "Reconnected... possibly")
	}
}

type ConsumerHandlerFunc func(deliveries <-chan amqp.Delivery) error

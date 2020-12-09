package kafkaex

import (
	"context"
	"fmt"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"sercomm.com/demeter/commons/datetime"
	"time"
)

type KafkaProducer struct {
	terminated           chan bool
	producer             *kafka.Producer
	kafkaProducerHandler KafkaProducerHandler
	Properties           map[string]interface{}
}

type KafkaProducerHandler interface {
	// "err" indicates if the message was delivered successfully or not
	MessageDeliveredResult(
		kafkaProducer *KafkaProducer,
		err error,
		topic string,
		partition int32,
		offset string,
		message string)
	// these errors should generally be considered informational
	// the client will try to automatically recover
	ErrorOccurred(
		kafkaProducer *KafkaProducer,
		errorCode kafka.ErrorCode,
		errorMessage string)
	// kafka session was terminated
	Closed(kafkaProducer *KafkaProducer)
}

func NewKafkaProducer(brokerAddr string, kafkaProducerHandler KafkaProducerHandler) (*KafkaProducer, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":  brokerAddr,
		"session.timeout.ms": 6000, // 6s
	})

	if nil != err {
		return nil, fmt.Errorf("FAILED TO CREATE PRODUCER: %+v", err)
	}

	kafkaProducer := &KafkaProducer{
		terminated:           make(chan bool),
		producer:             producer,
		kafkaProducerHandler: kafkaProducerHandler,
		Properties:           make(map[string]interface{}),
	}

	go func() {
		defer func() {
			if nil != kafkaProducer.kafkaProducerHandler {
				kafkaProducer.kafkaProducerHandler.Closed(kafkaProducer)
			}
		}()

		defer close(kafkaProducer.terminated)

		for event := range kafkaProducer.producer.Events() {
			switch entity := event.(type) {
			case *kafka.Message:
				if nil != kafkaProducer.kafkaProducerHandler {

					kafkaProducer.kafkaProducerHandler.MessageDeliveredResult(
						kafkaProducer,
						entity.TopicPartition.Error,
						*entity.TopicPartition.Topic,
						entity.TopicPartition.Partition,
						entity.TopicPartition.Offset.String(),
						string(entity.Value))

					if nil != entity.Opaque {
						close((entity.Opaque).(chan bool))
					}
				}
			case kafka.Error:
				if nil != kafkaProducer.kafkaProducerHandler {
					kafkaProducer.kafkaProducerHandler.ErrorOccurred(
						kafkaProducer,
						entity.Code(),
						entity.Error())
				}
			}
		}
	}()

	return kafkaProducer, nil
}

func (kafkaProducer *KafkaProducer) Close() {
	if nil != kafkaProducer.producer {
		kafkaProducer.producer.Close()
	}

	_ = <-kafkaProducer.terminated
}

func (kafkaProducer *KafkaProducer) CreateTopic(topic string, partitionCount int) error {
	if nil == kafkaProducer.producer {
		return fmt.Errorf("INVALID INSTANCE")
	}

	adminClient, err := kafka.NewAdminClientFromProducer(kafkaProducer.producer)
	if nil != err {
		return err
	}
	defer adminClient.Close()

	// contexts are used to abort or limit the amount of time
	// the admin-call blocks waiting for a result
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// duration is 5s defined by Cloud
	maxDuration := (time.Second * 5)

	results, err := adminClient.CreateTopics(
		ctx,
		// multiple topics can be created simultaneously
		// by providing more TopicSpecification structs here.
		[]kafka.TopicSpecification{{
			Topic:             topic,
			NumPartitions:     partitionCount,
			ReplicationFactor: 1}},
		kafka.SetAdminOperationTimeout(maxDuration))

	if err != nil {
		return fmt.Errorf("FAILED TO CREATE TOPIC: %+v", err)
	}

	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError {
			err = fmt.Errorf("ERROR RETURNED WHEN CREATING TOPIC: %+v", result.Error)
			return err
		}
	}

	return nil
}

func (kafkaProducer *KafkaProducer) DescribeTopic(topic string) (map[string]kafka.ConfigEntryResult, error) {
	if nil == kafkaProducer.producer {
		return nil, fmt.Errorf("INVALID INSTANCE")
	}

	adminClient, err := kafka.NewAdminClientFromProducer(kafkaProducer.producer)
	if nil != err {
		return nil, err
	}
	defer adminClient.Close()

	// contexts are used to abort or limit the amount of time
	// the admin-call blocks waiting for a result
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// duration is 5s defined by Cloud
	maxDuration := (time.Second * 5)

	// query cluster for the resource's current configuration
	results, err := adminClient.DescribeConfigs(
		ctx,
		[]kafka.ConfigResource{{Type: kafka.ResourceTopic, Name: topic}},
		kafka.SetAdminRequestTimeout(maxDuration))

	if nil != err {
		return nil, fmt.Errorf("FAILED TO DESCRIBE TOPIC: %+v", err)
	}

	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError {
			err = fmt.Errorf("ERROR RETURNED WHEN DESCRIBING TOPIC: %+v", result.Error)
			return nil, err
		}

		return result.Config, nil
	}

	return nil, fmt.Errorf("FAILED TO DESCRIBE TOPIC: RESULTS IS EMPTY")
}

func (kafkaProducer *KafkaProducer) DeleteTopic(topic string) error {
	if nil == kafkaProducer.producer {
		return fmt.Errorf("INVALID INSTANCE")
	}

	adminClient, err := kafka.NewAdminClientFromProducer(kafkaProducer.producer)
	if nil != err {
		return err
	}
	defer adminClient.Close()

	// contexts are used to abort or limit the amount of time
	// the admin-call blocks waiting for a result
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// duration is 60s defined by Cloud
	maxDuration := (time.Second * 60)

	topics := []string{
		topic,
	}

	results, err := adminClient.DeleteTopics(ctx, topics, kafka.SetAdminOperationTimeout(maxDuration))
	if err != nil {
		return fmt.Errorf("FAILED TO DELETE TOPIC: %+v", err)
	}

	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError {
			err = fmt.Errorf("ERROR RETURNED WHEN CREATING TOPIC: %+v", result.Error)
			return err
		}
	}

	return nil
}

func (kafkaProducer *KafkaProducer) DeliverMessage(
	topic string,
	partition int32,
	senderID string,
	receiverID string,
	messageType MessageType,
	message string,
	expirationInterval int64) error {

	if nil == kafkaProducer.producer {
		return fmt.Errorf("INVALID INSTANCE")
	}

	dateTime := datetime.Now()
	deliveryTime := dateTime.UnixTimestamp()
	expirationTime := deliveryTime + expirationInterval

	headers := []kafka.Header{
		{Key: HeaderSenderID, Value: []byte(senderID)},
		{Key: HeaderReceiverID, Value: []byte(receiverID)},
		{Key: HeaderDeliveryTime, Value: []byte(int64ToBytes(deliveryTime))},
		{Key: HeaderExpirationTime, Value: []byte(int64ToBytes(expirationTime))},
		{Key: HeaderMessageType, Value: []byte(messageType.String())},
	}

	blocker := make(chan bool)
	kafkaProducer.producer.ProduceChannel() <- &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: partition},
		Value:          []byte(message),
		Opaque:         blocker,
		Headers:        headers,
	}

	_ = <-blocker

	return nil
}

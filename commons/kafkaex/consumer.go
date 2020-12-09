package kafkaex

import (
	"context"
	"fmt"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"time"
)

type KafkaConsumer struct {
	terminated           bool
	consumer             *kafka.Consumer
	kafkaConsumerHandler KafkaConsumerHandler
	Properties           map[string]interface{}
}

type KafkaConsumerHandler interface {
	// receive message
	MessageReceived(
		kafkaConsumer *KafkaConsumer,
		topic string,
		partition int32,
		offset string,
		senderID string,
		receiverID string,
		deliveryTime int64,
		expirationTime int64,
		messageType MessageType,
		message string)
	// end of the partition
	PartitionEOF(
		kafkaConsumer *KafkaConsumer,
		topic string,
		partition int32,
		offset string)
	// errors should generally be considered informational
	// the client will try to automatically recover
	ErrorOccurred(
		kafkaConsumer *KafkaConsumer,
		errorCode kafka.ErrorCode,
		errorMessage string)
	// kafka session was terminated
	Closed(kafkaConsumer *KafkaConsumer)
}

func NewKafkaConsumer(brokerAddr string, groupName string, kafkaConsumerHandler KafkaConsumerHandler) (*KafkaConsumer, error) {
	// create kafka.Consumer
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":               brokerAddr,
		"group.id":                        groupName,
		"session.timeout.ms":              6000,
		"go.application.rebalance.enable": true,
		"enable.partition.eof":            true, // enable generation of PartitionEOF when the end of a partition is reached.
		"auto.offset.reset":               "latest",
	})

	if nil != err {
		return nil, fmt.Errorf("FAILED TO CREATE CONSUMER: %+v", err)
	}

	kafkaConsumer := &KafkaConsumer{
		terminated:           false,
		consumer:             consumer,
		kafkaConsumerHandler: kafkaConsumerHandler,
		Properties:           make(map[string]interface{}),
	}

	// goruntime: handle events
	go func() {
		defer func() {
			if nil != kafkaConsumer.kafkaConsumerHandler {
				kafkaConsumer.kafkaConsumerHandler.Closed(kafkaConsumer)
			}
		}()

		for kafkaConsumer.terminated == false {
			event := consumer.Poll(100)
			if nil == event {
				continue
			}

			switch entity := event.(type) {
			case kafka.PartitionEOF:
				if nil != kafkaConsumer.kafkaConsumerHandler {
					kafkaConsumer.kafkaConsumerHandler.PartitionEOF(
						kafkaConsumer,
						*entity.Topic,
						entity.Partition,
						entity.Offset.String())
				}
			case kafka.AssignedPartitions:
				kafkaConsumer.consumer.Assign(entity.Partitions)
			case kafka.RevokedPartitions:
				kafkaConsumer.consumer.Unassign()
			case kafka.Error:
				if nil != kafkaConsumer.kafkaConsumerHandler {
					kafkaConsumer.kafkaConsumerHandler.ErrorOccurred(
						kafkaConsumer,
						entity.Code(),
						entity.Error())
				}
			case *kafka.Message:
				if nil != kafkaConsumer.kafkaConsumerHandler {

					var senderID, receiverID string
					var deliveryTime int64 = 0
					var expirationTime int64 = 0
					var messageType MessageType = MessageTypeInvalid

					for _, header := range entity.Headers {
						switch header.Key {
						case HeaderSenderID:
							senderID = string(header.Value)
						case HeaderReceiverID:
							receiverID = string(header.Value)
						case HeaderDeliveryTime:
							deliveryTime = bytesToInt64(header.Value)
						case HeaderExpirationTime:
							expirationTime = bytesToInt64(header.Value)
						case HeaderMessageType:
							messageType = ParseMessageTypeString(string(header.Value))
						}
					}

					kafkaConsumer.kafkaConsumerHandler.MessageReceived(
						kafkaConsumer,
						*entity.TopicPartition.Topic,
						entity.TopicPartition.Partition,
						entity.TopicPartition.Offset.String(),
						senderID,
						receiverID,
						deliveryTime,
						expirationTime,
						messageType,
						string(entity.Value))
				}
			}
		}

		_ = kafkaConsumer.consumer.Close()
		kafkaConsumer.consumer = nil
	}()

	return kafkaConsumer, nil
}

func (kafkaConsumer *KafkaConsumer) Close() error {
	if nil == kafkaConsumer.consumer {
		return nil
	}

	kafkaConsumer.terminated = true
	return nil
}

func (kafkaConsumer *KafkaConsumer) CreateTopic(topic string, partitionCount int) error {
	if nil == kafkaConsumer.consumer {
		return fmt.Errorf("INVALID INSTANCE")
	}

	adminClient, err := kafka.NewAdminClientFromConsumer(kafkaConsumer.consumer)
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

func (kafkaConsumer *KafkaConsumer) DescribeTopic(topic string) (map[string]kafka.ConfigEntryResult, error) {
	if nil == kafkaConsumer.consumer {
		return nil, fmt.Errorf("INVALID INSTANCE")
	}

	adminClient, err := kafka.NewAdminClientFromConsumer(kafkaConsumer.consumer)
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

func (kafkaConsumer *KafkaConsumer) DeleteTopic(topic string) error {
	if nil == kafkaConsumer.consumer {
		return fmt.Errorf("INVALID INSTANCE")
	}

	adminClient, err := kafka.NewAdminClientFromConsumer(kafkaConsumer.consumer)
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

func (kafkaConsumer *KafkaConsumer) SubscribeTopics(topics []string) error {
	if nil == kafkaConsumer.consumer {
		return fmt.Errorf("INVALID INSTANCE")
	}

	err := kafkaConsumer.consumer.SubscribeTopics(topics, nil)
	if nil != err {
		return fmt.Errorf("FAILED TO SUBSCRIBE TOPIC: %+v", err)
	}

	return nil
}

func (kafkaConsumer *KafkaConsumer) Unsubscribe() error {
	if nil == kafkaConsumer.consumer {
		return fmt.Errorf("INVALID INSTANCE")
	}

	err := kafkaConsumer.consumer.Unsubscribe()
	if nil != err {
		return fmt.Errorf("FAILED TO UNSUBSCRIBE: %+v", err)
	}

	return nil
}

func (kafkaConsumer *KafkaConsumer) GetSubscriptions() ([]string, error) {
	if nil == kafkaConsumer.consumer {
		return nil, fmt.Errorf("INVALID INSTANCE")
	}

	topics, err := kafkaConsumer.consumer.Subscription()
	if nil != err {
		return nil, fmt.Errorf("FAILED TO QUERY SUBSCRIPTIONS: %+v", err)
	}

	return topics, nil
}

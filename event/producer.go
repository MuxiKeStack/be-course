package event

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
)

type Producer interface {
	ProduceCourseListEvent(ctx context.Context, evt CourseFromXkEvent) error
	BatchProduceCourseListEvent(ctx context.Context, evt []CourseFromXkEvent) error
}

type SaramaProducer struct {
	producer sarama.SyncProducer
}

func NewSaramaProducer(client sarama.Client) (Producer, error) {
	p, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return nil, err
	}
	return &SaramaProducer{
		p,
	}, nil
}

func (s *SaramaProducer) ProduceCourseListEvent(ctx context.Context, evt CourseFromXkEvent) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	_, _, err = s.producer.SendMessage(&sarama.ProducerMessage{
		Topic: evt.Topic(),
		Value: sarama.ByteEncoder(data),
	})
	return err
}

func (s *SaramaProducer) BatchProduceCourseListEvent(ctx context.Context, evts []CourseFromXkEvent) error {
	msgs := make([]*sarama.ProducerMessage, 0, len(evts))
	for _, evt := range evts {
		data, err := json.Marshal(evt)
		if err != nil {
			return err
		}
		msgs = append(msgs, &sarama.ProducerMessage{
			Topic: evt.Topic(),
			Value: sarama.ByteEncoder(data),
		})
	}
	return s.producer.SendMessages(msgs)
}

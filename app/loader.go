package app

import (
	"github.com/magento-mcom/inventory-prototype/consumer"
	"github.com/magento-mcom/inventory-prototype/repository"
	"github.com/magento-mcom/inventory-prototype/configuration"
)

func NewLoader(config configuration.Config) Loader {
	return Loader{config, nil, nil}
}

type Loader struct {
	config     configuration.Config
	consumer   consumer.Consumer
	repository repository.Repository
}

func (l *Loader) Consumer() consumer.Consumer {

	if l.consumer == nil {
		l.consumer = consumer.NewSQSConsumer(l.config)
	}

	return l.consumer
}

func (l *Loader) Repository() repository.Repository {

	if l.repository == nil {
		l.repository = repository.NewMysqlRepository(l.config)
	}

	return l.repository
}

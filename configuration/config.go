package configuration

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Database struct {
		DSN 	string
		Engine 	string
	}
	Consumer struct {
		Queuestock 	 	string
		Queuereindex 	string
		Region   		string
		Profile  		string
		Attribute  		string
	}
}

func Load(file string) (cfg Config, err error) {
	data, err := ioutil.ReadFile(file)

	if err != nil {
		return
	}

	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return
	}

	return
}

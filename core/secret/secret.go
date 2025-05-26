package secret

import (
	"fmt"
	"os"

	"github.com/end1essrage/indigo-core/config"
)

//регестрация секретов с помощью конфига, заполнение с помощью env
//проброс секретов в сервисы, но не доступ внутри луа
//возможность из луа приложить секрет и возможность потом в http клиенте например зарезолвить его

type SecretsOperator struct {
	buffer map[string]string
}

func New(sec []config.Secret) *SecretsOperator {
	o := &SecretsOperator{buffer: make(map[string]string, len(sec))}

	for _, n := range sec {
		if os.Getenv(n.Name) == "" {
			panic(fmt.Errorf("env: %s not provided", n))
		}
		o.buffer[n.Name] = os.Getenv(n.Name)
	}

	return o
}

func (s *SecretsOperator) RevealSecret(key string) string {
	return s.buffer[key]
}

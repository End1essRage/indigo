package config

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

func Validate(config *YamlConfig) (bool, string) {
	//параллельно?
	for _, k := range config.Keyboards {
		logrus.Debugf("Validating %s", k.Name)
		if validateKeyboard(k) != nil {
			return false, fmt.Sprintf("ошибка валидации в клавиатуре %s", k.Name)
		}
	}
	return true, ""
}

func validateKeyboard(kb Keyboard) error {
	if kb.Buttons == nil {
		return fmt.Errorf("пустая клавиатура")
	}

	//ограничения тг
	if len(*kb.Buttons) > 10 {
		return fmt.Errorf("слишком много рядов")
	}

	for i, r := range *kb.Buttons {
		if len(r.Row) > 8 {
			return fmt.Errorf("слишком много кнопок в ряду %v", i)
		}
	}

	return nil
}
func validateScripts() {}

func validateMiddleWares() {}

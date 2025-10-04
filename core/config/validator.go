package config

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

func Validate(config *YamlConfig) (bool, string) {
	if err := validateStorage(&config.Storage); err != nil {
		return false, fmt.Sprintf("ошибка валидации Storage %v", err)
	}

	for _, cmd := range config.Commands {
		if validateCommand(&cmd) != nil {
			return false, fmt.Sprintf("ошибка валидации в команде %s", cmd.Name)
		}
	}

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

func validateStorage(config *StorageConfig) error {
	if config.Type == Storage_File {
		if config.File == nil {
			return fmt.Errorf("заполните конфигурацию для файлового хранилища")
		}
	}

	if config.Type == Storage_Mongo {
		if config.Mongo == nil {
			return fmt.Errorf("заполните конфигурацию для монго дб")
		}
	}

	return nil
}

func validateCommand(config *Command) error {
	if config.Use == CmdUse_Group {
		if config.Form != nil {
			return fmt.Errorf("Нельзя передавать формы в группы")
		}
	}

	return nil
}

func validateScripts() {}

func validateMiddleWares() {}

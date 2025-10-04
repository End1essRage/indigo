package config

import (
	"strings"
	"testing"
)

func TestValidateKeyboard(t *testing.T) {
	testCases := []struct {
		name        string
		keyboard    Keyboard
		wantErr     bool
		errContains string
	}{
		{
			name: "valid buttons only",
			keyboard: Keyboard{
				Name: "buttons_kb",
				Buttons: &[]KeyboardRow{
					{Row: make([]Button, 5)},
					{Row: make([]Button, 3)},
				},
			},
			wantErr: false,
		},
		{
			name: "empty keyboard",
			keyboard: Keyboard{
				Name: "empty_kb",
			},
			wantErr:     true,
			errContains: "пустая клавиатура",
		},
		{
			name: "too many rows",
			keyboard: Keyboard{
				Name: "too_many_rows",
				Buttons: &[]KeyboardRow{
					{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {},
				},
			},
			wantErr:     true,
			errContains: "слишком много рядов",
		},
		{
			name: "too many buttons in row",
			keyboard: Keyboard{
				Name: "too_many_buttons",
				Buttons: &[]KeyboardRow{
					{Row: make([]Button, 9)},
				},
			},
			wantErr:     true,
			errContains: "слишком много кнопок в ряду",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateKeyboard(tc.keyboard)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("error %q should contain %q", err.Error(), tc.errContains)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateCommands(t *testing.T) {
	form := ""

	//проверить что в групповых командах нет форм
	t.Run("valid config", func(t *testing.T) {
		cfg := &YamlConfig{
			Commands: []Command{
				{
					Name:        "valid",
					Description: "",
					Use:         CmdUse_Private,
					Form:        &form,
				},
			},
		}

		valid, msg := Validate(cfg)
		if !valid {
			t.Errorf("config should be valid, got error: %s", msg)
		}
	})

	t.Run("invalid config", func(t *testing.T) {
		cfg := &YamlConfig{
			Commands: []Command{
				{
					Name:        "invalid",
					Description: "",
					Use:         CmdUse_Group,
					Form:        &form,
				},
			},
		}

		valid, msg := Validate(cfg)
		if valid {
			t.Errorf("config should be invalid, got error: %s", msg)
		}
	})
}

func TestValidateConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := &YamlConfig{
			Keyboards: []Keyboard{
				{
					Name: "valid2",
					Buttons: &[]KeyboardRow{
						{Row: make([]Button, 5)},
					},
				},
			},
		}

		valid, msg := Validate(cfg)
		if !valid {
			t.Errorf("config should be valid, got error: %s", msg)
		}
	})

	t.Run("invalid config", func(t *testing.T) {
		cfg := &YamlConfig{
			Keyboards: []Keyboard{
				{
					Name: "valid",
					Buttons: &[]KeyboardRow{
						{Row: make([]Button, 5)},
					},
				},
				{
					Name: "invalid",
					Buttons: &[]KeyboardRow{
						{Row: make([]Button, 9)},
					},
				},
			},
		}

		valid, msg := Validate(cfg)
		if valid {
			t.Error("config should be invalid")
		}
		if !strings.Contains(msg, "ошибка валидации в клавиатуре invalid") {
			t.Errorf("unexpected error message: %s", msg)
		}
	})
}

func BenchmarkValidate(b *testing.B) {
	cfg := generateLargeConfig(10000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Validate(cfg)
	}
}

func generateLargeConfig(n int) *YamlConfig {
	kbds := make([]Keyboard, n)
	for i := 0; i < n; i++ {
		if i%10 == 0 {
			// Добавляем ошибки в каждую 10-ю клавиатуру
			kbds[i] = Keyboard{
				Name:    "invalid",
				Buttons: &[]KeyboardRow{{Row: make([]Button, 9)}},
			}
		} else {
			kbds[i] = Keyboard{
				Name:    "valid",
				Buttons: &[]KeyboardRow{{Row: make([]Button, 5)}},
			}
		}
	}
	return &YamlConfig{Keyboards: kbds}
}

func strPtr(s string) *string {
	return &s
}

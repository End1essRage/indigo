# Имя выходного файла (можно изменить)
BINARY_NAME = indigo.exe

.PHONY: build_and_run

build_and_run:
	@echo "Building binary..."
	@cd core && go build -o ../$(BINARY_NAME)
	@echo "Running binary..."
	@./$(BINARY_NAME)

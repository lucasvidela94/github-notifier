.PHONY: build run install enable disable logs test test-install deps clean release

BINARY  := github-notifier
VERSION ?= dev
INSTALL := $(HOME)/.local/bin/$(BINARY)

## Instala dependencias del sistema (Arch Linux)
deps:
	sudo pacman -S --needed gtk3 libappindicator-gtk3 libnotify

## Descarga módulos Go
mod:
	go mod tidy

## Compila el binario
build:
	go build -ldflags="-s -w -X main.Version=$(VERSION)" -o $(BINARY) .

## Compila y ejecuta (requiere GITHUB_TOKEN)
run: build
	./$(BINARY)

## Instala via install.sh
install:
	bash install.sh

## Corre todos los tests (Go + installer)
test:
	go test ./internal/... -v -count=1

## Corre tests del installer
test-install: build
	bash install_test.sh

## Habilita e inicia el servicio systemd
enable:
	systemctl --user daemon-reload
	systemctl --user enable --now github-notifier.service
	@echo "Servicio activo. Usa 'make logs' para ver salida."

## Detiene y deshabilita el servicio
disable:
	systemctl --user disable --now github-notifier.service

## Ver logs del servicio en tiempo real
logs:
	journalctl --user -u github-notifier.service -f

## Crea un tag y pushea para triggear release
release:
	@test -n "$(V)" || (echo "Usa: make release V=v1.0.0" && exit 1)
	git tag -a $(V) -m "Release $(V)"
	git push origin $(V)
	@echo "Tag $(V) pushed. GitHub Actions will build and release."

## Elimina binario, servicio y autoarranque
clean:
	rm -f $(BINARY) $(INSTALL)
	-systemctl --user disable --now github-notifier.service 2>/dev/null
	rm -f $(HOME)/.config/systemd/user/github-notifier.service

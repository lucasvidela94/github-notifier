.PHONY: build run install enable disable logs test deps clean

BINARY  := github-notifier
INSTALL := $(HOME)/.local/bin/$(BINARY)
AUTOSTART_DIR := $(HOME)/.config/autostart
AUTOSTART_FILE := $(AUTOSTART_DIR)/$(BINARY).desktop

## Instala dependencias del sistema (Arch Linux)
deps:
	sudo pacman -S --needed gtk3 libappindicator-gtk3 libnotify

## Descarga módulos Go
mod:
	go mod tidy

## Compila el binario
build:
	go build -o $(BINARY) .

## Compila y ejecuta (requiere GITHUB_TOKEN)
run: build
	./$(BINARY)

## Instala el binario en ~/.local/bin y configura autoarranque
install: build
	@mkdir -p $(dir $(INSTALL)) $(AUTOSTART_DIR)
	cp $(BINARY) $(INSTALL)
	@echo "[Desktop Entry]"                           >  $(AUTOSTART_FILE)
	@echo "Type=Application"                         >> $(AUTOSTART_FILE)
	@echo "Name=GitHub Notifier"                     >> $(AUTOSTART_FILE)
	@echo "Exec=env GITHUB_TOKEN=$$GITHUB_TOKEN $(INSTALL)" >> $(AUTOSTART_FILE)
	@echo "Hidden=false"                             >> $(AUTOSTART_FILE)
	@echo "NoDisplay=false"                          >> $(AUTOSTART_FILE)
	@echo "X-GNOME-Autostart-enabled=true"           >> $(AUTOSTART_FILE)
	@echo "Comment=Notificaciones de GitHub en la barra del sistema" >> $(AUTOSTART_FILE)
	@echo ""
	@echo "✓ Instalado en $(INSTALL)"
	@echo "✓ Autoarranque en $(AUTOSTART_FILE)"
	@echo ""
	@echo "Asegurate de exportar GITHUB_TOKEN en tu shell antes de iniciar sesión."

## Corre todos los tests
test:
	go test ./internal/... -v

## Habilita e inicia el servicio systemd (arranca solo con la sesión)
enable: install
	systemctl --user daemon-reload
	systemctl --user enable --now github-notifier.service
	@echo "Servicio activo. Usa 'make logs' para ver salida."

## Detiene y deshabilita el servicio
disable:
	systemctl --user disable --now github-notifier.service

## Ver logs del servicio en tiempo real
logs:
	journalctl --user -u github-notifier.service -f

## Elimina binario, servicio y autoarranque
clean:
	rm -f $(BINARY) $(INSTALL) $(AUTOSTART_FILE)
	-systemctl --user disable --now github-notifier.service 2>/dev/null
	rm -f $(HOME)/.config/systemd/user/github-notifier.service

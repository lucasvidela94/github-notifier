# Support

## Getting Help

- **Issues**: Open a [GitHub Issue](https://github.com/your-username/github-notifier/issues)
  for bugs and feature requests
- **Discussions**: Use [GitHub Discussions](https://github.com/your-username/github-notifier/discussions)
  for questions

## FAQ

### The tray icon doesn't appear

Make sure you have `libayatana-appindicator3` installed:
- Arch: `sudo pacman -S libayatana-appindicator3`
- Ubuntu/Debian: `sudo apt install libayatana-appindicator3`
- Fedora: `sudo dnf install libayatana-appindicator3`

### Notifications aren't showing

Check that `notify-send` is available:
```bash
which notify-send
notify-send "Test" "This is a test"
```

### Service won't start

Check the logs:
```bash
make logs
```

Make sure your GitHub token has the correct permissions (`notifications` and `repo` scopes).

### How do I update the app?

Pull the latest changes and rebuild:
```bash
git pull
make disable
make enable
```

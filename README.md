# Jellyfin TUI

A terminal user interface (TUI) for browsing and playing media from your Jellyfin server.

![Jellyfin TUI Screenshot](screenshot.png)

## Features

- Browse movies and TV shows from your Jellyfin server
- Navigate through TV show seasons and episodes
- Search for content across your media library
- Play media using MPV player
- Simple configuration management
- Keyboard-driven interface

## Installation

### Prerequisites

- Go 1.16 or higher
- MPV media player

### Building from source

```bash
git clone https://github.com/fabean/jellyfin-tui.git
cd jellyfin-tui
go build -o jellyfin-tui ./cmd/jellyfin-tui
```

For development, you can also install it directly:

```bash
go install ./cmd/jellyfin-tui
```
## Usage

### First Run

On first run, the application will create a default configuration file at `~/.config/jellyfin-tui/config`. You'll need to update this with your Jellyfin server details through the configuration menu.

### Navigation

- **Arrow keys**: Navigate through lists
- **Enter**: Select an item
- **Escape**: Go back to the previous screen
- **q or Ctrl+C**: Quit the application

### Main Menu

- **Movies**: Browse your movie library
- **TV Shows**: Browse your TV show library
- **Search**: Search for content
- **Configure**: Update your Jellyfin server settings

### Configuration

You can configure your Jellyfin server connection through the application:

1. Select "Configure" from the main menu
2. Enter your Jellyfin server URL (e.g., `https://jellyfin.example.com`)
3. Enter your Jellyfin API key
4. Press Enter to save

The configuration is stored in `~/.config/jellyfin-tui/config`.

### Playing Media

When you select a movie or TV episode, it will automatically play using MPV. Make sure MPV is installed and available in your PATH.

## Getting a Jellyfin API Key

1. Log in to your Jellyfin server web interface
2. Go to Dashboard → API Keys
3. Create a new API key
4. Copy the key and use it in the Jellyfin TUI configuration

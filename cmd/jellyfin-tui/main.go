package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fabean/jellyfin-tui/jellyfin"
)

// Config holds the Jellyfin server configuration
type Config struct {
	ServerURL string `json:"server_url"`
	APIKey    string `json:"api_key"`
}

// MediaItem represents a movie or TV show
type MediaItem struct {
	ID           string
	ItemTitle    string
	Type         string
	ImageURL     string
	StreamURL    string
	ParentID     string
	IndexNumber  int    // Add this field for episode numbers
	DisplayTitle string // Add this for formatted display title
}

// Implement the list.Item interface for MediaItem
func (m MediaItem) Title() string {
	if m.DisplayTitle != "" {
		return m.DisplayTitle
	}
	return m.ItemTitle
}

func (m MediaItem) Description() string { return m.Type }
func (m MediaItem) FilterValue() string { return m.ItemTitle }

// Model represents the application state
type Model struct {
	config       Config
	currentView  string // "main", "movies", "tvshows", "seasons", "episodes", "search", "config"
	mainList     list.Model
	moviesList   list.Model
	tvShowsList  list.Model
	seasonsList  list.Model
	episodesList list.Model
	searchInput  textinput.Model
	searchList   list.Model
	configInputs []textinput.Model // Add this for config inputs
	currentItem  MediaItem
	err          error
}

// Initialize the application
func initialModel() Model {
	// Load or create config
	config, err := loadConfig()
	if err != nil {
		// If there's an error loading the config, create a default one
		config = Config{
			ServerURL: "https://jellyfin.example.com",
			APIKey:    "your_api_key_here",
		}
		// Save the default config
		saveConfig(config)
	}

	// Set up the main menu
	mainItems := []list.Item{
		MediaItem{ItemTitle: "Movies", Type: "category"},
		MediaItem{ItemTitle: "TV Shows", Type: "category"},
		MediaItem{ItemTitle: "Search", Type: "action"},
		MediaItem{ItemTitle: "Configure", Type: "action"},
	}

	mainList := list.New(mainItems, list.NewDefaultDelegate(), 0, 0)
	mainList.Title = "Jellyfin TUI"

	// Set up empty lists for movies and TV shows
	moviesList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	moviesList.Title = "Movies"

	tvShowsList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	tvShowsList.Title = "TV Shows"

	// Set up empty lists for seasons and episodes
	seasonsList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	seasonsList.Title = "Seasons"

	episodesList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	episodesList.Title = "Episodes"

	// Set up search input
	searchInput := textinput.New()
	searchInput.Placeholder = "Search for movies and TV shows..."
	searchInput.Focus()

	// Set up empty search results list
	searchList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	searchList.Title = "Search Results"

	// Set up config inputs
	serverInput := textinput.New()
	serverInput.Placeholder = "Jellyfin Server URL"
	serverInput.Focus()
	serverInput.Width = 40
	serverInput.SetValue(config.ServerURL)

	apiKeyInput := textinput.New()
	apiKeyInput.Placeholder = "Jellyfin API Key"
	apiKeyInput.Width = 40
	apiKeyInput.SetValue(config.APIKey)

	configInputs := []textinput.Model{serverInput, apiKeyInput}

	return Model{
		config:       config,
		currentView:  "main",
		mainList:     mainList,
		moviesList:   moviesList,
		tvShowsList:  tvShowsList,
		seasonsList:  seasonsList,
		episodesList: episodesList,
		searchInput:  searchInput,
		searchList:   searchList,
		configInputs: configInputs,
	}
}

// loadConfig loads the configuration from ~/.config/jellyfin-tui/config
func loadConfig() (Config, error) {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Config{}, fmt.Errorf("failed to get home directory: %v", err)
	}

	// Create config directory path
	configDir := filepath.Join(homeDir, ".config", "jellyfin-tui")
	configFile := filepath.Join(configDir, "config")

	// Check if config file exists
	_, err = os.Stat(configFile)
	if os.IsNotExist(err) {
		return Config{}, fmt.Errorf("config file does not exist")
	}

	// Read config file
	data, err := os.ReadFile(configFile)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %v", err)
	}

	// Parse config
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse config file: %v", err)
	}

	return config, nil
}

// saveConfig saves the configuration to ~/.config/jellyfin-tui/config
func saveConfig(config Config) error {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}

	// Create config directory path
	configDir := filepath.Join(homeDir, ".config", "jellyfin-tui")
	
	// Create directory if it doesn't exist
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	// Create config file path
	configFile := filepath.Join(configDir, "config")

	// Marshal config to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	// Write config file
	err = os.WriteFile(configFile, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// Define message types
type fetchMoviesMsg []MediaItem
type fetchTVShowsMsg []MediaItem
type fetchSeasonsMsg []MediaItem
type fetchEpisodesMsg []MediaItem
type searchResultsMsg []MediaItem
type errorMsg error

// Update function handles all the application logic
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			// Handle navigation back up the hierarchy
			switch m.currentView {
			case "episodes":
				m.currentView = "seasons"
				return m, nil
			case "seasons":
				m.currentView = "tvshows"
				return m, nil
			case "movies", "tvshows", "search":
				m.currentView = "main"
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		h, v := lipgloss.NewStyle().Margin(1, 2).GetFrameSize()
		m.mainList.SetSize(msg.Width-h, msg.Height-v)
		m.moviesList.SetSize(msg.Width-h, msg.Height-v)
		m.tvShowsList.SetSize(msg.Width-h, msg.Height-v)
		m.seasonsList.SetSize(msg.Width-h, msg.Height-v)
		m.episodesList.SetSize(msg.Width-h, msg.Height-v)
		m.searchList.SetSize(msg.Width-h, msg.Height-v)

	case fetchMoviesMsg:
		m.moviesList.SetItems(convertToListItems(msg))
		return m, nil

	case fetchTVShowsMsg:
		m.tvShowsList.SetItems(convertToListItems(msg))
		return m, nil

	case fetchSeasonsMsg:
		m.seasonsList.SetItems(convertToListItems(msg))
		return m, nil

	case fetchEpisodesMsg:
		m.episodesList.SetItems(convertToListItems(msg))
		return m, nil

	case searchResultsMsg:
		m.searchList.SetItems(convertToListItems(msg))
		return m, nil

	case errorMsg:
		m.err = msg
		return m, nil
	}

	// Handle different views
	switch m.currentView {
	case "main":
		m.mainList, cmd = m.mainList.Update(msg)
		
		// Handle selection in main menu
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" {
			selectedItem, ok := m.mainList.SelectedItem().(MediaItem)
			if ok {
				switch selectedItem.ItemTitle {
				case "Movies":
					m.currentView = "movies"
					return m, fetchMovies(m.config)
				case "TV Shows":
					m.currentView = "tvshows"
					return m, fetchTVShows(m.config)
				case "Search":
					m.currentView = "search"
					m.searchInput.SetValue("")
					return m, nil
				case "Configure":
					m.currentView = "config"
					m.configInputs[0].SetValue(m.config.ServerURL)
					m.configInputs[1].SetValue(m.config.APIKey)
					m.configInputs[0].Focus()
					return m, nil
				}
			}
		}

	case "movies", "tvshows":
		var list *list.Model
		if m.currentView == "movies" {
			list = &m.moviesList
		} else {
			list = &m.tvShowsList
		}
		
		*list, cmd = list.Update(msg)
		
		// Handle selection of a movie or TV show
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" {
			selectedItem, ok := list.SelectedItem().(MediaItem)
			if ok && selectedItem.ID != "" {
				m.currentItem = selectedItem
				m.currentView = "seasons"
				return m, fetchSeasons(m.config, selectedItem.ID)
			}
		}

	case "seasons":
		m.seasonsList, cmd = m.seasonsList.Update(msg)
		
		// Handle selection of a season
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" {
			selectedItem, ok := m.seasonsList.SelectedItem().(MediaItem)
			if ok && selectedItem.ID != "" {
				m.currentItem = selectedItem
				m.currentView = "episodes"
				return m, fetchEpisodes(m.config, selectedItem.ID)
			}
		}

	case "episodes":
		m.episodesList, cmd = m.episodesList.Update(msg)
		
		// Handle selection of an episode
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" {
			selectedItem, ok := m.episodesList.SelectedItem().(MediaItem)
			if ok && selectedItem.ID != "" {
				return m, playMedia(selectedItem)
			}
		}

	case "search":
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" {
			query := m.searchInput.Value()
			if query != "" {
				return m, searchMedia(m.config, query)
			}
		} else {
			m.searchInput, cmd = m.searchInput.Update(msg)
		}

		// If we're viewing search results
		if len(m.searchList.Items()) > 0 {
			m.searchList, cmd = m.searchList.Update(msg)
			
			// Handle selection of a search result
			if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" {
				selectedItem, ok := m.searchList.SelectedItem().(MediaItem)
				if ok && selectedItem.ID != "" {
					return m, playMedia(selectedItem)
				}
			}
		}

	case "config":
		// Handle tab to switch between inputs
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "tab", "down":
				// Move focus to next input
				if m.configInputs[0].Focused() {
					m.configInputs[0].Blur()
					m.configInputs[1].Focus()
				} else {
					m.configInputs[0].Focus()
					m.configInputs[1].Blur()
				}
				return m, nil
				
			case "shift+tab", "up":
				// Move focus to previous input
				if m.configInputs[0].Focused() {
					m.configInputs[0].Blur()
					m.configInputs[1].Focus()
				} else {
					m.configInputs[0].Focus()
					m.configInputs[1].Blur()
				}
				return m, nil
				
			case "enter":
				// Save config
				newConfig := Config{
					ServerURL: m.configInputs[0].Value(),
					APIKey:    m.configInputs[1].Value(),
				}
				
				err := saveConfig(newConfig)
				if err != nil {
					m.err = err
					return m, nil
				}
				
				m.config = newConfig
				m.currentView = "main"
				return m, nil
			}
		}

		// Update the focused input
		for i := range m.configInputs {
			if m.configInputs[i].Focused() {
				m.configInputs[i], cmd = m.configInputs[i].Update(msg)
				break
			}
		}
	}

	return m, cmd
}

// View renders the current UI
func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress any key to exit.", m.err)
	}

	switch m.currentView {
	case "main":
		return m.mainList.View()
	case "movies":
		return m.moviesList.View()
	case "tvshows":
		return m.tvShowsList.View()
	case "seasons":
		return m.seasonsList.View()
	case "episodes":
		return m.episodesList.View()
	case "search":
		if len(m.searchList.Items()) > 0 {
			return m.searchList.View()
		}
		return fmt.Sprintf(
			"Search: %s\n\nType a search query and press Enter",
			m.searchInput.View(),
		)
	case "config":
		return fmt.Sprintf(
			"Configure Jellyfin Connection\n\n"+
				"Server URL: %s\n\n"+
				"API Key: %s\n\n"+
				"(Press Enter to save and return to main menu)",
			m.configInputs[0].View(),
			m.configInputs[1].View(),
		)
	default:
		return "Unknown view"
	}
}

// Helper function to convert MediaItems to list.Items
func convertToListItems(items []MediaItem) []list.Item {
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}
	return listItems
}

// Command to fetch movies from Jellyfin
func fetchMovies(config Config) tea.Cmd {
	return func() tea.Msg {
		client := jellyfin.NewClient(config.ServerURL, config.APIKey)
		items, err := client.GetMovies()
		if err != nil {
			return errorMsg(err)
		}
		
		// Convert jellyfin.MediaItem to our MediaItem
		mediaItems := make([]MediaItem, len(items))
		for i, item := range items {
			mediaItems[i] = MediaItem{
				ID:        item.ID,
				ItemTitle: item.Name,
				Type:      item.MediaType,
				// You can construct image URL if needed
				StreamURL: client.GetStreamURL(item.ID),
			}
		}
		
		return fetchMoviesMsg(mediaItems)
	}
}

// Command to fetch TV shows from Jellyfin
func fetchTVShows(config Config) tea.Cmd {
	return func() tea.Msg {
		client := jellyfin.NewClient(config.ServerURL, config.APIKey)
		items, err := client.GetTVShows()
		if err != nil {
			return errorMsg(err)
		}
		
		// Convert jellyfin.MediaItem to our MediaItem
		mediaItems := make([]MediaItem, len(items))
		for i, item := range items {
			mediaItems[i] = MediaItem{
				ID:        item.ID,
				ItemTitle: item.Name,
				Type:      "tvshow",
				// You can construct image URL if needed
				StreamURL: client.GetStreamURL(item.ID),
			}
		}
		
		return fetchTVShowsMsg(mediaItems)
	}
}

// Command to fetch seasons for a TV show
func fetchSeasons(config Config, seriesID string) tea.Cmd {
	return func() tea.Msg {
		client := jellyfin.NewClient(config.ServerURL, config.APIKey)
		endpoint := fmt.Sprintf("%s/Shows/%s/Seasons?api_key=%s", 
			config.ServerURL, seriesID, config.APIKey)
		
		items, err := client.FetchItems(endpoint)
		if err != nil {
			return errorMsg(err)
		}
		
		// Convert jellyfin.MediaItem to our MediaItem
		mediaItems := make([]MediaItem, len(items))
		for i, item := range items {
			mediaItems[i] = MediaItem{
				ID:        item.ID,
				ItemTitle: item.Name,
				Type:      "season",
				ParentID:  seriesID,
				StreamURL: "",
			}
		}
		
		return fetchSeasonsMsg(mediaItems)
	}
}

// Command to fetch episodes for a season
func fetchEpisodes(config Config, seasonID string) tea.Cmd {
	return func() tea.Msg {
		client := jellyfin.NewClient(config.ServerURL, config.APIKey)
		
		// Fix the endpoint URL format - this is the correct Jellyfin API path
		endpoint := fmt.Sprintf("%s/Items?ParentId=%s&api_key=%s&SortBy=SortName", 
			config.ServerURL, seasonID, config.APIKey)
		
		items, err := client.FetchItems(endpoint)
		if err != nil {
			return errorMsg(err)
		}
		
		// Convert jellyfin.MediaItem to our MediaItem
		mediaItems := make([]MediaItem, len(items))
		for i, item := range items {
			// Format the display title to include episode number
			displayTitle := item.Name
			if item.IndexNumber > 0 {
				displayTitle = fmt.Sprintf("E%02d: %s", item.IndexNumber, item.Name)
			}
			
			mediaItems[i] = MediaItem{
				ID:           item.ID,
				ItemTitle:    item.Name,
				Type:         "episode",
				ParentID:     seasonID,
				StreamURL:    client.GetStreamURL(item.ID),
				IndexNumber:  item.IndexNumber,
				DisplayTitle: displayTitle,
			}
		}
		
		// Sort episodes by index number
		sort.Slice(mediaItems, func(i, j int) bool {
			return mediaItems[i].IndexNumber < mediaItems[j].IndexNumber
		})
		
		return fetchEpisodesMsg(mediaItems)
	}
}

// Command to search for media
func searchMedia(config Config, query string) tea.Cmd {
	return func() tea.Msg {
		client := jellyfin.NewClient(config.ServerURL, config.APIKey)
		items, err := client.Search(query)
		if err != nil {
			return errorMsg(err)
		}
		
		// Convert jellyfin.MediaItem to our MediaItem
		mediaItems := make([]MediaItem, len(items))
		for i, item := range items {
			mediaItems[i] = MediaItem{
				ID:        item.ID,
				ItemTitle: item.Name,
				Type:      item.MediaType,
				// You can construct image URL if needed
				StreamURL: client.GetStreamURL(item.ID),
			}
		}
		
		return searchResultsMsg(mediaItems)
	}
}

// Command to play media with MPV
func playMedia(item MediaItem) tea.Cmd {
	return func() tea.Msg {
		fmt.Printf("Playing %s (%s) with MPV\n", item.ItemTitle, item.ID)
		
		// Actually play the media with MPV
		cmd := exec.Command("mpv", item.StreamURL)
		err := cmd.Start()
		if err != nil {
			return errorMsg(fmt.Errorf("failed to start MPV: %v", err))
		}
		
		return nil
	}
}

// Add the Init method to implement the tea.Model interface
func (m Model) Init() tea.Cmd {
	return nil
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
} 
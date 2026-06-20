package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	pokecache "github.com/rickypurnell/pokedex-cli/internal"
)

var (
	LocURL   = "https://pokeapi.co/api/v2/location-area/?offset=0limit=20"
	PokeURL  = "https://pokeapi.co/api/v2/location-area/"
	CatchURL = "https://pokeapi.co/api/v2/pokemon/"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*config, string) error
}

type locationData struct {
	Count    int              `json:"count"`
	Next     string           `json:"next"`
	Previous string           `json:"previous"`
	Results  []map[string]any `json:"results"`
}

type config struct {
	LocURL   string
	Next     string
	Previous string
	PCache   *pokecache.Cache
	Caught   map[string]pokemonData
}

type explorePoke struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

type pokemonData struct {
	// name, height, weight, stats and type(s)
	Name    string `json:"name"`
	Height  int    `json:"height"`
	Weight  int    `json:"weight"`
	BaseExp int    `json:"base_experience"`
	Stats   []struct {
		BaseStat int `json:"base_stat"`
		Stat     struct {
			Name string `json:"name"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
}

func getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Provides features available",
			callback:    commandHelp,
		},
		"map": {
			name:        "map",
			description: "Displays 20 locations and each subsequent call displays the next 20",
			callback:    commandMapup,
		},
		"mapb": {
			name:        "mapb",
			description: "Displays the previous 20 locations",
			callback:    commandMapback,
		},
		"explore": {
			name:        "explore",
			description: "Lists all the Pokemon located in an area",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Attempt to capture a specified Pokemon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Provides the stats to a caught Pokemon",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "Lists all Pokemon captured",
			callback:    commandPokedex,
		},
	}
}

// Checks cache to retrieve data before sending GET request.
// Data from GET request is saved in cache for future use.
func getLocations(cfg *config, url string) (locationData, error) {
	bytes, ok := cfg.PCache.Get(url)
	if ok {
		var loc locationData
		err := json.Unmarshal(bytes, &loc)
		if err != nil {
			return locationData{}, err
		}
		return loc, nil
	}
	response, err := http.Get(url)
	if err != nil {
		return locationData{}, err
	}
	defer response.Body.Close()

	var loc locationData
	err = json.NewDecoder(response.Body).Decode(&loc)
	if err != nil {
		return locationData{}, err
	}
	data, err := json.Marshal(loc)
	if err != nil {
		return locationData{}, nil
	}
	cfg.PCache.Add(url, data)
	return loc, nil
}

// Prints the locations returned from getLocations() using
// the provided URL from commandMapup/back
func commandMap(cfg *config, url string) error {
	locations, err := getLocations(cfg, url)
	if err != nil {
		fmt.Println("error retrieving json")
		return err
	}
	for _, result := range locations.Results {
		fmt.Println(result["name"])
	}
	cfg.Next = locations.Next
	cfg.Previous = locations.Previous
	return nil
}

// Paginates forward through the Pokemon world locations
func commandMapup(cfg *config, _ string) error {
	err := commandMap(cfg, cfg.Next)
	if err != nil {
		fmt.Println("error commandMapup")
		return err
	}
	return nil
}

// Paginates backward through the Pokemon world locations
func commandMapback(cfg *config, _ string) error {
	if cfg.Previous == "" {
		err := commandMap(cfg, cfg.LocURL)
		if err != nil {
			return err
		}
		return nil
	}
	err := commandMap(cfg, cfg.Previous)
	if err != nil {
		return err
	}
	return nil
}

// Displays all of the user available functions
func commandHelp(cfg *config, _ string) error {
	fmt.Print("Usage:\n\n")
	cmd := getCommands()
	for k := range cmd {
		fmt.Printf("%s: %s\n", cmd[k].name, cmd[k].description)
	}
	return nil
}

// Closes the Pokedex
func commandExit(cfg *config, _ string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandExplore(cfg *config, area string) error {
	exploreURL := PokeURL + area

	if bytes, ok := cfg.PCache.Get(exploreURL); ok {
		var explorepoke explorePoke
		err := json.Unmarshal(bytes, &explorepoke)
		if err != nil {
			return err
		}
		fmt.Printf("Exploring %s...\n", area)
		fmt.Println("Found Pokemon:")
		for _, encounter := range explorepoke.PokemonEncounters {
			fmt.Printf("- %s\n", encounter.Pokemon.Name)
		}
		return nil

	}

	resp, err := http.Get(exploreURL)
	if err != nil {
		return err
	}
	if resp.StatusCode == 404 {
		return http.ErrMissingFile
	}
	defer resp.Body.Close()

	var explorepoke explorePoke
	err = json.NewDecoder(resp.Body).Decode(&explorepoke)
	if err != nil {
		return err
	}
	fmt.Printf("Exploring %s...\n", area)
	fmt.Println("Found Pokemon:")
	for _, encounter := range explorepoke.PokemonEncounters {
		fmt.Printf("- %s\n", encounter.Pokemon.Name)
	}
	data, err := json.Marshal(explorepoke)
	if err != nil {
		return err
	}
	cfg.PCache.Add(exploreURL, data)
	return nil
}

func commandCatch(cfg *config, pokemon string) error {
	fmt.Printf("Throwing a Pokeball at %s...\n", pokemon)
	pokemonURL := CatchURL + pokemon
	if data, ok := cfg.PCache.Get(pokemonURL); ok {
		var pokeData pokemonData
		err := json.Unmarshal(data, &pokeData)
		if err != nil {
			return err
		}
		displayPokemon(pokeData)
		if ok := catchChance(cfg, pokeData.BaseExp, pokemon); ok {
			cfg.Caught[pokemon] = pokeData
		}
		return nil
	}
	resp, err := http.Get(pokemonURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var pokeData pokemonData
	err = json.NewDecoder(resp.Body).Decode(&pokeData)
	if err != nil {
		return err
	}
	displayPokemon(pokeData)
	if ok := catchChance(cfg, pokeData.BaseExp, pokemon); ok {
		cfg.Caught[pokemon] = pokeData
	}

	data, err := json.Marshal(pokeData)
	if err != nil {
		return err
	}
	cfg.PCache.Add(pokemonURL, data)
	return nil
}

func displayPokemon(pokemon pokemonData) {
	fmt.Printf("Name: %s\n", pokemon.Name)
	fmt.Printf("Height: %d\n", pokemon.Height)
	fmt.Printf("Weight: %d\n", pokemon.Weight)
	fmt.Println("Stats:")
	for _, stat := range pokemon.Stats {
		fmt.Printf("  - %s: %d\n", stat.Stat.Name, stat.BaseStat)
	}
	fmt.Println("Types:")
	for _, t := range pokemon.Types {
		fmt.Printf("  - %s\n", t.Type.Name)
	}
}

func catchChance(cfg *config, exp int, pokemon string) bool {
	randVal := rand.Intn(800)
	if randVal > exp {
		fmt.Printf("%s was caught!\n", pokemon)
		fmt.Println("You may now inspect it with the inspect command.")
		return true
	}
	fmt.Printf("%s escaped!\n", pokemon)
	return false
}

func commandInspect(cfg *config, name string) error {
	if pokemon, ok := cfg.Caught[name]; ok {
		displayPokemon(pokemon)
		return nil
	}
	fmt.Println("This pokemon hasn't been caught yet.")
	return nil
}

func commandPokedex(cfg *config, _ string) error {
	fmt.Println("Your Pokedex:")
	for name := range cfg.Caught {
		fmt.Printf("  - %s\n", name)
	}
	return nil
}

func cleanInput(text string) ([]string, int) {
	trimlower := strings.TrimSpace(strings.ToLower(text))
	stringSlice := strings.Split(trimlower, " ")
	sliceLength := len(stringSlice)

	return stringSlice, sliceLength
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	commands := getCommands()
	config := config{
		LocURL:   LocURL,
		Next:     LocURL,
		Previous: "",
		PCache:   pokecache.NewCache(30 * time.Second),
		Caught:   make(map[string]pokemonData),
	}

	fmt.Println("Welcome to the Pokedex!")
	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		textInput := scanner.Text()
		inputSlice, _ := cleanInput(textInput)
		cmd, ok := commands[inputSlice[0]]
		if !ok {
			fmt.Println("Unknown command")
			continue
		}
		if len(inputSlice) == 1 {
			err := cmd.callback(&config, "")
			if err != nil {
				fmt.Printf("Error: %v", err)
				continue
			}
			continue
		}
		err := cmd.callback(&config, inputSlice[1])
		if err != nil {
			fmt.Printf("Error: %v", err)
			continue
		}
	}
}

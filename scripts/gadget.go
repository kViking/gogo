package scripts

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
)

// Gadget represents a single user-defined command shortcut
// Name is not stored in the struct, but as the key in the map
// (for compatibility with existing user_scripts.json)
type Gadget struct {
	Description string            `json:"description"`
	Command     string            `json:"command"`
	Variables   map[string]string `json:"variables"`
}

type GadgetStore struct {
	gadgets map[string]Gadget
	path    string
}

// NewGadgetStore loads gadgets from the default config path
func NewGadgetStore() (*GadgetStore, error) {
	path := getUserScriptsPath()
	gadgets := make(map[string]Gadget)
	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		_ = json.Unmarshal(data, &gadgets)
	}
	return &GadgetStore{gadgets: gadgets, path: path}, nil
}

// Save writes the current gadgets to disk
func (s *GadgetStore) Save() error {
	data, err := json.MarshalIndent(s.gadgets, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

// Add adds or updates a gadget
func (s *GadgetStore) Add(name string, g Gadget) {
	s.gadgets[name] = g
}

// Delete removes a gadget by name
func (s *GadgetStore) Delete(name string) error {
	if _, ok := s.gadgets[name]; !ok {
		return errors.New("gadget not found")
	}
	delete(s.gadgets, name)
	return nil
}

// Get returns a gadget by name
func (s *GadgetStore) Get(name string) (Gadget, bool) {
	g, ok := s.gadgets[name]
	return g, ok
}

// List returns all gadgets
func (s *GadgetStore) List() map[string]Gadget {
	return s.gadgets
}

// ExtractVariables returns a list of variable names in the command string
func ExtractVariables(command string) []string {
	var vars []string
	seen := map[string]bool{}
	re := regexp.MustCompile(`\{\{([A-Za-z0-9_]+)\}\}`)
	matches := re.FindAllStringSubmatch(command, -1)
	for _, m := range matches {
		if !seen[m[1]] {
			vars = append(vars, m[1])
			seen[m[1]] = true
		}
	}
	return vars
}

// Validation and business logic for gadgets
func ValidateGadgetName(name string) error {
	nameRe := regexp.MustCompile(`^[A-Za-z0-9_\-]+$`)
	if !nameRe.MatchString(name) {
		return errors.New("Gadget names cannot contain spaces or punctuation. Use only letters, numbers, dashes, or underscores.")
	}
	return nil
}

func ValidateGadgetCommand(command string) error {
	if command == "" {
		return errors.New("Command cannot be empty.")
	}
	return nil
}

func CreateGadget(store *GadgetStore, name, command, desc string, variables map[string]string) error {
	if err := ValidateGadgetName(name); err != nil {
		return err
	}
	if err := ValidateGadgetCommand(command); err != nil {
		return err
	}
	store.Add(name, Gadget{
		Description: desc,
		Command:     command,
		Variables:   variables,
	})
	return store.Save()
}

func EditGadget(store *GadgetStore, name string, updates map[string]interface{}) error {
	g, ok := store.Get(name)
	if !ok {
		return errors.New("Gadget not found")
	}
	if v, ok := updates["name"]; ok {
		if newName, ok := v.(string); ok && newName != "" && newName != name {
			if err := ValidateGadgetName(newName); err != nil {
				return err
			}
			store.Add(newName, g)
			_ = store.Delete(name)
			name = newName
		}
	}
	if v, ok := updates["description"]; ok {
		if desc, ok := v.(string); ok {
			g.Description = desc
		}
	}
	if v, ok := updates["command"]; ok {
		if cmd, ok := v.(string); ok {
			if err := ValidateGadgetCommand(cmd); err != nil {
				return err
			}
			g.Command = cmd
		}
	}
	if v, ok := updates["variables"]; ok {
		if vars, ok := v.(map[string]string); ok {
			g.Variables = vars
		}
	}
	store.Add(name, g)
	return store.Save()
}

// getUserScriptsPath returns the user-writable path for user_scripts.json
func getUserScriptsPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir, _ = os.UserHomeDir()
	}
	dir = filepath.Join(dir, "GoGoGadget")
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "user_scripts.json")
}

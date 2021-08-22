package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/fullstaq-labs/sqedule/lib"
)

type State struct {
	AuthToken               string    `json:"auth_token"`
	AuthTokenExpirationTime time.Time `json:"auth_token_expiration_time"`
}

var MockState *State

func LoadState(reader io.Reader) (State, error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return State{}, err
	}

	var result State
	err = json.Unmarshal(data, &result)
	if err != nil {
		return State{}, err
	}

	return result, nil
}

// LoadStateFromFilesystem attempts to load state from the state file.
// If the state file does not exists, it returns the empty state.
func LoadStateFromFilesystem() (State, error) {
	if MockState != nil {
		return *MockState, nil
	}

	stateFilePath, err := StateFilePath()
	if err != nil {
		return State{}, fmt.Errorf("Error determining state file path: %w", err)
	}

	f, err := os.Open(stateFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return State{}, nil
		} else {
			return State{}, fmt.Errorf("Error opening %s for reading: %w", stateFilePath, err)
		}
	}

	defer f.Close()
	return LoadState(f)
}

func StateFilePath() (string, error) {
	dir, err := lib.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("Error determining user config directory: %w", err)
	}

	return path.Join(dir, "sqedule-cli", "state.json"), nil
}

func (state State) RequireAuthToken() error {
	if !SupportLogin {
		return nil
	}
	if len(state.AuthToken) == 0 {
		return errors.New("Login required. Please run 'sqedule-cli login'")
	} else if time.Now().After(state.AuthTokenExpirationTime) {
		return errors.New("Authentication token expired. Please re-run 'sqedule-cli login'")
	} else {
		return nil
	}
}

func (state State) SaveToFilesystem() error {
	if MockState != nil {
		*MockState = state
		return nil
	}

	stateFilePath, err := StateFilePath()
	if err != nil {
		return fmt.Errorf("Error determining state file path: %w", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("Error marshalling state data: %w", err)
	}

	dir := path.Dir(stateFilePath)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("Error creating state file directory %s:% w", dir, err)
	}

	return os.WriteFile(stateFilePath, data, 0600)
}

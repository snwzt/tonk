package store

import (
	"fmt"
	"sync"
	"tonk/models"
)

type Store interface {
	SetReceivedTanks(string, string, models.Tank) error
	SetReceivedShots(string, string, []models.Shot) error
	SetReceivedMines(string, string, []models.Mine) error
	GetState(string) (*models.State, error) // in the future, get limited state on basis of user's position
}

type Storage struct {
	sync.RWMutex
	Data map[string]models.State
}

func (s *Storage) SetReceivedTanks(gameID string, userID string, tank models.Tank) error {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.Data[gameID]; !ok {
		return fmt.Errorf("gameID not found")
	}
	s.Data[gameID].UserTanks[userID] = tank

	return nil
}

func (s *Storage) SetReceivedShots(gameID string, userID string, shot []models.Shot) error {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.Data[gameID]; !ok {
		return fmt.Errorf("gameID not found")
	}
	s.Data[gameID] = models.State{
		UserTanks: s.Data[gameID].UserTanks,
		Shots:     append(s.Data[gameID].Shots, shot...),
		Mines:     s.Data[gameID].Mines,
	}

	return nil
}

func (s *Storage) SetReceivedMines(gameID string, userID string, mines []models.Mine) error {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.Data[gameID]; !ok {
		return fmt.Errorf("gameID not found")
	}
	s.Data[gameID] = models.State{
		UserTanks: s.Data[gameID].UserTanks,
		Shots:     s.Data[gameID].Shots,
		Mines:     append(s.Data[gameID].Mines, mines...),
	}

	return nil
}

func (s *Storage) GetState(gameID string) (*models.State, error) {
	s.RLock()
	defer s.RUnlock()

	state, ok := s.Data[gameID]
	if !ok {
		return nil, fmt.Errorf("gameID not found")
	}

	return &state, fmt.Errorf("no games found")
}

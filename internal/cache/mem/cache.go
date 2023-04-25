package mem

import (
	"ratingserver/internal/domain"
	"ratingserver/internal/normalize"
	"sync"
)

type Cache struct {
	mu      sync.RWMutex
	valid   bool
	players map[string]domain.Player
}

func New() *Cache {
	return &Cache{
		players: make(map[string]domain.Player),
	}
}

func (c *Cache) Update(players []domain.Player) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.players = make(map[string]domain.Player)
	for i := range players {
		name := normalize.Name(players[i].Name)
		c.players[name] = players[i]
	}
	c.valid = true
}

func (c *Cache) GetPlayerByName(name string) (domain.Player, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	name = normalize.Name(name)
	player, ok := c.players[name]
	if !ok {
		return domain.Player{}, false
	}
	return player, true
}

func (c *Cache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.valid = false
	c.players = make(map[string]domain.Player)
}

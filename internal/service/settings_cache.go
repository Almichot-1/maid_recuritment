package service

import (
	"sync"
	"time"

	"maid-recruitment-tracking/internal/domain"
)

// SettingsCacheService provides in-memory caching for platform settings
type SettingsCacheService struct {
	reader    PlatformSettingsReader
	cache     *domain.PlatformSettings
	cacheTTL  time.Duration
	lastFetch time.Time
	mu        sync.RWMutex
}

// NewSettingsCacheService creates a new settings cache service
// cacheTTL specifies how long the settings are cached before a fresh fetch from the database
func NewSettingsCacheService(reader PlatformSettingsReader, cacheTTL time.Duration) *SettingsCacheService {
	if cacheTTL <= 0 {
		cacheTTL = 5 * time.Minute // Default to 5 minutes
	}
	return &SettingsCacheService{
		reader:   reader,
		cacheTTL: cacheTTL,
	}
}

// Get retrieves platform settings from cache if available and not expired,
// otherwise fetches from the underlying reader and updates the cache
func (s *SettingsCacheService) Get() (*domain.PlatformSettings, error) {
	s.mu.RLock()
	// Check if cache is valid
	if s.cache != nil && time.Since(s.lastFetch) < s.cacheTTL {
		defer s.mu.RUnlock()
		return s.cache, nil
	}
	s.mu.RUnlock()

	// Cache miss or expired, fetch from database
	settings, err := s.reader.Get()
	if err != nil {
		return nil, err
	}

	// Update cache
	s.mu.Lock()
	s.cache = settings
	s.lastFetch = time.Now()
	s.mu.Unlock()

	return settings, nil
}

// InvalidateCache clears the cached settings, forcing a fresh fetch on the next Get() call
func (s *SettingsCacheService) InvalidateCache() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache = nil
	s.lastFetch = time.Time{}
}

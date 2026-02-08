// Маппинг команд и владельцев
package slo

import "strings"

type TeamOwner struct {
	Name         string   `yaml:"name"`
	SlackChannel string   `yaml:"SlackChannel"`
	Email        string   `yaml:"email"`
	Services     []string `yaml:"services"`
}

type OwnerManager struct {
	teams map[string]TeamOwner
}

func NewOwnerManager() *OwnerManager {
	return &OwnerManager{
		teams: make(map[string]TeamOwner),
	}
}

func (o *OwnerManager) LoadFromFile(path string) error {
	// Загружаем команды из YAML файла
	// Можно расширить для загрузки из разных источников
	// если б сделать из CMDB - вообще б огонь было
	return nil
}

func (o *OwnerManager) GetTeam(serviceName string) (TeamOwner, bool) {
	// Ищем команду по имени сервиса или другим критериям
	for _, team := range o.teams {
		for _, service := range team.Services {
			if service == serviceName {
				return team, true
			}
		}
	}
	return TeamOwner{}, false
}

func (o *OwnerManager) SuggestTeam(serviceName string) string {
	// Простая логика предложения команды на основе имени сервиса
	serviceName = strings.ToLower(serviceName)

	switch {
	case strings.Contains(serviceName, "api") || strings.Contains(serviceName, "compute"):
		return "compute"
	case strings.Contains(serviceName, "web") || strings.Contains(serviceName, "ui") || strings.Contains(serviceName, "frontend"):
		return "frontend"
	case strings.Contains(serviceName, "db") || strings.Contains(serviceName, "database"):
		return "dbaas"
	case strings.Contains(serviceName, "data") || strings.Contains(serviceName, "pipeline"):
		return "data"
	default:
		return "compute"
	}
}

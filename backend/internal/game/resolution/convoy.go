package resolution

import (
	"diplomacy-cli/backend/internal/game"
	"strings"
)

func (re *ResolutionEngine) processConvoys() {
	re.convoys = make(map[ConvoyKey][]ProvinceCoast)

	convoyFleets := make([]ProvinceCoast, 0)
	for _, order := range re.orders {
		if order.Type == game.Convoy {
			fleet := NewProvinceCoast(order.OrigTerritory, order.FromCoast)
			convoyFleets = append(convoyFleets, fleet)
		}
	}

	for _, order := range re.orders {
		if order.Type == game.Move && order.UnitType == game.Army {
			origin := NewProvinceCoast(order.OrigTerritory, order.FromCoast)
			destination := NewProvinceCoast(order.To, order.ToCoast)
			convoyKey := ConvoyKey{Origin: origin, Destination: destination}

			path := re.findConvoyPath(origin, destination, convoyFleets)
			if len(path) > 0 {
				re.convoys[convoyKey] = path
			}
		}
	}
}

func (re *ResolutionEngine) findConvoyPath(origin, destination ProvinceCoast, convoyFleets []ProvinceCoast) []ProvinceCoast {
	queue := [][]ProvinceCoast{{origin}}
	visited := make(map[string]bool)
	visited[origin.String()] = true

	for len(queue) > 0 {
		currentPath := queue[0]
		queue = queue[1:]
		current := currentPath[len(currentPath)-1]

		if current.Province == destination.Province {
			return currentPath
		}

		currentProvince := re.board.GetProvince(current.Province)
		if currentProvince == nil {
			continue
		}

		var neighbors []string
		if current.HasCoast() {
			neighbors = currentProvince.CoastNeighbors[current.Coast]
		} else {
			neighbors = currentProvince.FleetNeighbors
		}

		for _, neighborStr := range neighbors {
			neighbor := NewProvinceCoast(neighborStr, "")
			if strings.Contains(neighborStr, "/") {
				parts := strings.Split(neighborStr, "/")
				neighbor = NewProvinceCoast(parts[0], parts[1])
			}

			if visited[neighbor.String()] {
				continue
			}

			isValidNext := false
			if neighbor.Province == destination.Province {
				isValidNext = true
			} else {
				for _, fleet := range convoyFleets {
					if fleet.Province == neighbor.Province {
						isValidNext = true
						break
					}
				}
			}

			if isValidNext {
				visited[neighbor.String()] = true
				newPath := make([]ProvinceCoast, len(currentPath)+1)
				copy(newPath, currentPath)
				newPath[len(currentPath)] = neighbor
				queue = append(queue, newPath)
			}
		}
	}

	return nil
}

func parseProvinceCoastString(s string) ProvinceCoast {
	for i, char := range s {
		if char == '/' {
			return NewProvinceCoast(s[:i], s[i+1:])
		}
	}
	return NewProvinceCoast(s, "")
}

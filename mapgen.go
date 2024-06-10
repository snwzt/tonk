package main

import (
	"math/rand"
	"time"
)

const (
	emptySpace = "⬜" // white color small dot
	stone      = "⚫" // grey dot
	bush       = "🌿" // green dot
	water      = "💧" // blue dot
)

func generateMap(width, height int, obstacleProbability map[string]float64, oceanProbability float64,
	generator *rand.Rand) [][]string {
	area := make([][]string, height)
	for i := range area {
		area[i] = make([]string, width)
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// water
			if y == 0 || y == height-1 || x == 0 || x == width-1 {
				if generator.Float64() < oceanProbability {
					area[y][x] = water
					if y+1 < height {
						area[y+1][x] = water
					}
					if y-1 >= 0 {
						area[y-1][x] = water
					}
					if x+1 < width {
						area[y][x+1] = water
					}
					if x-1 >= 0 {
						area[y][x-1] = water
					}
					continue
				}
			}

			// stone and bush
			r := generator.Float64()
			for k, v := range obstacleProbability {
				if k == "stone" && r < v {
					area[y][x] = stone
					if y+1 < height {
						area[y+1][x] = stone
					}
					if x+1 < width {
						area[y][x+1] = stone
					}
					if y+1 < height && x+1 < width {
						area[y+1][x+1] = stone
					}
					break
				}

				if k == "bush" && r < v {
					area[y][x] = bush
					if y+1 < height {
						area[y+1][x] = bush
					}
					if y-1 >= 0 {
						area[y-1][x] = bush
					}
					if x+1 < width {
						area[y][x+1] = bush
					}
					if x-1 >= 0 {
						area[y][x-1] = bush
					}
					break
				}

				if k == "water" && r < v {
					lakeWidth := 6
					lakeHeight := 4

					for dy := -lakeHeight; dy <= lakeHeight; dy++ {
						for dx := -lakeWidth; dx <= lakeWidth; dx++ {
							if y+dy >= 0 && y+dy < height && x+dx >= 0 && x+dx < width && (dx*dx*lakeHeight*lakeHeight+dy*dy*lakeWidth*lakeWidth) <= (lakeWidth*lakeWidth*lakeHeight*lakeHeight) {
								area[y+dy][x+dx] = water
							}
						}
					}
				}

				r -= v
			}

			if area[y][x] == "" {
				area[y][x] = emptySpace
			}
		}
	}

	return area
}

func NewMap() [][]string {
	generator := rand.New(rand.NewSource(time.Now().UnixNano()))

	width, height := 150, 75

	obstacleProbability := map[string]float64{"bush": 0.009, "stone": 0.003, "water": 0.0009}
	lakeProbability := 0.5

	area := generateMap(width, height, obstacleProbability, lakeProbability, generator)

	return area
}

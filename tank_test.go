package main

import (
	"math"
	"math/rand"
	"testing"
)

func testTank(width, height int) *tank {
	return newTank(width, height, rand.New(rand.NewSource(42)), population{})
}

func TestPopulationOverridesAcrossResize(t *testing.T) {
	fishCount, plantCount, zero := 30, 45, 0
	tests := []struct {
		name       string
		population population
	}{
		{"both", population{fishCount: &fishCount, plantCount: &plantCount}},
		{"fish only", population{fishCount: &fishCount}},
		{"plants only", population{plantCount: &plantCount}},
		{"empty", population{fishCount: &zero, plantCount: &zero}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tank := newTank(80, 24, rand.New(rand.NewSource(42)), tt.population)
			for _, size := range [][2]int{{80, 24}, {500, 300}, {20, 8}, {0, 0}, {120, 40}} {
				tank.resize(size[0], size[1])
				if !tank.usable() {
					tank.update(1)
					continue
				}
				defaults := testTank(size[0], size[1])
				wantFish, wantPlants := len(defaults.fish), len(defaults.plants)
				if tt.population.fishCount != nil {
					wantFish = *tt.population.fishCount
				}
				if tt.population.plantCount != nil {
					wantPlants = *tt.population.plantCount
				}
				if len(tank.fish) != wantFish || len(tank.plants) != wantPlants {
					t.Fatalf("size %v: got %d fish, %d plants; want %d fish, %d plants",
						size, len(tank.fish), len(tank.plants), wantFish, wantPlants)
				}
				tank.update(1)
				assertTankBounds(t, tank)
			}
		})
	}
}

func TestFishMovement(t *testing.T) {
	tests := []struct {
		name      string
		x         float64
		direction int
		seconds   float64
		wantX     float64
		wantDir   int
	}{
		{"right", 5, 1, 1, 9, 1},
		{"left", 5, -1, 0.5, 3, -1},
		{"right wall", 15, 1, 1, 13, -1},
		{"left wall", 2, -1, 1, 4, 1},
		{"multiple reflections", 5, 1, 20, 7, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tank := testTank(20, 8)
			tank.fish = []fish{{x: tt.x, direction: tt.direction, speed: 4, kind: 4, depth: 3}}
			tank.update(tt.seconds)
			f := tank.fish[0]
			if math.Abs(f.x-tt.wantX) > 1e-9 || f.direction != tt.wantDir {
				t.Fatalf("got x=%v direction=%v; want x=%v direction=%v", f.x, f.direction, tt.wantX, tt.wantDir)
			}
		})
	}
}

func TestResizePopulationAndBounds(t *testing.T) {
	tank := testTank(80, 24)
	if len(tank.fish) != 10 || len(tank.plants) != 7 {
		t.Fatalf("unexpected initial population: %d fish, %d plants", len(tank.fish), len(tank.plants))
	}
	first := tank.fish[0]
	tank.resize(500, 300)
	if len(tank.fish) != maxFish || len(tank.plants) != maxPlants {
		t.Fatalf("population caps not applied: %d fish, %d plants", len(tank.fish), len(tank.plants))
	}
	if tank.fish[0] != first {
		t.Fatal("growing replaced a surviving fish")
	}
	tank.fish[0].x, tank.fish[0].depth = 490, 290
	tank.resize(20, 8)
	if len(tank.fish) != 1 || len(tank.plants) != 1 {
		t.Fatalf("small tank population: %d fish, %d plants", len(tank.fish), len(tank.plants))
	}
	for range 100 {
		tank.update(0.7)
		assertTankBounds(t, tank)
	}
	tank.resize(120, 40)
	assertTankBounds(t, tank)
	if tank.plants[0].root <= 10 {
		t.Fatal("plant roots did not redistribute across expanded width")
	}
}

func assertTankBounds(t *testing.T, tank *tank) {
	t.Helper()
	for _, f := range tank.fish {
		if f.x < 1 || f.x+float64(f.width()) > float64(tank.width-1) {
			t.Fatalf("fish extends beyond tank: %+v", f)
		}
		if row := tank.fishRow(f); row < 1 || row > tank.height-2 {
			t.Fatalf("fish row %d outside tank", row)
		}
	}
	for _, p := range tank.plants {
		if p.height > tank.height-3 {
			t.Fatalf("plant too tall: %+v", p)
		}
		for level := 0; level <= p.height; level++ {
			if x := tank.plantX(p, level); x < 1 || x > tank.width-2 {
				t.Fatalf("plant x=%d outside tank", x)
			}
		}
	}
}

func TestUnicodeFishWallBounds(t *testing.T) {
	tank := testTank(100, 30)
	tank.fish = []fish{{kind: 5, direction: 1, x: 90, depth: 3, speed: 1}}
	tank.resize(20, 8)
	if got := tank.fish[0].x; got != 13 {
		t.Fatalf("right edge x=%v; want 13 for a six-column fish", got)
	}
	tank.update(1)
	if f := tank.fish[0]; f.x != 12 || f.direction != -1 {
		t.Fatalf("incorrect reflection: x=%v direction=%v", f.x, f.direction)
	}
}

func TestSpriteWidthsAcrossResize(t *testing.T) {
	for kind := range fishSprites {
		for _, direction := range []int{-1, 1} {
			tank := testTank(100, 30)
			tank.fish = []fish{{kind: kind, direction: direction, x: 90, depth: 20, speed: 7}}
			tank.resize(20, 8)
			for range 50 {
				tank.update(0.2)
				assertTankBounds(t, tank)
			}
		}
	}
}

func TestSmallTankPausesAndResumes(t *testing.T) {
	tank := testTank(80, 24)
	tank.update(1)
	before, elapsed := tank.fish[0], tank.elapsed
	for _, size := range [][2]int{{19, 24}, {80, 7}, {0, 0}} {
		tank.resize(size[0], size[1])
		tank.update(100)
		if tank.elapsed != elapsed || tank.fish[0] != before {
			t.Fatal("simulation advanced below the minimum size")
		}
	}
	tank.resize(80, 24)
	tank.update(0.5)
	if tank.elapsed != elapsed+0.5 || tank.fish[0].x == before.x {
		t.Fatal("simulation failed to resume")
	}
}

func TestPlantWavesWithFixedRoot(t *testing.T) {
	tank := testTank(80, 24)
	p := plant{root: 20, height: 8, phase: 0}
	positions := make(map[int]bool)
	for range 30 {
		if x := tank.plantX(p, 0); x != p.root {
			t.Fatalf("root moved to %d", x)
		}
		positions[tank.plantX(p, p.height)] = true
		tank.update(0.1)
	}
	if len(positions) < 3 {
		t.Fatalf("plant tip did not wave: %v", positions)
	}
}

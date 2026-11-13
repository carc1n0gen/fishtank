package main

import (
	"math"
	"math/rand"

	"github.com/rivo/uniseg"
)

const (
	minWidth  = 20
	minHeight = 8
	maxFish   = 60
	maxPlants = 40
)

var fishSprites = [...][2]string{
	{".", "."},
	{"o", "o"},
	{"-<", ">-"},
	{"o-", "-o"},
	{"O<", ">O"},
	{"<><", "><>"},
	{"<°))><", "><((°>"},
	{"<°)))><", "><(((°>"},
}

type fish struct {
	x         float64
	depth     float64
	speed     float64
	direction int
	kind      int
	phase     float64
}

func (f fish) sprite() string {
	if f.direction < 0 {
		return fishSprites[f.kind][0]
	}
	return fishSprites[f.kind][1]
}

func (f fish) width() int {
	return uniseg.StringWidth(f.sprite())
}

type plant struct {
	root   int
	height int
	phase  float64
}

type population struct {
	fishCount  *int
	plantCount *int
}

type tank struct {
	width, height int
	elapsed       float64
	fish          []fish
	plants        []plant
	random        *rand.Rand
	population    population
}

func newTank(width, height int, random *rand.Rand, population population) *tank {
	t := &tank{random: random, population: population}
	t.resize(width, height)
	return t
}

func (t *tank) usable() bool {
	return t.width >= minWidth && t.height >= minHeight
}

func (t *tank) resize(width, height int) {
	t.width, t.height = width, height
	if !t.usable() {
		return
	}

	fishCount := min(maxFish, max(1, (width-2)*(height-2)/160))
	plantCount := min(maxPlants, max(1, (width-2)/10))
	if t.population.fishCount != nil {
		fishCount = *t.population.fishCount
	}
	if t.population.plantCount != nil {
		plantCount = *t.population.plantCount
	}
	t.fish = t.fish[:min(len(t.fish), fishCount)]
	for len(t.fish) < fishCount {
		f := fish{
			kind:      t.random.Intn(len(fishSprites)),
			speed:     2 + t.random.Float64()*5,
			direction: 1,
			phase:     t.random.Float64() * 2 * math.Pi,
			depth:     1 + t.random.Float64()*float64(height-3),
		}
		if t.random.Intn(2) == 0 {
			f.direction = -1
		}
		f.x = 1 + t.random.Float64()*float64(width-2-f.width())
		t.fish = append(t.fish, f)
	}
	for i := range t.fish {
		f := &t.fish[i]
		f.x = clamp(f.x, 1, float64(width-1-f.width()))
		f.depth = clamp(f.depth, 1, float64(height-2))
	}

	t.plants = t.plants[:min(len(t.plants), plantCount)]
	for len(t.plants) < plantCount {
		t.plants = append(t.plants, plant{
			height: 2 + t.random.Intn(max(1, (height-2)/2)),
			phase:  t.random.Float64() * 2 * math.Pi,
		})
	}
	for i := range t.plants {
		p := &t.plants[i]
		p.root = 2 + (i+1)*(width-5)/(plantCount+1)
		p.height = min(p.height, height-3)
	}
}

func (t *tank) update(seconds float64) {
	if !t.usable() || seconds <= 0 {
		return
	}
	t.elapsed += seconds
	for i := range t.fish {
		f := &t.fish[i]
		span := float64(t.width - 2 - f.width())
		phase := f.x - 1
		if f.direction < 0 {
			phase = 2*span - phase
		}
		// Reflect a complete travel cycle, including ticks crossing several walls.
		phase = math.Mod(phase+f.speed*seconds, 2*span)
		if phase < span {
			f.x, f.direction = 1+phase, 1
		} else {
			f.x, f.direction = 1+2*span-phase, -1
		}
	}
}

func (t *tank) fishRow(f fish) int {
	return int(math.Round(clamp(f.depth+0.6*math.Sin(t.elapsed*0.8+f.phase), 1, float64(t.height-2))))
}

func (t *tank) plantX(p plant, level int) int {
	amplitude := 1.8 * float64(level) / float64(p.height)
	offset := int(math.Round(amplitude * math.Sin(t.elapsed*1.7+p.phase-float64(level)*0.65)))
	return min(t.width-2, max(1, p.root+offset))
}

func clamp(value, low, high float64) float64 {
	return math.Min(high, math.Max(low, value))
}

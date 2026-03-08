package assets

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
)

// ImageSize is the standard image size for all bot images
const ImageSize = 512

// GenerateAllImages creates all game images in the assets folder
func GenerateAllImages(baseDir string) error {
	specs := getAllImageSpecs()
	for _, spec := range specs {
		path := filepath.Join(baseDir, spec.path)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		// Skip if already exists
		if _, err := os.Stat(path); err == nil {
			continue
		}
		img := generateImage(spec)
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		if err := png.Encode(f, img); err != nil {
			f.Close()
			return err
		}
		f.Close()
	}
	return nil
}

// =============================================
// IMAGE SPECS
// =============================================

type imageSpec struct {
	path    string
	bg1     color.RGBA // primary background color
	bg2     color.RGBA // secondary background color
	accent  color.RGBA // accent/glow color
	symbol  [][]int    // pixel art symbol (0=transparent, 1=light, 2=dark, 3=accent)
	borderColor color.RGBA
}

func getAllImageSpecs() []imageSpec {
	return []imageSpec{
		// ── UI ──────────────────────────────────────────────
		{
			path: "ui/welcome.png",
			bg1: c(15, 15, 40), bg2: c(30, 30, 80), accent: c(180, 140, 60),
			borderColor: c(180, 140, 60),
			symbol: castleSymbol,
		},
		{
			path: "ui/menu.png",
			bg1: c(20, 30, 55), bg2: c(35, 55, 90), accent: c(100, 160, 255),
			borderColor: c(100, 160, 255),
			symbol: menuSymbol,
		},
		{
			path: "ui/status.png",
			bg1: c(25, 50, 35), bg2: c(40, 80, 55), accent: c(80, 220, 120),
			borderColor: c(80, 200, 100),
			symbol: statusSymbol,
		},
		{
			path: "ui/inventory.png",
			bg1: c(50, 35, 20), bg2: c(85, 60, 30), accent: c(220, 170, 80),
			borderColor: c(200, 150, 60),
			symbol: bagSymbol,
		},
		{
			path: "ui/skills.png",
			bg1: c(45, 15, 60), bg2: c(75, 30, 100), accent: c(200, 120, 255),
			borderColor: c(180, 100, 240),
			symbol: skillSymbol,
		},
		{
			path: "ui/shop.png",
			bg1: c(55, 45, 15), bg2: c(90, 75, 25), accent: c(255, 220, 50),
			borderColor: c(240, 200, 40),
			symbol: coinSymbol,
		},
		{
			path: "ui/travel.png",
			bg1: c(15, 40, 55), bg2: c(25, 65, 90), accent: c(70, 200, 240),
			borderColor: c(60, 180, 220),
			symbol: mapSymbol,
		},
		{
			path: "ui/combat.png",
			bg1: c(50, 10, 10), bg2: c(90, 20, 20), accent: c(240, 80, 60),
			borderColor: c(220, 60, 40),
			symbol: swordSymbol,
		},
		{
			path: "ui/victory.png",
			bg1: c(10, 50, 15), bg2: c(20, 90, 30), accent: c(255, 230, 50),
			borderColor: c(240, 210, 40),
			symbol: trophySymbol,
		},
		{
			path: "ui/defeat.png",
			bg1: c(30, 10, 10), bg2: c(60, 15, 15), accent: c(180, 60, 50),
			borderColor: c(140, 40, 30),
			symbol: skullSymbol,
		},
		{
			path: "ui/rest.png",
			bg1: c(10, 20, 50), bg2: c(20, 35, 80), accent: c(120, 160, 255),
			borderColor: c(100, 140, 240),
			symbol: moonSymbol,
		},

		// ── RACES ────────────────────────────────────────────
		{
			path: "races/human.png",
			bg1: c(40, 70, 100), bg2: c(55, 95, 140), accent: c(200, 220, 255),
			borderColor: c(150, 190, 230),
			symbol: humanSymbol,
		},
		{
			path: "races/elf.png",
			bg1: c(15, 60, 35), bg2: c(25, 100, 55), accent: c(120, 255, 160),
			borderColor: c(80, 220, 120),
			symbol: elfSymbol,
		},
		{
			path: "races/dwarf.png",
			bg1: c(70, 50, 20), bg2: c(110, 80, 30), accent: c(220, 180, 80),
			borderColor: c(200, 160, 60),
			symbol: dwarfSymbol,
		},
		{
			path: "races/halforc.png",
			bg1: c(40, 55, 20), bg2: c(60, 85, 30), accent: c(160, 220, 80),
			borderColor: c(140, 200, 60),
			symbol: orcSymbol,
		},

		// ── CLASSES ──────────────────────────────────────────
		{
			path: "classes/warrior.png",
			bg1: c(60, 20, 20), bg2: c(100, 35, 35), accent: c(240, 100, 80),
			borderColor: c(220, 80, 60),
			symbol: warriorSymbol,
		},
		{
			path: "classes/mage.png",
			bg1: c(30, 15, 65), bg2: c(50, 25, 110), accent: c(160, 120, 255),
			borderColor: c(140, 100, 240),
			symbol: mageSymbol,
		},
		{
			path: "classes/rogue.png",
			bg1: c(20, 20, 20), bg2: c(40, 40, 40), accent: c(180, 180, 180),
			borderColor: c(150, 150, 150),
			symbol: rogueSymbol,
		},
		{
			path: "classes/archer.png",
			bg1: c(25, 50, 20), bg2: c(40, 80, 30), accent: c(140, 210, 90),
			borderColor: c(120, 190, 70),
			symbol: archerSymbol,
		},

		// ── MONSTERS ─────────────────────────────────────────
		{
			path: "monsters/rat.png",
			bg1: c(45, 35, 30), bg2: c(70, 55, 45), accent: c(180, 160, 120),
			borderColor: c(160, 140, 100),
			symbol: ratSymbol,
		},
		{
			path: "monsters/goblin.png",
			bg1: c(30, 50, 20), bg2: c(50, 80, 30), accent: c(120, 200, 70),
			borderColor: c(100, 180, 50),
			symbol: goblinSymbol,
		},
		{
			path: "monsters/slime.png",
			bg1: c(20, 60, 40), bg2: c(30, 100, 65), accent: c(80, 230, 150),
			borderColor: c(60, 210, 130),
			symbol: slimeSymbol,
		},
		{
			path: "monsters/wolf.png",
			bg1: c(30, 30, 50), bg2: c(50, 50, 80), accent: c(150, 150, 220),
			borderColor: c(130, 130, 200),
			symbol: wolfSymbol,
		},
		{
			path: "monsters/orc.png",
			bg1: c(40, 55, 15), bg2: c(65, 90, 25), accent: c(160, 220, 60),
			borderColor: c(140, 200, 40),
			symbol: bigOrcSymbol,
		},
		{
			path: "monsters/troll.png",
			bg1: c(35, 55, 30), bg2: c(55, 85, 45), accent: c(140, 210, 120),
			borderColor: c(120, 190, 100),
			symbol: trollSymbol,
		},
		{
			path: "monsters/bandit_leader.png",
			bg1: c(40, 30, 20), bg2: c(65, 50, 30), accent: c(200, 160, 90),
			borderColor: c(180, 140, 70),
			symbol: banditSymbol,
		},
		{
			path: "monsters/bat.png",
			bg1: c(20, 10, 35), bg2: c(40, 20, 60), accent: c(150, 80, 200),
			borderColor: c(130, 60, 180),
			symbol: batSymbol,
		},
		{
			path: "monsters/spider.png",
			bg1: c(20, 10, 10), bg2: c(40, 20, 20), accent: c(200, 80, 80),
			borderColor: c(180, 60, 60),
			symbol: spiderSymbol,
		},
		{
			path: "monsters/golem.png",
			bg1: c(50, 50, 60), bg2: c(80, 80, 95), accent: c(180, 180, 220),
			borderColor: c(160, 160, 200),
			symbol: golemSymbol,
		},
		{
			path: "monsters/undead_knight.png",
			bg1: c(20, 30, 40), bg2: c(35, 50, 65), accent: c(100, 200, 180),
			borderColor: c(80, 180, 160),
			symbol: undeadSymbol,
		},
		{
			path: "monsters/demon.png",
			bg1: c(60, 10, 10), bg2: c(100, 15, 15), accent: c(255, 60, 40),
			borderColor: c(240, 40, 20),
			symbol: demonSymbol,
		},
		{
			path: "monsters/necromancer.png",
			bg1: c(25, 10, 40), bg2: c(45, 20, 70), accent: c(160, 80, 230),
			borderColor: c(140, 60, 210),
			symbol: necroSymbol,
		},
		{
			path: "monsters/vampire_lord.png",
			bg1: c(40, 5, 20), bg2: c(70, 10, 35), accent: c(220, 50, 100),
			borderColor: c(200, 30, 80),
			symbol: vampireSymbol,
		},
		{
			path: "monsters/dragon_young.png",
			bg1: c(55, 25, 10), bg2: c(90, 40, 15), accent: c(240, 140, 50),
			borderColor: c(220, 120, 30),
			symbol: dragonYoungSymbol,
		},
		{
			path: "monsters/dragon_elder.png",
			bg1: c(60, 10, 10), bg2: c(100, 15, 15), accent: c(255, 80, 20),
			borderColor: c(240, 60, 0),
			symbol: dragonElderSymbol,
		},

		// ── NEW MONSTERS ─────────────────────────────────────
		{
			path: "monsters/mushroom.png",
			bg1: c(40, 25, 50), bg2: c(65, 40, 80), accent: c(200, 120, 240),
			borderColor: c(180, 100, 220),
			symbol: slimeSymbol,
		},
		{
			path: "monsters/crow.png",
			bg1: c(15, 15, 20), bg2: c(30, 30, 40), accent: c(140, 140, 180),
			borderColor: c(120, 120, 160),
			symbol: genericFlyingSymbol,
		},
		{
			path: "monsters/harpy.png",
			bg1: c(50, 20, 55), bg2: c(80, 35, 90), accent: c(210, 100, 230),
			borderColor: c(190, 80, 210),
			symbol: genericFlyingSymbol,
		},
		{
			path: "monsters/werewolf.png",
			bg1: c(30, 25, 35), bg2: c(55, 45, 65), accent: c(180, 160, 210),
			borderColor: c(160, 140, 190),
			symbol: genericBeastSymbol,
		},
		{
			path: "monsters/stone_golem_shard.png",
			bg1: c(55, 50, 45), bg2: c(85, 80, 70), accent: c(200, 190, 170),
			borderColor: c(180, 170, 150),
			symbol: genericRockSymbol,
		},
		{
			path: "monsters/crystal_wraith.png",
			bg1: c(10, 40, 55), bg2: c(20, 70, 90), accent: c(80, 220, 255),
			borderColor: c(60, 200, 240),
			symbol: genericRockSymbol,
		},
		{
			path: "monsters/shadow_assassin.png",
			bg1: c(10, 10, 15), bg2: c(20, 20, 30), accent: c(100, 80, 160),
			borderColor: c(80, 60, 140),
			symbol: rogueSymbol,
		},
		{
			path: "monsters/lich.png",
			bg1: c(15, 5, 30), bg2: c(30, 10, 55), accent: c(150, 60, 220),
			borderColor: c(130, 40, 200),
			symbol: skullSymbol,
		},
		{
			path: "monsters/wyvern.png",
			bg1: c(15, 45, 20), bg2: c(25, 75, 35), accent: c(80, 210, 100),
			borderColor: c(60, 190, 80),
			symbol: genericDragonSymbol,
		},
		{
			path: "monsters/phoenix.png",
			bg1: c(70, 35, 5), bg2: c(110, 55, 10), accent: c(255, 180, 40),
			borderColor: c(240, 160, 20),
			symbol: genericFlyingSymbol,
		},

		// ── MAPS ─────────────────────────────────────────────
		{
			path: "maps/village.png",
			bg1: c(30, 70, 40), bg2: c(50, 110, 65), accent: c(150, 230, 160),
			borderColor: c(120, 210, 130),
			symbol: villageMapSymbol,
		},
		{
			path: "maps/village_outskirts.png",
			bg1: c(50, 75, 30), bg2: c(80, 120, 45), accent: c(180, 230, 110),
			borderColor: c(160, 210, 90),
			symbol: outskirtMapSymbol,
		},
		{
			path: "maps/dark_forest.png",
			bg1: c(10, 30, 15), bg2: c(20, 55, 25), accent: c(60, 160, 80),
			borderColor: c(40, 140, 60),
			symbol: forestMapSymbol,
		},
		{
			path: "maps/forest_camp.png",
			bg1: c(55, 40, 20), bg2: c(90, 65, 30), accent: c(220, 170, 80),
			borderColor: c(200, 150, 60),
			symbol: campMapSymbol,
		},
		{
			path: "maps/crystal_cave.png",
			bg1: c(20, 30, 60), bg2: c(30, 50, 100), accent: c(100, 180, 255),
			borderColor: c(80, 160, 240),
			symbol: caveMapSymbol,
		},
		{
			path: "maps/ancient_dungeon.png",
			bg1: c(30, 20, 10), bg2: c(55, 35, 15), accent: c(180, 130, 60),
			borderColor: c(160, 110, 40),
			symbol: dungeonMapSymbol,
		},
		{
			path: "maps/dungeon_outpost.png",
			bg1: c(50, 40, 30), bg2: c(80, 65, 50), accent: c(200, 180, 140),
			borderColor: c(180, 160, 120),
			symbol: outpostMapSymbol,
		},
		{
			path: "maps/dragon_peak.png",
			bg1: c(50, 20, 10), bg2: c(90, 35, 15), accent: c(240, 120, 40),
			borderColor: c(220, 100, 20),
			symbol: peakMapSymbol,
		},

		// ── ITEMS ─────────────────────────────────────────────
		{
			path: "items/weapons.png",
			bg1: c(55, 25, 25), bg2: c(90, 40, 40), accent: c(240, 130, 100),
			borderColor: c(220, 110, 80),
			symbol: weaponItemSymbol,
		},
		{
			path: "items/armors.png",
			bg1: c(40, 45, 55), bg2: c(65, 75, 90), accent: c(160, 190, 230),
			borderColor: c(140, 170, 210),
			symbol: armorItemSymbol,
		},
		{
			path: "items/consumables.png",
			bg1: c(20, 50, 30), bg2: c(30, 80, 50), accent: c(100, 220, 130),
			borderColor: c(80, 200, 110),
			symbol: potionItemSymbol,
		},
		{
			path: "items/accessories.png",
			bg1: c(45, 25, 55), bg2: c(75, 40, 90), accent: c(220, 170, 240),
			borderColor: c(200, 140, 220),
			symbol: coinSymbol,
		},
		{
			path: "items/materials.png",
			bg1: c(45, 45, 35), bg2: c(75, 75, 55), accent: c(215, 200, 140),
			borderColor: c(190, 175, 120),
			symbol: genericRockSymbol,
		},
	}
}

// =============================================
// IMAGE RENDERER
// =============================================

func c(r, g, b uint8) color.RGBA {
	return color.RGBA{r, g, b, 255}
}

func generateImage(spec imageSpec) image.Image {
	size := ImageSize
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Draw gradient background
	drawGradientBG(img, spec.bg1, spec.bg2)

	// Draw decorative corner ornaments
	drawCornerOrnaments(img, spec.accent)

	// Draw center glow circle
	drawGlowCircle(img, size/2, size/2, size/3, spec.accent, 0.25)

	// Draw inner glow circle
	drawGlowCircle(img, size/2, size/2, size/5, spec.accent, 0.45)

	// Draw pixel art symbol
	if spec.symbol != nil {
		drawPixelSymbol(img, spec.symbol, size/2, size/2, size/2-40, spec.accent, spec.bg2)
	}

	// Draw border
	drawBorder(img, spec.borderColor, 6)

	// Draw inner border
	drawBorder2(img, spec.borderColor, 14, 0.4)

	return img
}

func drawGradientBG(img *image.RGBA, c1, c2 color.RGBA) {
	size := img.Bounds().Max.X
	for y := 0; y < size; y++ {
		t := float64(y) / float64(size)
		r := uint8(float64(c1.R)*(1-t) + float64(c2.R)*t)
		g := uint8(float64(c1.G)*(1-t) + float64(c2.G)*t)
		b := uint8(float64(c1.B)*(1-t) + float64(c2.B)*t)
		for x := 0; x < size; x++ {
			img.SetRGBA(x, y, color.RGBA{r, g, b, 255})
		}
	}
}

func drawGlowCircle(img *image.RGBA, cx, cy, radius int, col color.RGBA, alpha float64) {
	size := img.Bounds().Max.X
	r2 := float64(radius * radius)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x - cx)
			dy := float64(y - cy)
			dist2 := dx*dx + dy*dy
			if dist2 < r2 {
				t := 1.0 - math.Sqrt(dist2)/float64(radius)
				a := t * alpha
				orig := img.RGBAAt(x, y)
				nr := uint8(float64(orig.R)*(1-a) + float64(col.R)*a)
				ng := uint8(float64(orig.G)*(1-a) + float64(col.G)*a)
				nb := uint8(float64(orig.B)*(1-a) + float64(col.B)*a)
				img.SetRGBA(x, y, color.RGBA{nr, ng, nb, 255})
			}
		}
	}
}

func drawBorder(img *image.RGBA, col color.RGBA, thickness int) {
	size := img.Bounds().Max.X
	for i := 0; i < thickness; i++ {
		alpha := uint8(255 - i*30)
		col2 := color.RGBA{col.R, col.G, col.B, alpha}
		for p := 0; p < size; p++ {
			img.SetRGBA(p, i, col2)
			img.SetRGBA(p, size-1-i, col2)
			img.SetRGBA(i, p, col2)
			img.SetRGBA(size-1-i, p, col2)
		}
	}
}

func drawBorder2(img *image.RGBA, col color.RGBA, offset int, alpha float64) {
	size := img.Bounds().Max.X
	col2 := color.RGBA{col.R, col.G, col.B, uint8(alpha * 255)}
	for p := offset; p < size-offset; p++ {
		img.SetRGBA(p, offset, col2)
		img.SetRGBA(p, size-1-offset, col2)
		img.SetRGBA(offset, p, col2)
		img.SetRGBA(size-1-offset, p, col2)
	}
}

func drawCornerOrnaments(img *image.RGBA, col color.RGBA) {
	size := img.Bounds().Max.X
	ornament := []image.Point{
		{20, 20}, {30, 20}, {40, 20}, {20, 30}, {20, 40},
	}
	corners := [][2]int{{0, 0}, {size - 1, 0}, {0, size - 1}, {size - 1, size - 1}}
	sx := [4]int{1, -1, 1, -1}
	sy := [4]int{1, 1, -1, -1}
	col2 := color.RGBA{col.R, col.G, col.B, 200}
	for i, corner := range corners {
		for _, p := range ornament {
			x := corner[0] + p.X*sx[i]
			y := corner[1] + p.Y*sy[i]
			if x >= 0 && x < size && y >= 0 && y < size {
				img.SetRGBA(x, y, col2)
			}
		}
	}
}

// drawPixelSymbol draws a pixel art symbol centered at (cx, cy) with given size
func drawPixelSymbol(img *image.RGBA, symbol [][]int, cx, cy, maxSize int, accent, dark color.RGBA) {
	if len(symbol) == 0 {
		return
	}
	rows := len(symbol)
	cols := len(symbol[0])
	scale := maxSize / rows
	if scale < 1 {
		scale = 1
	}
	offsetX := cx - (cols*scale)/2
	offsetY := cy - (rows*scale)/2

	light := color.RGBA{
		clamp(int(accent.R) + 60),
		clamp(int(accent.G) + 60),
		clamp(int(accent.B) + 60),
		255,
	}
	shadow := color.RGBA{
		clamp(int(dark.R) - 20),
		clamp(int(dark.G) - 20),
		clamp(int(dark.B) - 20),
		255,
	}

	size := img.Bounds().Max.X
	for row, line := range symbol {
		for col, val := range line {
			if val == 0 {
				continue
			}
			var px color.RGBA
			switch val {
			case 1:
				px = light
			case 2:
				px = shadow
			case 3:
				px = accent
			}
			for dy := 0; dy < scale; dy++ {
				for dx := 0; dx < scale; dx++ {
					x := offsetX + col*scale + dx
					y := offsetY + row*scale + dy
					if x >= 0 && x < size && y >= 0 && y < size {
						// Slight shadow offset for depth
						if dx == scale-1 || dy == scale-1 {
							px2 := color.RGBA{
								clamp(int(px.R) - 30),
								clamp(int(px.G) - 30),
								clamp(int(px.B) - 30),
								255,
							}
							img.SetRGBA(x, y, px2)
						} else {
							img.SetRGBA(x, y, px)
						}
					}
				}
			}
		}
	}
}

func clamp(v int) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

// =============================================
// PIXEL ART SYMBOLS (16x16 grid, values 0-3)
// 0 = transparent, 1 = light, 2 = dark, 3 = accent
// =============================================

var _ = draw.Draw // ensure draw is used

var castleSymbol = [][]int{
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 1, 0, 0, 1, 0, 0, 0, 1, 0, 0, 1, 0, 0, 0},
	{0, 0, 1, 1, 1, 1, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0},
	{0, 0, 1, 3, 3, 1, 1, 1, 1, 1, 3, 3, 1, 0, 0, 0},
	{0, 0, 1, 3, 3, 1, 3, 3, 3, 1, 3, 3, 1, 0, 0, 0},
	{0, 0, 1, 1, 1, 1, 3, 3, 3, 1, 1, 1, 1, 0, 0, 0},
	{0, 0, 0, 1, 3, 1, 3, 3, 3, 1, 3, 1, 0, 0, 0, 0},
	{0, 0, 0, 1, 1, 1, 1, 2, 1, 1, 1, 1, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 1, 1, 2, 1, 1, 3, 1, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 1, 1, 2, 1, 1, 3, 1, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 1, 1, 2, 1, 1, 3, 1, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 1, 1, 2, 1, 1, 3, 1, 0, 0, 0, 0},
	{0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0},
	{0, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var menuSymbol = [][]int{
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0},
	{0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0},
	{0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0},
	{0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0},
	{0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0},
	{0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0},
	{0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var statusSymbol = [][]int{
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 0, 0, 1, 3, 1, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 0, 0, 1, 3, 1, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var bagSymbol = [][]int{
	{0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 1, 3, 3, 3, 1, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0},
	{0, 0, 1, 1, 3, 3, 3, 3, 3, 3, 3, 1, 1, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 0, 1, 3, 1, 3, 3, 3, 3, 3, 1, 3, 1, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 0, 1, 3, 1, 3, 3, 3, 3, 3, 1, 3, 1, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 0, 1, 1, 3, 3, 3, 3, 3, 3, 3, 1, 1, 0, 0, 0},
	{0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var skillSymbol = [][]int{
	{0, 0, 0, 0, 0, 0, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 1, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 1, 3, 1, 3, 3, 3, 1, 0, 0, 0},
	{0, 1, 3, 3, 3, 1, 3, 3, 3, 1, 3, 3, 3, 1, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 1, 3, 1, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var coinSymbol = [][]int{
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 1, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 1, 3, 1, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 1, 3, 1, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 1, 1, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var mapSymbol = [][]int{
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 1, 1, 1, 1, 1, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0},
	{0, 1, 3, 3, 3, 1, 0, 0, 0, 1, 3, 3, 3, 3, 1, 0},
	{0, 1, 3, 3, 3, 1, 0, 0, 0, 1, 3, 3, 3, 3, 1, 0},
	{0, 1, 3, 3, 3, 1, 0, 0, 0, 1, 3, 3, 3, 3, 1, 0},
	{0, 1, 1, 1, 1, 1, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 1, 1, 1, 1, 1, 1, 0, 0, 0, 1, 1, 1, 1, 0, 0},
	{0, 1, 3, 3, 3, 3, 1, 0, 0, 1, 3, 3, 3, 3, 1, 0},
	{0, 1, 3, 3, 3, 3, 1, 0, 0, 1, 3, 3, 3, 3, 1, 0},
	{0, 1, 3, 3, 3, 3, 1, 0, 0, 1, 3, 3, 3, 3, 1, 0},
	{0, 1, 1, 1, 1, 1, 1, 0, 0, 0, 1, 1, 1, 1, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var swordSymbol = [][]int{
	{0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 1, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 3, 3, 3, 3, 1, 1, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{1, 1, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{1, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 1, 1, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 1, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 2, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var trophySymbol = [][]int{
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0},
	{0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0},
	{0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0},
	{0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var skullSymbol = [][]int{
	{0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 0, 1, 3, 2, 2, 3, 3, 3, 2, 2, 3, 1, 0, 0, 0},
	{0, 0, 1, 3, 2, 2, 3, 3, 3, 2, 2, 3, 1, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 2, 3, 2, 3, 3, 3, 1, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 0, 1, 1, 2, 1, 3, 1, 2, 1, 1, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 2, 1, 3, 1, 2, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var moonSymbol = [][]int{
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 1, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 1, 3, 3, 1, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 1, 3, 3, 1, 0, 3, 3, 3, 0, 0, 0, 0, 0, 0, 0},
	{0, 1, 3, 3, 1, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 1, 3, 3, 3, 1, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

// Character/Race symbols
var humanSymbol = [][]int{
	{0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 1, 3, 3, 1, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 1, 3, 3, 1, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 0, 0, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 0, 0, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var elfSymbol = [][]int{
	{0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 1, 3, 3, 3, 3, 1, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 1, 3, 3, 1, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 1, 3, 3, 3, 1, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 1, 3, 3, 1, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 0, 0, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 0, 0, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var dwarfSymbol = [][]int{
	{0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 1, 1, 3, 3, 1, 1, 1, 0, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 0, 1, 1, 1, 3, 3, 1, 1, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 1, 1, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var orcSymbol = [][]int{
	{0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 2, 3, 2, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 1, 3, 3, 3, 3, 1, 1, 0, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 0, 1, 1, 1, 3, 3, 1, 1, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 1, 1, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

// Class symbols
var warriorSymbol = swordSymbol

var mageSymbol = [][]int{
	{0, 0, 0, 0, 0, 0, 1, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 1, 3, 1, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 3, 1, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 1, 0, 1, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var rogueSymbol = [][]int{
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 3, 3, 3, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 1, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 1, 3, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 1, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{1, 3, 3, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var archerSymbol = [][]int{
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 0, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0},
	{0, 0, 1, 3, 0, 0, 1, 1, 1, 0, 0, 1, 0, 0, 0, 0},
	{0, 1, 3, 0, 0, 1, 3, 0, 0, 1, 0, 0, 1, 0, 0, 0},
	{1, 3, 0, 0, 1, 3, 0, 0, 0, 0, 1, 0, 0, 1, 0, 0},
	{1, 1, 3, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0},
	{0, 1, 1, 3, 0, 0, 1, 1, 1, 0, 0, 0, 1, 0, 0, 0},
	{0, 0, 1, 1, 3, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0},
	{0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

// Monster symbols - simplified shapes
var ratSymbol = [][]int{
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 3, 3, 1, 0},
	{0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 3, 3, 3, 1, 0},
	{0, 0, 0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0},
	{0, 0, 0, 0, 1, 3, 3, 2, 3, 3, 3, 3, 3, 1, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0},
	{0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 0, 1, 1, 3, 3, 3, 3, 3, 3, 1, 1, 0, 0, 0, 0},
	{0, 0, 1, 0, 1, 1, 1, 1, 1, 1, 0, 1, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var goblinSymbol = [][]int{
	{0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 1, 1, 1, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 2, 3, 2, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 2, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 1, 3, 3, 3, 3, 3, 1, 1, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 0, 1, 3, 1, 3, 3, 3, 3, 3, 1, 3, 1, 0, 0, 0},
	{0, 0, 0, 1, 1, 1, 3, 3, 1, 1, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 0, 0, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

// Reuse simpler shapes for remaining monsters/maps to keep file size manageable
var slimeSymbol = [][]int{
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 2, 3, 3, 2, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 2, 2, 2, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 1, 3, 3, 1, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 1, 3, 1, 0, 1, 1, 0, 1, 3, 1, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

// Generic shapes for remaining monsters/maps (distinct colors differentiate them)
var wolfSymbol    = genericBeastSymbol
var bigOrcSymbol  = genericWarriorSymbol
var trollSymbol   = genericBigBeastSymbol
var banditSymbol  = genericWarriorSymbol
var batSymbol     = genericFlyingSymbol
var spiderSymbol  = genericSpiderSymbol
var golemSymbol   = genericRockSymbol
var undeadSymbol  = skullSymbol
var demonSymbol   = genericDemonSymbol
var necroSymbol   = mageSymbol
var vampireSymbol = genericVampireSymbol
var dragonYoungSymbol = genericDragonSymbol
var dragonElderSymbol = genericDragonSymbol

var villageMapSymbol  = castleSymbol
var outskirtMapSymbol = genericTreeSymbol
var forestMapSymbol   = genericForestSymbol
var campMapSymbol     = genericCampSymbol
var caveMapSymbol     = genericCaveSymbol
var dungeonMapSymbol  = genericDungeonSymbol
var outpostMapSymbol  = castleSymbol
var peakMapSymbol     = genericMountainSymbol

var weaponItemSymbol    = swordSymbol
var armorItemSymbol     = genericShieldSymbol
var potionItemSymbol    = genericPotionSymbol

// Generic symbols
var genericBeastSymbol = [][]int{
	{0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0},
	{0, 1, 1, 1, 0, 0, 0, 0, 0, 1, 1, 1, 0, 0, 0, 0},
	{0, 1, 3, 3, 1, 1, 1, 1, 1, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 1, 3, 2, 3, 3, 3, 2, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 2, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 1, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 0, 1, 1, 0, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 1, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var genericWarriorSymbol = [][]int{
	{0, 0, 0, 0, 0, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 1, 1, 1, 3, 3, 3, 1, 1, 1, 0, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 1, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 0, 0, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 1, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var genericBigBeastSymbol = [][]int{
	{0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0},
	{0, 1, 1, 1, 0, 0, 0, 0, 0, 0, 1, 1, 1, 0, 0, 0},
	{1, 3, 3, 3, 1, 1, 1, 1, 1, 1, 3, 3, 3, 1, 0, 0},
	{1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0},
	{0, 1, 3, 3, 2, 3, 3, 3, 3, 2, 3, 3, 1, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 2, 2, 2, 2, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 0, 1, 1, 1, 3, 3, 1, 1, 1, 0, 0, 0, 0, 0},
	{0, 0, 1, 3, 0, 0, 0, 0, 0, 0, 3, 1, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var genericFlyingSymbol = [][]int{
	{0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 1, 0, 0, 0, 1, 3, 3, 3, 1, 0, 0, 0, 1, 0, 0},
	{1, 3, 1, 1, 1, 3, 3, 3, 3, 3, 1, 1, 1, 3, 1, 0},
	{1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0},
	{0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 0, 0, 1, 1, 1, 1, 3, 1, 1, 1, 1, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var genericSpiderSymbol = [][]int{
	{0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0},
	{0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 1, 1, 1, 3, 3, 3, 3, 3, 1, 1, 1, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 2, 3, 2, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 1, 1, 1, 3, 3, 3, 3, 3, 1, 1, 1, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0},
	{0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var genericRockSymbol = [][]int{
	{0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 1, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 1, 3, 3, 3, 2, 3, 3, 2, 3, 3, 3, 1, 0, 0, 0},
	{0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 1, 3, 3, 3, 3, 2, 2, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 1, 1, 3, 3, 1, 1, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var genericDemonSymbol = [][]int{
	{0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0},
	{0, 1, 3, 1, 0, 1, 1, 1, 1, 0, 1, 3, 1, 0, 0, 0},
	{0, 0, 1, 1, 1, 3, 3, 3, 3, 1, 1, 1, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 2, 3, 3, 2, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 2, 2, 2, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 1, 1, 3, 3, 3, 3, 3, 3, 1, 1, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 0, 1, 1, 3, 3, 3, 3, 1, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 0, 0, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 1, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var genericVampireSymbol = [][]int{
	{0, 1, 0, 0, 0, 1, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0},
	{0, 0, 1, 0, 1, 3, 3, 3, 1, 0, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 1, 3, 1, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 1, 1, 3, 3, 3, 3, 3, 1, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 0, 0, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 1, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var genericDragonSymbol = [][]int{
	{0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0},
	{0, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 0},
	{1, 3, 3, 3, 1, 0, 1, 1, 1, 0, 0, 1, 3, 3, 3, 1},
	{0, 1, 3, 3, 3, 1, 3, 3, 3, 1, 1, 3, 3, 3, 1, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0},
	{0, 0, 0, 1, 3, 3, 2, 3, 2, 3, 3, 3, 1, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 0, 0, 1, 1, 3, 3, 3, 3, 3, 1, 1, 0, 0, 0, 0},
	{0, 0, 1, 3, 0, 1, 1, 0, 1, 1, 0, 3, 1, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var genericTreeSymbol = [][]int{
	{0, 0, 0, 0, 0, 0, 1, 3, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var genericForestSymbol = [][]int{
	{0, 0, 1, 3, 0, 0, 1, 3, 0, 0, 1, 3, 0, 0, 0, 0},
	{0, 1, 3, 3, 1, 1, 3, 3, 1, 1, 3, 3, 1, 0, 0, 0},
	{1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0},
	{1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0},
	{0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var genericCampSymbol = [][]int{
	{0, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 3, 3, 3, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var genericCaveSymbol = [][]int{
	{0, 0, 0, 1, 1, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 1, 0, 0, 1, 3, 3, 1, 0, 0, 0, 0},
	{0, 1, 3, 3, 3, 3, 1, 1, 3, 3, 3, 3, 1, 0, 0, 0},
	{1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0},
	{1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0},
	{0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var genericDungeonSymbol = [][]int{
	{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0},
	{1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0},
	{1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0},
	{1, 3, 3, 1, 1, 3, 3, 3, 3, 1, 1, 3, 3, 1, 0, 0},
	{1, 3, 3, 1, 0, 1, 3, 3, 1, 0, 1, 3, 3, 1, 0, 0},
	{1, 3, 3, 1, 0, 0, 1, 1, 0, 0, 1, 3, 3, 1, 0, 0},
	{0, 1, 3, 1, 0, 0, 0, 0, 0, 0, 1, 3, 1, 0, 0, 0},
	{0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var genericMountainSymbol = [][]int{
	{0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0},
	{1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0},
	{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var genericShieldSymbol = [][]int{
	{0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 1, 3, 3, 1, 3, 3, 3, 3, 1, 3, 3, 1, 0, 0, 0},
	{0, 1, 3, 1, 3, 3, 3, 3, 3, 3, 1, 3, 1, 0, 0, 0},
	{0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 3, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var genericPotionSymbol = [][]int{
	{0, 0, 0, 0, 0, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 1, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 3, 1, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 1, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 1, 3, 1, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 1, 3, 3, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 1, 3, 3, 3, 3, 3, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

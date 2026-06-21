// Phase 1 prototype:驗證 CJK 注入點。
//
// 模擬 kazzmir/master-of-magic 引擎 lib/font/font.go 的繪字機制:
// 把每個 rune 預先 rasterize 成一張 ebiten.Image 並快取 (對應引擎的 GlyphImages),
// 繪字時逐 rune 取出 glyph image、依字寬前進 x (對應 doPrint)。
//
// 差別只在 glyph 來源:引擎原本用 fonts.lbx 的 ASCII 點陣 (int(c)-32),
// 這裡證明「碼點 >= 0x80 改用 TTF rasterize」這條分流可行 —— 即 Phase 1 要打進 font.go 的注入。
//
// 跑法 (headless 截圖):
//   xvfb-run -a go run . -font /path/uming.ttc -out /out/cjk-hello.png
package main

import (
	"flag"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

const (
	canvasW  = 420
	canvasH  = 150
	cjkSize  = 24 // CJK glyph 邏輯尺寸 (px);對應「拉高畫布不縮字」鐵則
	cellH    = 30
	baseline = 23
)

// glyphCache 對應引擎 Font.GlyphImages:rune -> 預渲染 glyph image。
type glyphCache struct {
	face  font.Face
	cache map[rune]*ebiten.Image
}

func newGlyphCache(face font.Face) *glyphCache {
	return &glyphCache{face: face, cache: make(map[rune]*ebiten.Image)}
}

// glyphFor 對應引擎 getGlyphImage:第一次 rasterize,之後走快取。
func (g *glyphCache) glyphFor(r rune) (*ebiten.Image, int) {
	if img, ok := g.cache[r]; ok {
		_, adv, _ := g.face.GlyphBounds(r)
		return img, adv.Round()
	}
	adv, _ := g.face.GlyphAdvance(r)
	w := adv.Round()
	if w <= 0 {
		w = cjkSize
	}
	rgba := image.NewRGBA(image.Rect(0, 0, w, cellH))
	d := &font.Drawer{
		Dst:  rgba,
		Src:  image.NewUniform(color.White),
		Face: g.face,
		Dot:  fixed.P(0, baseline),
	}
	d.DrawString(string(r))
	img := ebiten.NewImageFromImage(rgba)
	g.cache[r] = img
	return img, w
}

// drawText 對應引擎 doPrint 的核心迴圈:逐 rune 取 glyph、blit、前進 x。
func (g *glyphCache) drawText(dst *ebiten.Image, x, y float64, scale float64, text string) {
	useX := x
	for _, r := range text {
		if r == ' ' {
			useX += float64(cjkSize/2) * scale
			continue
		}
		img, adv := g.glyphFor(r)
		var op ebiten.DrawImageOptions
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(useX, y)
		dst.DrawImage(img, &op)
		useX += float64(adv+1) * scale
	}
}

type Game struct {
	gc     *glyphCache
	canvas *ebiten.Image
	tick   int
	outPNG string
}

func (game *Game) render(dst *ebiten.Image) {
	dst.Fill(color.RGBA{R: 24, G: 18, B: 48, A: 255}) // 深紫底,模擬魔法書背景
	// 1) 遊戲標題
	game.gc.drawText(dst, 12, 14, 2.0, "工作魔法大帝")
	// 2) 第一批 item 譯文 (取自 docs/strings/item-powers.tsv)
	game.gc.drawText(dst, 12, 70, 1.0, "烈焰  吸血  神聖復仇者")
	game.gc.drawText(dst, 12, 104, 1.0, "+3 攻擊   魔法免疫   隱形")
}

func (game *Game) Update() error {
	game.tick++
	// 第 3 個 frame 時,從 offscreen canvas 讀回 pixel 存 PNG,然後結束。
	if game.tick == 3 && game.outPNG != "" {
		game.render(game.canvas)
		pix := make([]byte, canvasW*canvasH*4)
		game.canvas.ReadPixels(pix)
		out := image.NewRGBA(image.Rect(0, 0, canvasW, canvasH))
		copy(out.Pix, pix)
		f, err := os.Create(game.outPNG)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := png.Encode(f, out); err != nil {
			return err
		}
		log.Printf("已輸出截圖:%s", game.outPNG)
		return ebiten.Termination
	}
	return nil
}

func (game *Game) Draw(screen *ebiten.Image) {
	game.render(screen)
}

func (game *Game) Layout(int, int) (int, int) { return canvasW, canvasH }

func main() {
	fontPath := flag.String("font", "/usr/share/fonts/truetype/arphic/uming.ttc", "CJK TTF/TTC 路徑")
	out := flag.String("out", "", "截圖輸出 PNG 路徑 (給定即 headless 截圖後結束)")
	flag.Parse()

	data, err := os.ReadFile(*fontPath)
	if err != nil {
		log.Fatalf("讀字型失敗:%v", err)
	}
	tt, err := opentype.Parse(data)
	if err != nil {
		// .ttc (TrueType collection,如 uming.ttc / NotoSansCJK) 走 ParseCollection 取第 0 個 font
		coll, cerr := opentype.ParseCollection(data)
		if cerr != nil {
			log.Fatalf("解析字型失敗:%v / collection:%v", err, cerr)
		}
		tt, err = coll.Font(0)
		if err != nil {
			log.Fatalf("取 collection font 0 失敗:%v", err)
		}
	}
	face, err := opentype.NewFace(tt, &opentype.FaceOptions{Size: cjkSize, DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		log.Fatalf("建立 face 失敗:%v", err)
	}

	game := &Game{
		gc:     newGlyphCache(face),
		canvas: ebiten.NewImage(canvasW, canvasH),
		outPNG: *out,
	}
	ebiten.SetWindowSize(canvasW*3, canvasH*3)
	ebiten.SetWindowTitle("工作魔法大帝 — CJK prototype")
	if err := ebiten.RunGame(game); err != nil && err != ebiten.Termination {
		log.Fatal(err)
	}
}

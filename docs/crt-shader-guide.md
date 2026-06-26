# CRT Shader 從零實作教學(Ebiten / Kage)

把銳利的方形像素變回「老 CRT 螢幕」的觀感——掃描線、子像素光罩、螢幕曲面、暗角、螢光暈。
這份文件從第一性原理拆解每個效果**為什麼這樣寫**,附可運行的 Kage 程式碼,最後示範怎麼接進本專案
(沿用 [`docs/dos-vs-remake-ui.md`](dos-vs-remake-ui.md) 的 DOS 長寬比那條離屏渲染路徑)。

> 適合想學 shader coding 的人:每段先講「螢幕上實際發生什麼物理」,再看程式怎麼模擬。

---

## 0. 先搞懂:fragment shader 到底在算什麼

GPU 畫一張圖時,會對**輸出畫面的每一個像素**各跑一次一個小程式,問它:「這個位置該是什麼顏色?」
這個小程式就是 **fragment(片段)shader**。它的輸入是「我是哪個像素」(座標),輸出是「我這格的 RGBA」。

CRT shader 是一種 **post-process(後製)**:遊戲已經畫好一張完整畫面(一張貼圖 texture),
我們再對它整張跑一次 fragment shader,**讀原畫面的顏色、加工後輸出**。所以它不碰遊戲邏輯,只改「最後呈現」。

```
遊戲畫面(離屏 texture)  ──► fragment shader(逐像素加工)──► 螢幕
   每像素原始 RGB              掃描線×光罩×曲面×暈光×暗角         CRT 觀感
```

關鍵詞:
- **UV 座標**:把畫面位置正規化成 0~1 的 `(u, v)`,`(0,0)` 左上、`(1,1)` 右下。shader 多用 UV 思考,因為和解析度無關。
- **取樣(sample)**:`tex(uv)` 從原畫面 UV 位置讀一個顏色。
- **每像素獨立**:shader 不能「看旁邊算好沒」,只能讀輸入 texture——所以效果都寫成「這個 UV 該輸出什麼」。

---

## 1. Ebiten 的 shader 語言:Kage

Ebiten 用自家的 **Kage** 語言寫 shader(語法像 Go,編譯成各平台 GPU 程式)。一個最小 fragment shader:

```go
//kage:unit pixels
package main

// Fragment 是進入點。GPU 對每個輸出像素呼叫一次。
//   dstPos: 此像素在「目的地(螢幕)」的座標(pixels,vec4 但用 .xy)
//   srcPos: 對應到「來源 texture(遊戲畫面)」的座標(pixels)
//   color:  頂點顏色(這裡用不到)
// 回傳:這個像素最終的 RGBA(premultiplied alpha)
func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    return imageSrc0At(srcPos)   // 原樣輸出來源畫面(還沒做任何效果)
}
```

幾個內建(Ebiten 提供):
- `imageSrc0At(pos vec2) vec4` — 從來源 texture 0 在 `pos`(pixels)取一個顏色。
- `imageSrc0Size() vec2` — 來源 texture 的尺寸(pixels)。用來把 pixels 換成 UV:`uv := srcPos / imageSrc0Size()`。
- 型別:`vec2/vec3/vec4`、`mat2..4`;數學:`sin cos pow sqrt abs floor fract mix clamp smoothstep min max dot length`…(和 GLSL 幾乎一樣)。
- `//kage:unit pixels` 指定座標用 pixels(另一個選項是 texels);本文用 pixels。

> **premultiplied alpha 注意**:Ebiten 的顏色是 premultiplied(RGB 已乘上 A)。對不透明畫面(A=1)做乘法調亮暗沒問題;
> 要做 alpha 相關運算時要留意。CRT 後製畫的是整張不透明畫面,通常 A 維持 1。

---

## 2. 逐效果拆解(每個都先講物理,再看程式)

以下每段都假設已算好 `uv := srcPos / imageSrc0Size()`(0~1)、`src := imageSrc0At(srcPos).rgb`(原始顏色)。

### 2.1 掃描線(scanlines)— 最有感的一招

**物理**:CRT 用電子束**一條一條水平掃描線**畫面,線與線之間有一點點暗縫;低解析度(320×200)時這些暗縫很明顯。

**第一性原理**:沿著畫面的 **Y 方向**,做一個明暗週期變化——亮帶(掃描線中心)+ 暗縫(線之間)。
用 `sin` 產生週期;週期數 = 要模擬的掃描線條數(通常 = 來源高度,或玩家設定)。

```go
// lines = 掃描線條數(例如來源高度,或固定 240)
// uv.y * lines 走一圈是一條線;乘 2π 給 sin
scan := sin(uv.y * lines * 2.0 * 3.14159265)   // -1..1
// 把 -1..1 映成「暗縫 0.x ~ 亮帶 1.0」的調暗係數
darken := mix(scanlineDepth, 1.0, scan*0.5+0.5) // scanlineDepth 例如 0.6(縫多暗)
rgb := src * darken
```

`mix(a, b, t)` = `a*(1-t) + b*t` 線性內插。這裡 `t = scan*0.5+0.5`(把 -1..1 變 0..1),
所以亮帶(t=1)係數 1.0、暗縫(t=0)係數 `scanlineDepth`。`scanlineDepth` 越小縫越黑、CRT 味越重。

> 進階:掃描線會讓整體變暗,通常要再乘一個 `brightnessBoost`(例如 1.3)補回亮度。

### 2.2 子像素光罩(aperture / shadow mask)

**物理**:CRT 的彩色畫面是靠**紅綠藍三種螢光點**交錯排列(aperture grille 是直條、shadow mask 是點)。
近看會看到 RGB 直條紋,這也是「像素感」的來源之一。

**第一性原理**:沿 **X 方向**每 3 個輸出像素一組,分別「強化 R、G、B」其中一色、壓暗另兩色。
用 `dstPos.x`(目的地像素)取模 3 決定這格屬於哪個子像素。

```go
// 用目的地像素的 X 決定 RGB 直條(每 3 px 一組)
phase := mod(dstPos.x, 3.0)        // 0,1,2 循環
mask := vec3(0.0)
if phase < 1.0 {      mask = vec3(1.0, maskDark, maskDark) } // 這格偏紅
else if phase < 2.0 { mask = vec3(maskDark, 1.0, maskDark) } // 偏綠
else {                mask = vec3(maskDark, maskDark, 1.0) } // 偏藍
rgb := src * mask     // maskDark 例如 0.7;越小條紋越強
```

> 光罩會明顯壓暗整體(平均只剩約 (1+2·maskDark)/3),同樣要配亮度補償。子像素效果只有在「輸出解析度夠高」
> (每個來源像素被放大成好幾個輸出像素)時才看得清楚,否則會變成摩爾紋(見 §4 踩雷)。

### 2.3 螢幕曲面 / 桶形畸變(barrel distortion)

**物理**:老 CRT 玻璃是**外凸的弧面**,畫面邊緣會往外膨脹一點。

**第一性原理**:把 UV 從畫面中心往外**依距離平方推開**(離中心越遠、推越多),取樣時用畸變後的 UV。
這是「取樣座標的扭曲」,不是顏色的扭曲。

```go
func barrel(uv vec2, amount float) vec2 {
    cc := uv - 0.5            // 以中心為原點(-0.5..0.5)
    d := dot(cc, cc)          // 距中心的平方(0..0.5)
    return uv + cc * d * amount  // 依 d 往外推;amount 例如 0.15
}
// 用法:distortedUV := barrel(uv, curvature); 再用它去取樣
```

畸變後 UV 可能落到 0~1 之外(畫面外)——那裡應該是黑邊(CRT 玻璃外框),記得判斷後填黑(見 §3 組合)。

### 2.4 暗角(vignette)

**物理**:CRT 邊角受電子束打不到/玻璃遮擋,亮度較低。

**第一性原理**:離中心越遠越暗,用距離做一個平滑衰減。

```go
vig := smoothstep(0.0, vignetteSize, 0.5 - length(uv-0.5)) // 中心 1、邊角趨近 0
rgb *= mix(1.0, vig, vignetteStrength)
```

`smoothstep(edge0, edge1, x)` 在 edge0~edge1 間做 S 形平滑 0→1,比線性自然。

### 2.5 螢光暈 / bloom(亮處外溢)

**物理**:CRT 螢光在亮處會「暈開」一點,亮的東西邊緣有柔光。

**第一性原理**:取樣**周圍幾點**的顏色取最大/平均,疊一點回來(等於一個便宜的模糊)。
真正的 bloom 要先抽亮部再多階模糊,後製 shader 裡常用簡化版:鄰近幾點平均當 glow。

```go
texel := 1.0 / imageSrc0Size()   // 一個來源像素在 UV 的大小
glow := src
glow += imageSrc0At((uv + vec2( texel.x, 0)) * imageSrc0Size()).rgb
glow += imageSrc0At((uv + vec2(-texel.x, 0)) * imageSrc0Size()).rgb
glow += imageSrc0At((uv + vec2(0,  texel.y)) * imageSrc0Size()).rgb
glow += imageSrc0At((uv + vec2(0, -texel.y)) * imageSrc0Size()).rgb
glow /= 5.0
rgb = mix(rgb, max(rgb, glow), bloomStrength)  // 只讓亮處外溢
```

> bloom 最吃效能(每像素多次取樣)。要更好品質得分成「抽亮 → 降採樣模糊 → 疊回」多 pass,本文先給單 pass 簡化版。

### 2.6 gamma / 對比(收尾)

CRT 的亮度響應不是線性(gamma ≈ 2.2)。所有效果做完後,常套一個 gamma 微調讓暗部更沉、對比更像 CRT:

```go
rgb = pow(rgb, vec3(1.0/crtGamma))   // crtGamma 例如 1.1~1.3
```

---

## 3. 組合成一支完整 CRT shader

把上面拼起來,順序通常是:**曲面畸變(改取樣座標)→ 取樣 + bloom → 掃描線 → 光罩 → 暗角 → gamma → 亮度補償**。

```go
//kage:unit pixels
package main

// 可調參數(實作時可從 Go 端用 uniform 傳進來,這裡先寫常數方便讀)
const scanDepth   = 0.65   // 掃描線暗縫深度(越小越暗)
const maskDark    = 0.75   // 子像素光罩暗色(越小條紋越強)
const curvature   = 0.10   // 螢幕曲面量(0=平面)
const vignette    = 0.35   // 暗角強度
const bloom       = 0.25   // 螢光暈強度
const crtGamma    = 1.20   // gamma
const brighten    = 1.35   // 亮度補償(補回掃描線+光罩壓暗)
const Pi          = 3.14159265

func barrel(uv vec2, amount float) vec2 {
    cc := uv - 0.5
    return uv + cc * dot(cc, cc) * amount
}

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    size := imageSrc0Size()
    uv := srcPos / size

    // 1) 螢幕曲面:扭曲取樣座標
    cuv := barrel(uv, curvature)

    // 畫面外(玻璃框外)→ 黑邊
    if cuv.x < 0.0 || cuv.x > 1.0 || cuv.y < 0.0 || cuv.y > 1.0 {
        return vec4(0.0, 0.0, 0.0, 1.0)
    }

    // 2) 取樣 + 便宜 bloom(鄰近 4 點平均)
    texel := 1.0 / size
    src := imageSrc0At(cuv * size).rgb
    glow := src
    glow += imageSrc0At((cuv + vec2( texel.x, 0)) * size).rgb
    glow += imageSrc0At((cuv + vec2(-texel.x, 0)) * size).rgb
    glow += imageSrc0At((cuv + vec2(0,  texel.y)) * size).rgb
    glow += imageSrc0At((cuv + vec2(0, -texel.y)) * size).rgb
    glow /= 5.0
    rgb := mix(src, max(src, glow), bloom)

    // 3) 掃描線(沿 Y;線數 = 來源高度)
    scan := sin(cuv.y * size.y * Pi) * 0.5 + 0.5
    rgb *= mix(scanDepth, 1.0, scan)

    // 4) 子像素光罩(沿目的地 X,每 3 px 一組)
    phase := mod(dstPos.x, 3.0)
    mask := vec3(maskDark, maskDark, maskDark)
    if phase < 1.0 {
        mask.r = 1.0
    } else if phase < 2.0 {
        mask.g = 1.0
    } else {
        mask.b = 1.0
    }
    rgb *= mask

    // 5) 暗角
    vig := smoothstep(0.0, 0.7, 0.75 - length(uv-0.5))
    rgb *= mix(1.0, vig, vignette)

    // 6) 亮度補償 + gamma
    rgb *= brighten
    rgb = pow(clamp(rgb, 0.0, 1.0), vec3(1.0/crtGamma))

    return vec4(rgb, 1.0)
}
```

> 想讓參數可在遊戲內即時調,把那些 `const` 改成 **uniform**:Kage 端宣告 `var Scanline float`(在 Fragment 外),
> Go 端 `op.Uniforms = map[string]any{"Scanline": 0.65}` 傳入。

---

## 4. 怎麼接進本引擎(沿用 DOS 長寬比那條離屏路徑)

本專案已經有現成的後製 hook:`main.go` 的 `Draw` 在 DOS 長寬比模式下,把遊戲畫到一張**未拉伸的離屏 buffer**
再貼到螢幕(見 [`docs/dos-vs-remake-ui.md`](dos-vs-remake-ui.md))。CRT shader 就插在「離屏 → 螢幕」這一步:

```go
// 1) 載入 shader(啟動時一次)
//go:embed crt.kage
var crtSrc []byte
var crtShader *ebiten.Shader
// crtShader, _ = ebiten.NewShader(crtSrc)

func (game *MagicGame) Draw(screen *ebiten.Image) {
    // ...(照舊)把遊戲畫到 dosOffscreen(960×600 或拉伸後的尺寸)...

    if data.Settings.CrtShader && crtShader != nil {
        op := &ebiten.DrawRectShaderOptions{}
        op.Images[0] = dosOffscreen                 // 來源 = 遊戲畫面
        // 若同時要 DOS 長寬比,GeoM 仍可先做 Y×1.2,再交給 shader
        op.GeoM.Scale(1, aspectStretch)
        op.Uniforms = map[string]any{               // 可選:即時調參
            "Scanline": float32(0.65),
        }
        w, h := dosOffscreen.Bounds().Dx(), dosOffscreen.Bounds().Dy()
        screen.DrawRectShader(w, h, crtShader, op)  // 用 shader 把來源畫到螢幕
    } else {
        // 照舊直接貼
        screen.DrawImage(dosOffscreen, normalOptions)
    }
}
```

要點:
- 用 `screen.DrawRectShader(w, h, shader, op)`,`op.Images[0]` 是來源 texture(就是 `imageSrc0At` 讀的那張)。
- shader 內看到的 `imageSrc0Size()` 是來源(`dosOffscreen`)尺寸——掃描線數、texel 都依它算。
- 跟 DOS 長寬比**正交**:長寬比是 `GeoM` 的幾何拉伸,CRT 是顏色加工,兩個可同時開(先拉伸再 shader)。
- 設定面:加 `data.Settings.CrtShader bool`(預設關)+ 設定畫面一個切換鈕,跟現有 17 個 toggle 同模式。
  滑鼠座標不受影響(shader 只改顏色不改 Layout),所以**不用**像 DOS 長寬比那樣校正輸入。

---

## 5. 調參、效能、踩雷

- **掃描線/光罩需要足夠的放大倍率**:320×200 放大到 960×600 是 3×。掃描線(來源高度 200 條)在 600px 高上每條約 3px——
  剛好夠。若輸出解析度不夠(每來源像素 < 2~3 輸出像素),掃描線/光罩會和像素網格打架產生**摩爾紋**。
  解法:確保整數倍放大、或 shader 內用 `cuv.y * outputHeight` 而非來源高度來定線數。
- **亮度**:掃描線 + 光罩會把畫面壓暗到原本的 ~40~60%,務必用 `brighten` 補回,否則整個畫面灰暗。
- **效能**:bloom 的多次取樣最貴。先用單 pass 簡化版;要更好品質再拆「抽亮→降採樣模糊→疊回」多 pass(多張離屏)。
  行動裝置(Android)上 CRT shader 要實測幀率。
- **premultiplied alpha**:對不透明全螢幕畫面做乘法 OK;若來源有透明區要小心。
- **不要過頭**:曲面 + 暗角 + 強掃描線疊起來很容易「太重」變成濾鏡感。老玩家要的是「依稀記得的 CRT 味」,
  建議參數從輕到重做幾組預設(輕/中/重)讓玩家選。
- **和 DOS 長寬比的關係**:長寬比(Y×1.2)是「對的幾何」、CRT shader 是「對的質地」——兩者互補。
  最像 DOS 的組合 = 長寬比開 + 輕度 CRT shader。

---

## 6. 延伸

- 真實感更高的參考實作:**CRT-Geom**、**CRT-Royale**(libretro/RetroArch 的經典 CRT shader,GLSL,可讀演算法移植到 Kage)。
- 曲面 + 邊框玻璃反光、aperture grille vs shadow mask vs slot mask 的差別、phosphor persistence(殘影)都可再加。
- Ebiten Kage 官方文件與範例:`examples/shader`。

## 來源 / 延伸閱讀

- Ebiten Kage 語言與 `DrawRectShader`:<https://ebitengine.org/en/documents/shader.html>
- libretro CRT shader(CRT-Geom / CRT-Royale)演算法:<https://github.com/libretro/glsl-shaders/tree/master/crt>
- 本專案的後製 hook(離屏渲染路徑):[`docs/dos-vs-remake-ui.md`](dos-vs-remake-ui.md)、`game/magic/main.go` 的 `Draw`。

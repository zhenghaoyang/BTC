// Lissajous generates GIF animations of random Lissajous figures.

// 修改Lissajous程序，修改其调色板来生成更丰富的颜色
// ，然后修改SetColorIndex的第三个参数，看看显示结果吧。
package main

import (
	"image"
	"image/color"
	"image/gif"
	"io"
	"math"
	"math/rand"
	"os"
	"time"
)

//练习 1.5： 修改前面的Lissajous程序里的调色板，由黑色改为绿色。
// 我们可以用color.RGBA{0xRR, 0xGG, 0xBB, 0xff}来得到#RRGGBB这个色值，
// 三个十六进制的字符串分别代表红、绿、蓝像素。

// var palette = []color.Color{color.White, color.Black}
var palettes = []color.Color{color.White, color.RGBA{0x00, 0x00, 0xbb, 0xff}, color.Black}

const (
	whiteIndex = 0 // first color in palette
	blackIndex = 1 // next color in palette
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	lissajous(os.Stdout)
}

//常量是指在程序编译后运行时始终都不会变化的值，比如圈数、帧数、延迟值。
// 常量声明和变量声明一般都会出现在包级别，所以这些常量在整个包中都是可以共享的，
// 或者你也可以把常量声明定义在函数体内部，那么这种常量就只能在函数体内用。
func lissajous(out io.Writer) {
	const (
		cycles  = 5     // number of complete x oscillator revolutions
		res     = 0.001 // angular resolution
		size    = 100   // image canvas covers [-size..+size]
		nframes = 64    // number of animation frames
		delay   = 8     // delay between frames in 10ms units
	)

	freq := rand.Float64() * 3.0 // relative frequency of y oscillator
	anim := gif.GIF{LoopCount: nframes}
	phase := 0.0 // phase difference
	// lissajous函数内部有两层嵌套的for循环。
	// 外层循环会循环64次，每一次都会生成一个单独的动画帧。
	for i := 0; i < nframes; i++ {
		//它生成了一个包含两种颜色的201*201大小的图片，白色和黑色。
		rect := image.Rect(0, 0, 2*size+1, 2*size+1)
		// 所有像素点都会被默认设置为其零值（也就是调色板palette里的第0个值）
		// 这里我们设置的是白色。每次外层循环都会生成一张新图片，并将一些像素设置为黑色。
		img := image.NewPaletted(rect, palettes)

		for t := 0.0; t < cycles*2*math.Pi; t += res {
			// /内层循环设置两个偏振值。x轴偏振使用sin函数。y轴偏振也是正弦波，
			// 但其相对x轴的偏振是一个0-3的随机值，
			// 初始偏振值是一个零值，随着动画的每一帧逐渐增加。
			// 循环会一直跑到x轴完成五次完整的循环。每一步它都会调用SetColorIndex来为(x, y)点来染黑色。
			x := math.Sin(t)
			y := math.Sin(t*freq + phase)
			img.SetColorIndex(size+int(x*size+0.2), size+int(y*size+0.2),
				2)
		}
		phase += 0.1
		anim.Delay = append(anim.Delay, delay)
		anim.Image = append(anim.Image, img)
	}
	gif.EncodeAll(out, &anim)
}

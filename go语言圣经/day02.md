### 第一章 入门

bufio.Scanner、ioutil.ReadFile和ioutil.WriteFile都使用*os.File的Read和Write方法，但是，大多数程序员很少需要直接调用那些低级（lower-level）函数。高级（higher-level）函数，像bufio和io/ioutil包中所提供的那些，用起来要容易点。

```
var palette = []color.Color{color.White, color.Black}
anim := gif.GIF{LoopCount: nframes} 

```
[]color.Color{...}和gif.GIF{...}这两个表达式就是我们说的复合声明（4.2和4.4.1节有说明）。这是实例化Go语言里的复合类型的一种写法。这里的前者生成的是一个slice切片，后者生成的是一个struct结构体。




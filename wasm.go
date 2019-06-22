package main

import (
	"syscall/js"
)

func main() {
	width := js.Global().Get("innerWidth").Int()
	height := js.Global().Get("innerHeight").Int()
	doc := js.Global().Get("document")
	canvasEl := doc.Call("getElementById", "mycanvas")
	canvasEl.Call("setAttribute", "width", width)
	canvasEl.Call("setAttribute", "height", height)
	ctx := canvasEl.Call("getContext", "2d")
	step := 10

	done := make(chan struct{}, 0)

	// Frame rendering function.  Just draws grid lines.
	var renderFrame js.Func
	renderFrame = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		ctx.Set("strokeStyle", "darkgrey")
		for z := 0; z < 100; z++ {
			for y := 0; y < 200; y += step {
				// Vertical dashed lines
				ctx.Call("beginPath")
				ctx.Call("moveTo", y+step, 0)
				ctx.Call("lineTo", y+step, 200)
				ctx.Call("stroke")

				// Horizontal dashed lines
				ctx.Call("beginPath")
				ctx.Call("moveTo", 0, y+step)
				ctx.Call("lineTo", 200, y+step)
				ctx.Call("stroke")
			}
		}
		js.Global().Call("requestAnimationFrame", renderFrame)
		return nil
	})

	// Start the render loop running
	js.Global().Call("requestAnimationFrame", renderFrame)
	<-done
}
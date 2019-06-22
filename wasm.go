package main

import (
	"math"
	"syscall/js"
)

type matrix []float64

type Point struct {
	Label      string
	LabelAlign string
	X          float64
	Y          float64
	Z          float64
}

type Edge []int
type Surface []int

type Object struct {
	C         string // Colour of the object
	P         []Point
	E         []Edge    // List of points to connect by edges
	S         []Surface // List of points to connect in order, to create a surface
	DrawOrder int       // Draw order for the object
	Name      string
}

var (
	// The empty world space
	worldSpace []Object

	// The point objects
	axes = Object{
		C:         "grey",
		DrawOrder: 0,
		Name:      "axes",
		P: []Point{
			{X: -0.1, Y: 0.1, Z: 0.0},
			{X: -0.1, Y: 10, Z: 0.0},
			{X: 0.1, Y: 10, Z: 0.0},
			{X: 0.1, Y: 0.1, Z: 0.0},
			{X: 10, Y: 0.1, Z: 0.0},
			{X: 10, Y: -0.1, Z: 0.0},
			{X: 0.1, Y: -0.1, Z: 0.0},
			{X: 0.1, Y: -10, Z: 0.0},
			{X: -0.1, Y: -10, Z: 0.0},
			{X: -0.1, Y: -0.1, Z: 0.0},
			{X: -10, Y: -0.1, Z: 0.0},
			{X: -10, Y: 0.1, Z: 0.0},
			{X: 10, Y: -1.0, Z: 0.0, Label: "X", LabelAlign: "center"},
			{X: -10, Y: -1.0, Z: 0.0, Label: "-X", LabelAlign: "center"},
			{X: 0.0, Y: 10.5, Z: 0.0, Label: "Y", LabelAlign: "center"},
			{X: 0.0, Y: -11, Z: 0.0, Label: "-Y", LabelAlign: "center"},
		},
		E: []Edge{
			{0, 1},
			{1, 2},
			{2, 3},
			{3, 4},
			{4, 5},
			{5, 6},
			{6, 7},
			{7, 8},
			{8, 9},
			{9, 10},
			{10, 11},
			{11, 0},
		},
		S: []Surface{
			{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		},
	}

	// The 4x4 identity matrix
	identityMatrix = matrix{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}

	// Initialise the transform matrix with the identity matrix
	transformMatrix = identityMatrix

	canvasEl, ctx, doc js.Value
	graphWidth         float64
	graphHeight        float64
	height, width      int
	renderFrame        js.Func
)

func main() {
	// The actual main function, run once at start
	width := js.Global().Get("innerWidth").Int()
	height := js.Global().Get("innerHeight").Int()
	doc = js.Global().Get("document")
	canvasEl = doc.Call("getElementById", "mycanvas")
	canvasEl.Call("setAttribute", "width", width)
	canvasEl.Call("setAttribute", "height", height)
	canvasEl.Set("tabIndex", 0) // Not sure if this is needed
	ctx = canvasEl.Call("getContext", "2d")

	// Add the X/Y axes object to the world space
	worldSpace = append(worldSpace, importObject(axes, 0.0, 0.0, 0.0))

	// Set up an initial movement transformation, so we can see the renderFrame working
	transformMatrix = rotateAroundX(transformMatrix, -25/float64(12))
	transformMatrix = rotateAroundY(transformMatrix, -25/float64(12))

	renderFrame = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		realRenderFrame()
		return nil
	})

	// Start running
	js.Global().Call("requestAnimationFrame", renderFrame)
}

// Renders one frame of the animation
func realRenderFrame() {
	// Handle window resizing
	curBodyW := js.Global().Get("innerWidth").Int()
	curBodyH := js.Global().Get("innerHeight").Int()
	if curBodyW != width || curBodyH != height {
		width, height = curBodyW, curBodyH
		canvasEl.Set("width", width)
		canvasEl.Set("height", height)
	}

	// Setup useful variables
	border := float64(2)
	gap := float64(3)
	left := border + gap
	top := border + gap
	graphWidth = float64(width) * 0.75
	graphHeight = float64(height) - 1
	centerX := graphWidth / 2
	centerY := graphHeight / 2

	// Clear the background
	ctx.Set("fillStyle", "white")
	ctx.Call("fillRect", 0, 0, width, height)

	// Draw grid lines
	step := math.Min(float64(width), float64(height)) / float64(30)
	ctx.Set("strokeStyle", "rgb(220, 220, 220)")
	for i := left; i < graphWidth-step; i += step {
		// Vertical dashed lines
		ctx.Call("beginPath")
		ctx.Call("moveTo", i+step, top)
		ctx.Call("lineTo", i+step, graphHeight)
		ctx.Call("stroke")
	}
	for i := top; i < graphHeight-step; i += step {
		// Horizontal dashed lines
		ctx.Call("beginPath")
		ctx.Call("moveTo", left, i+step)
		ctx.Call("lineTo", graphWidth-border, i+step)
		ctx.Call("stroke")
	}

	// Draw the axes
	var pointX, pointY float64
	ctx.Set("strokeStyle", "black")
	ctx.Set("lineWidth", "1")
	for _, o := range worldSpace {

		// Draw the surfaces
		ctx.Set("fillStyle", o.C)
		for _, l := range o.S {
			for m, n := range l {
				pointX = o.P[n].X
				pointY = o.P[n].Y
				if m == 0 {
					ctx.Call("beginPath")
					ctx.Call("moveTo", centerX+(pointX*step), centerY+((pointY*step)*-1))
				} else {
					ctx.Call("lineTo", centerX+(pointX*step), centerY+((pointY*step)*-1))
				}
			}
			ctx.Call("closePath")
			ctx.Call("fill")
		}

		// Draw the edges
		var point1X, point1Y, point2X, point2Y float64
		for _, l := range o.E {
			point1X = o.P[l[0]].X
			point1Y = o.P[l[0]].Y
			point2X = o.P[l[1]].X
			point2Y = o.P[l[1]].Y
			ctx.Call("beginPath")
			ctx.Call("moveTo", centerX+(point1X*step), centerY+((point1Y*step)*-1))
			ctx.Call("lineTo", centerX+(point2X*step), centerY+((point2Y*step)*-1))
			ctx.Call("stroke")
		}

		// Draw any point labels
		ctx.Set("fillStyle", "black")
		ctx.Set("font", "bold 16px serif")
		var px, py float64
		for _, l := range o.P {
			if l.Label != "" {
				ctx.Set("textAlign", l.LabelAlign)
				px = centerX + (l.X * step)
				py = centerY + ((l.Y * step) * -1)
				ctx.Call("fillText", l.Label, px, py)
			}
		}
	}
	ctx.Set("lineWidth", "2")

	js.Global().Call("requestAnimationFrame", renderFrame)
}

// Returns an object whose points have been transformed into 3D world space XYZ co-ordinates.  Also assigns a number
// to each point
func importObject(ob Object, x float64, y float64, z float64) (translatedObject Object) {
	// X and Y translation matrix.  Translates the objects into the world space at the given X and Y co-ordinates
	translateMatrix := matrix{
		1, 0, 0, x,
		0, 1, 0, y,
		0, 0, 1, z,
		0, 0, 0, 1,
	}

	// Translate the points
	var pt Point
	for _, j := range ob.P {
		pt = Point{
			Label:      j.Label,
			LabelAlign: j.LabelAlign,
			X:          (translateMatrix[0] * j.X) + (translateMatrix[1] * j.Y) + (translateMatrix[2] * j.Z) + (translateMatrix[3] * 1),   // 1st col, top
			Y:          (translateMatrix[4] * j.X) + (translateMatrix[5] * j.Y) + (translateMatrix[6] * j.Z) + (translateMatrix[7] * 1),   // 1st col, upper middle
			Z:          (translateMatrix[8] * j.X) + (translateMatrix[9] * j.Y) + (translateMatrix[10] * j.Z) + (translateMatrix[11] * 1), // 1st col, lower middle
		}
		translatedObject.P = append(translatedObject.P, pt)
	}

	// Copy the remaining object info across
	translatedObject.C = ob.C
	translatedObject.Name = ob.Name
	translatedObject.DrawOrder = ob.DrawOrder
	for _, j := range ob.E {
		translatedObject.E = append(translatedObject.E, j)
	}
	for _, j := range ob.S {
		translatedObject.S = append(translatedObject.S, j)
	}

	return translatedObject
}

// Apply each transformation, one small part at a time (this gives the animation effect)
//go:export applyTransformation
func applyTransformation() {
	// If the queue # if greater than zero, there are still transforms to do
	for j, o := range worldSpace {
		var newPoints []Point

		// Transform each point of in the object
		for _, j := range o.P {
			newPoints = append(newPoints, transform(transformMatrix, j))
		}
		o.P = newPoints

		// Update the object in world space
		worldSpace[j] = o
	}
}

// Multiplies one matrix by another
func matrixMult(opMatrix matrix, m matrix) (resultMatrix matrix) {
	top0 := m[0]
	top1 := m[1]
	top2 := m[2]
	top3 := m[3]
	upperMid0 := m[4]
	upperMid1 := m[5]
	upperMid2 := m[6]
	upperMid3 := m[7]
	lowerMid0 := m[8]
	lowerMid1 := m[9]
	lowerMid2 := m[10]
	lowerMid3 := m[11]
	bot0 := m[12]
	bot1 := m[13]
	bot2 := m[14]
	bot3 := m[15]

	resultMatrix = matrix{
		(opMatrix[0] * top0) + (opMatrix[1] * upperMid0) + (opMatrix[2] * lowerMid0) + (opMatrix[3] * bot0), // 1st col, top
		(opMatrix[0] * top1) + (opMatrix[1] * upperMid1) + (opMatrix[2] * lowerMid1) + (opMatrix[3] * bot1), // 2nd col, top
		(opMatrix[0] * top2) + (opMatrix[1] * upperMid2) + (opMatrix[2] * lowerMid2) + (opMatrix[3] * bot2), // 3rd col, top
		(opMatrix[0] * top3) + (opMatrix[1] * upperMid3) + (opMatrix[2] * lowerMid3) + (opMatrix[3] * bot3), // 4th col, top

		(opMatrix[4] * top0) + (opMatrix[5] * upperMid0) + (opMatrix[6] * lowerMid0) + (opMatrix[7] * bot0), // 1st col, upper middle
		(opMatrix[4] * top1) + (opMatrix[5] * upperMid1) + (opMatrix[6] * lowerMid1) + (opMatrix[7] * bot1), // 2nd col, upper middle
		(opMatrix[4] * top2) + (opMatrix[5] * upperMid2) + (opMatrix[6] * lowerMid2) + (opMatrix[7] * bot2), // 3rd col, upper middle
		(opMatrix[4] * top3) + (opMatrix[5] * upperMid3) + (opMatrix[6] * lowerMid3) + (opMatrix[7] * bot3), // 4th col, upper middle

		(opMatrix[8] * top0) + (opMatrix[9] * upperMid0) + (opMatrix[10] * lowerMid0) + (opMatrix[11] * bot0), // 1st col, lower middle
		(opMatrix[8] * top1) + (opMatrix[9] * upperMid1) + (opMatrix[10] * lowerMid1) + (opMatrix[11] * bot1), // 2nd col, lower middle
		(opMatrix[8] * top2) + (opMatrix[9] * upperMid2) + (opMatrix[10] * lowerMid2) + (opMatrix[11] * bot2), // 3rd col, lower middle
		(opMatrix[8] * top3) + (opMatrix[9] * upperMid3) + (opMatrix[10] * lowerMid3) + (opMatrix[11] * bot3), // 4th col, lower middle

		(opMatrix[12] * top0) + (opMatrix[13] * upperMid0) + (opMatrix[14] * lowerMid0) + (opMatrix[15] * bot0), // 1st col, bottom
		(opMatrix[12] * top1) + (opMatrix[13] * upperMid1) + (opMatrix[14] * lowerMid1) + (opMatrix[15] * bot1), // 2nd col, bottom
		(opMatrix[12] * top2) + (opMatrix[13] * upperMid2) + (opMatrix[14] * lowerMid2) + (opMatrix[15] * bot2), // 3rd col, bottom
		(opMatrix[12] * top3) + (opMatrix[13] * upperMid3) + (opMatrix[14] * lowerMid3) + (opMatrix[15] * bot3), // 4th col, bottom
	}
	return resultMatrix
}

// Rotates a transformation matrix around the X axis by the given degrees
func rotateAroundX(m matrix, degrees float64) matrix {
	rad := (math.Pi / 180) * degrees // The Go math functions use radians, so we convert degrees to radians
	rotateXMatrix := matrix{
		1, 0, 0, 0,
		0, math.Cos(rad), -math.Sin(rad), 0,
		0, math.Sin(rad), math.Cos(rad), 0,
		0, 0, 0, 1,
	}
	return matrixMult(rotateXMatrix, m)
}

// Rotates a transformation matrix around the Y axis by the given degrees
func rotateAroundY(m matrix, degrees float64) matrix {
	rad := (math.Pi / 180) * degrees // The Go math functions use radians, so we convert degrees to radians
	rotateYMatrix := matrix{
		math.Cos(rad), 0, math.Sin(rad), 0,
		0, 1, 0, 0,
		-math.Sin(rad), 0, math.Cos(rad), 0,
		0, 0, 0, 1,
	}
	return matrixMult(rotateYMatrix, m)
}

// Transform the XYZ co-ordinates using the values from the transformation matrix
func transform(m matrix, p Point) (t Point) {
	top0 := m[0]
	top1 := m[1]
	top2 := m[2]
	top3 := m[3]
	upperMid0 := m[4]
	upperMid1 := m[5]
	upperMid2 := m[6]
	upperMid3 := m[7]
	lowerMid0 := m[8]
	lowerMid1 := m[9]
	lowerMid2 := m[10]
	lowerMid3 := m[11]

	t.Label = p.Label
	t.LabelAlign = p.LabelAlign
	t.X = (top0 * p.X) + (top1 * p.Y) + (top2 * p.Z) + top3
	t.Y = (upperMid0 * p.X) + (upperMid1 * p.Y) + (upperMid2 * p.Z) + upperMid3
	t.Z = (lowerMid0 * p.X) + (lowerMid1 * p.Y) + (lowerMid2 * p.Z) + lowerMid3
	return
}

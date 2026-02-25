module github.com/faiface/gui/examples/paint

go 1.25.5

require (
	github.com/faiface/gui v0.0.0-20190522095505-ed00d80d15da
	github.com/faiface/mainthread v0.0.0-20171120011319-8b78f0a41ae3
	github.com/fogleman/gg v1.3.0
)

require (
	github.com/go-gl/gl v0.0.0-20231021071112-07e5d0ea2e71 // indirect
	github.com/go-gl/glfw v0.0.0-20250301202403-da16c1255728 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	golang.org/x/image v0.36.0 // indirect
)

replace github.com/faiface/gui => ../../

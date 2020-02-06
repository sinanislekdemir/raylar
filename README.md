# raylar render engine

Work in progress :)

install: _you need golang to install_

    go install github.com/sinanislekdemir/raylar

- [x] Raytracing
  - [x] KD-Tree
- [x] Texture support (png, jpeg)
- [x] Ambient Occlusion
- [x] Ambient Color
- [x] Point lights
- [x] Light Objects (and area light)
- [x] Basic Reflections

## Happy Buddha Example (1088700 triangles in 1h55m37s 3200x1800)
(This is a raw cropped image - No after-effects applied)

![metro](https://www.islekdemir.com/metro.png)

(Scene can be downloaded from [https://www.islekdemir.com/buddha.tar.gz](https://www.islekdemir.com/buddha.tar.gz))

![budha](https://www.islekdemir.com/budha.jpg)

    2020/01/27 17:57:06 Loading configuration from config.json
    2020/01/27 17:57:06 Unmarshal JSON
    2020/01/27 17:57:06 Loading file: scene.json
    2020/01/27 17:57:06 Unmarshal JSON
    2020/01/27 17:57:17 Fixing object Ws
    2020/01/27 17:57:17 Flatten Scene Objects
    2020/01/27 17:57:17 Transform object vertices to absolute and build KDTrees
    2020/01/27 17:57:17 Prepare object house
    2020/01/27 17:57:17 Local to absolute
    2020/01/27 17:57:17 Unify triangles
    2020/01/27 17:57:18 Build KDTree
    2020/01/27 17:58:15 Built 2004709 nodes with 26 max depth, object ready
    2020/01/27 17:58:15 Parse material textures
    2020/01/27 17:58:15 Calculating ambient parameters
    2020/01/27 17:58:15 Ambient max radius: 3.406166
    2020/01/27 17:58:15 Exterior Scene
    2020/01/27 17:58:15 Number of vertices: 0
    2020/01/27 17:58:15 Number of indices: 1088700
    2020/01/27 17:58:15 Number of materials: 8
    2020/01/27 17:58:15 Number of triangles: 1088700
    2020/01/27 17:58:15 Loaded scene in 69.398893 seconds
    2020/01/27 17:58:15 Start rendering scene
    2020/01/27 17:58:15 Output image size: 3200 x 1800
     5760000 / 5760000 [====================================] 100.00% 1h55m37s
    2020/01/27 19:53:53 Rendered scene in 6937.625351 seconds
    2020/01/27 19:53:53 Post processing and saving file


## Refraction - Glass Example
![refraction](https://www.islekdemir.com/refraction.png)

## Pisa Example (Light + Ambient + No Color):
![pisa](https://www.islekdemir.com/image.png)

# Wall (No Lights + Ambient + Textures)
![wall](https://www.islekdemir.com/wall.png)

# Reflections (Lights + Ambient + Reflection + Textures)
![reflection](https://www.islekdemir.com/reflections.png)

# Light Object
![light](https://www.islekdemir.com/area_lights.png)

# Lucy (3 Lights (RGB) + Ambient)
![lucy](https://www.islekdemir.com/image_1.png)

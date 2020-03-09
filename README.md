# raylar render engine

Work in progress :)

install: _you need golang to install_

    go install github.com/sinanislekdemir/raylar

![glasses](https://www.islekdemir.com/teapot_1200.png)

- [x] Raytracing
  - [x] KD-Tree
- [x] Texture support (png, jpeg)
- [x] Ambient Occlusion
- [x] Ambient Color
- [x] Point lights
- [x] Light Objects (and area light)
- [x] Basic Reflections
- [x] Bump Mapping
- [x] Alpha Channel

## Stages of rendering (without Caustics)

### Ambient Occlusion Only

![ao_kitchen](https://www.islekdemir.com/01_kitchen_ao.png)

### Ambient Occlusion with Colors

![ao_kitchen_color](https://www.islekdemir.com/02_kitchen_ao_color.png)

### AO with Colors and Reflections/Refractions

![ao_kitchen_ref_color](https://www.islekdemir.com/03_kitchen_ao_color_ref.png)

### Render with AO + Lights + Colors

![kitchen_full](https://www.islekdemir.com/04_kitchen_ao_color_ref_light.png)

![residential](https://www.islekdemir.com/residental.png)

![metro_int](https://www.islekdemir.com/mmetro.png)

(Scene can be downloaded from [https://www.islekdemir.com/buddha.tar.gz](https://www.islekdemir.com/buddha.tar.gz))

## Happy Buddha Example (1088700 triangles in 1h55m37s 3200x1800 in January 2020, only 9 minutes in March 2020)
(This is a raw cropped image - No after-effects applied)

![budha](https://www.islekdemir.com/buddha_new.jpg)

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

### After optimizations by 9th of March 2020:

    2020/03/09 09:14:30 Initializing the scene
    2020/03/09 09:14:30 Loading configuration from /home/sinan/Desktop/buddha/buddha/config.json
    2020/03/09 09:14:30 Unmarshal JSON
    2020/03/09 09:14:30 Loading file: /home/sinan/Desktop/buddha/buddha/scene.json
    2020/03/09 09:14:30 Unmarshal JSON
    2020/03/09 09:14:38 Fixing object Ws
    2020/03/09 09:14:38 Loaded scene in 8.879676 seconds
    2020/03/09 09:14:38 Render 100 percent of the image
    2020/03/09 09:14:38 Set size to 3200x1800
    2020/03/09 09:14:38 Start rendering scene
    2020/03/09 09:14:38 Init scene
    2020/03/09 09:14:38 Flatten Scene Objects
    2020/03/09 09:14:38 Transform object vertices to absolute and build KDTrees
    2020/03/09 09:14:38 Prepare object happy_vrip
    2020/03/09 09:14:38 Local to absolute
    2020/03/09 09:14:39 Unify triangles
    2020/03/09 09:14:40 Loaded object with 1087716 triangles
    2020/03/09 09:14:40 Prepare object Cube.001
    2020/03/09 09:14:40 Local to absolute
    2020/03/09 09:14:40 Unify triangles
    2020/03/09 09:14:40 Loaded object with 12 triangles
    2020/03/09 09:14:40 Prepare object Sphere
    2020/03/09 09:14:40 Local to absolute
    2020/03/09 09:14:40 Unify triangles
    2020/03/09 09:14:40 Loaded object with 960 triangles
    2020/03/09 09:14:40 Prepare object Cube
    2020/03/09 09:14:40 Local to absolute
    2020/03/09 09:14:40 Unify triangles
    2020/03/09 09:14:40 Loaded object with 54 triangles
    2020/03/09 09:14:40 Build KDTree
    2020/03/09 09:15:46 Built 2019287 nodes with 26 max depth, object ready
    2020/03/09 09:15:46 Parse material textures
    2020/03/09 09:15:46 Scanning pixels on view
    5760000 / 5760000 [==========================] 100.00% 4s
    2020/03/09 09:15:51 Done scanning pixels
    2020/03/09 09:15:51 Done init scene
    2020/03/09 09:15:51 Initial rendering: 3200 x 1800
    5760000 / 5760000 [==========================] 100.00% 7m35s
    2020/03/09 09:23:27 Rendered scene in 455.947010 seconds
    2020/03/09 09:23:27 Second pass for antialiasing and image generation
    5760000 / 5760000 [==========================] 100.00% 2m10s



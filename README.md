# Raylar Render Engine

Raylar is a simple and stand-alone raytracer.
Raylar does not intend to be a fully-featured high industry render engine. It does not give so many tweaks and detailed materials. It is more for hobbyists. 
It supports:
* Ambient Occlusion
* Direct/Sunlight
* Point light
* Area light by illuminating material
  * Object as a light source
* Transparency with refraction
* Reflective / Glossy materials with roughness
* Bump maps
* Textures with transparency (JPEG/PNG)
* Environment cube-map

Raylar comes with a Blender script for Blender 2.8x that can export your:
* Meshes
  * With basic materials
* Lights (point light / sun)
* Cameras

Raylar is written in Golang. It is probably not the best choice. Even though it has a nice garbage collection, it is still behind C, C++, Rust in terms of performance. But I tried to tighten every bolt to make it perform at the best speed. Raylar was a challenge for me to learn about profiling and debugging in Golang. Along with some algorithms like KD-Tree. 

## Installing

Raylar does not have any external dependencies. It uses the Golang standard library and can be built against _almost_ any Golang supported platform.
You can install Raylar by visiting the releases page in Github. 

[Raylar Releases Page](https://github.com/sinanislekdemir/raylar/releases)

Or you can build Raylar by yourself if you want to:

    go install github.com/sinanislekdemir/raylar

## Tests

Unfortunately, Raylar is missing it's unit tests. When I first started the project, I realized that unit tests were not enough to ensure the result. A raytracer has so many internal dynamics and I am not following rigid formulas. Also, there are so many routines. Sometimes, I just mingle with stuff or come up with an approximation instead of complex calculations that are decent enough.
As a result, I do have some unit scenes to test some features and dynamics of the engine. With each iteration, I run the engine against the test scenes to see if I can still get the same quality or better.
You can find the scenes inside the tests directory.

## Setting up Blender Addon and basic usage

### Install addon

![Install Addon](https://www.islekdemir.com/install_addon.gif)

### Using materials

Raylar uses Principled BSDF Material type in Blender.

#### Color Material

Just setting the Base Color option to any RGB should work fine.


## Some sample renders (some are old, so might not reflect the latest state of the engine)



## Stages of rendering (without Caustics)

### Ambient Occlusion Only

![ao_kitchen](https://www.islekdemir.com/01_kitchen_ao.png)

### Ambient Occlusion with Colors

![ao_kitchen_color](https://www.islekdemir.com/02_kitchen_ao_color.png)

### AO with Colors and Reflections/Refractions

![ao_kitchen_ref_color](https://www.islekdemir.com/03_kitchen_ao_color_ref.png)

### Render with AO + Lights + Colors

![kitchen_full](https://www.islekdemir.com/04_kitchen_ao_color_ref_light.png)


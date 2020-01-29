# Raylar Blender Scene Export

This is just a dummy add-on. Just like Raylar Raytracing is. I just created this add-on
to help myself to create my scene file easily without needing to go through Blender OBJ Export - Payton Script - Scene Setup - Export as json.

Just install the Addon on Blender 2.8x and from export, choose "Raylar Export (scene.json)"

![Addon](https://www.islekdemir.com/blender1.png)

## Materials:

This script assumes to find materials in "Principled BSDF" Surface material. 

![BSDF](https://www.islekdemir.com/blender2.png)

Currently, "Base Color" as color and Image are supported.

Also, to get reflections, you can change "Metallic" value;

To get a light material, change material shader form "Principled BSDF" to "Emission"

![Emission](https://www.islekdemir.com/blender3.png)

To get a transparent - glass like material, use "Transmission" value along with IOR.

IOR Stands for "Index of Refraction" so it is the medium index. Higher values will refract light in a bigger angle;

![Refraction](https://www.islekdemir.com/blender4.png)
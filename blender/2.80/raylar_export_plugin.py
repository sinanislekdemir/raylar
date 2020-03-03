import json
import math
import os
from math import *  # noqa

import bmesh
import bpy
import mathutils
from bpy.props import StringProperty
from bpy.types import Operator
from shutil import copyfile, SameFileError


# ExportHelper is a helper class, defines filename and
# invoke() function which calls the file selector.
from bpy_extras.io_utils import ExportHelper
from mathutils import Matrix, Vector

bl_info = {
    "name": "Raylar Export",
    "author": "sinan islekdemir",
    "version": (0, 0, 0, 2),
    "blender": (2, 80, 0),
}

global_matrix = mathutils.Matrix.Rotation(-math.pi / 2.0, 4, "X")
MAT_CONVERT_CAMERA = Matrix.Rotation(math.pi / 2.0, 4, "Y")
global_assets = []


def write_raylar_data(filepath):
    global global_assets
    print("running raylar export...")
    scene = construct_scene()
    context = json.dumps(scene)
    f = open(filepath, "w", encoding="utf-8")
    f.write(context)
    f.close()

    target_path = os.path.dirname(filepath)

    for f in global_assets:
        base_name = os.path.basename(f)
        try:
            copyfile(f, os.path.join(target_path, base_name))
        except SameFileError:
            pass
    global_assets = []

    return {"FINISHED"}


def export_object(obj):
    if obj.type != "MESH":
        return
    material_cache = {}
    for i, matslot in enumerate(obj.material_slots):
        material = matslot.material
        material_cache[material.name] = {"_index": i}
        mkeys = material.node_tree.nodes.keys()
        if "Principled BSDF" in mkeys:
            inp = material.node_tree.nodes["Principled BSDF"].inputs
            if "Base Color" in inp:
                material_cache[material.name]["color"] = [
                    inp["Base Color"].default_value[0],
                    inp["Base Color"].default_value[1],
                    inp["Base Color"].default_value[2],
                    inp["Base Color"].default_value[3],
                ]
                material_cache[material.name]["light"] = False
            else:
                material_cache[material.name]["color"] = [1, 1, 1, 1]
            if "Alpha" in inp:
                material_cache[material.name]["transmission"] = inp[
                    "Transmission"
                ].default_value
            if "IOR" in inp:
                material_cache[material.name]["index_of_refraction"] = inp[
                    "IOR"
                ].default_value
            if "Metallic" in inp:
                material_cache[material.name]["glossiness"] = inp[
                    "Metallic"
                ].default_value
            if "Roughness" in inp:
                material_cache[material.name]["roughness"] = inp[
                    "Roughness"
                ].default_value
        if "Emission" in mkeys:
            inp = material.node_tree.nodes["Emission"].inputs
            if "Color" in inp:
                material_cache[material.name]["color"] = [
                    inp["Color"].default_value[0],
                    inp["Color"].default_value[1],
                    inp["Color"].default_value[2],
                    inp["Color"].default_value[3],
                ]
                material_cache[material.name]["light"] = True
                material_cache[material.name]["light_strength"] = inp[
                    "Strength"
                ].default_value
        if "Image Texture" in mkeys:
            image = material.node_tree.nodes["Image Texture"].image
            inp = image.filepath_from_user()
            global_assets.append(inp)
            base_name = os.path.basename(inp)
            material_cache[material.name]["texture"] = base_name

    odata = obj.data
    original_data = odata.copy()  # Backup data
    bm = bmesh.new()
    bm.from_mesh(odata)
    bmesh.ops.triangulate(
        bm, faces=bm.faces[:], quad_method="BEAUTY", ngon_method="BEAUTY"
    )
    bm.to_mesh(odata)  # Triangulate the object

    vertices = []
    normals = []
    texcoords = []
    index = 0
    uvLayer = bm.loops.layers.uv.active

    for face in bm.faces:
        for loop in face.loops:
            # Get position (swizzled)

            vertices.append([loop.vert.co[0],
                             loop.vert.co[1],
                             loop.vert.co[2]])

            # Get normal (swizzled)
            # TODO: Should this be face, loop, or vertex normal?
            norm = loop.vert.normal
            normals.append([norm[0], norm[1], norm[2]])

            # Get first UV layer
            if uvLayer is not None:
                texcoords.append([loop[uvLayer].uv[0], loop[uvLayer].uv[1]])

        for mat in material_cache:
            if material_cache[mat]["_index"] == face.material_index:
                if "indices" not in material_cache[mat]:
                    material_cache[mat]["indices"] = []
                material_cache[mat]["indices"].append([index,
                                                       index + 1,
                                                       index + 2,
                                                       int(face.smooth)])
        index += 3

    obj_dict = {
        "vertices": vertices,
        "normals": normals,
        "texcoords": texcoords,
        "matrix": _conv_matrix(obj.matrix_local),
        "materials": material_cache,
        "children": {},
    }

    # Revert back the original object
    obj.data = original_data

    return obj_dict


def export_light(light):
    directional = False
    direction = [0, 0, 0, 0]
    if bpy.data.lights[light.name].type == 'SUN':
        directional = True
        lmw = light.matrix_world
        direction = lmw.to_quaternion() @ Vector((0.0, 0.0, -1.0))

    return {
        "position": list(light.location),
        "color": list(bpy.data.lights[light.name].color),
        "active": True,
        "light_strength": bpy.data.lights[light.name].energy / 10,
        "directional_light": directional,
        "direction": list(direction)
    }


def _conv_matrix(matrix):
    return [
        [matrix[0][0], matrix[1][0], matrix[2][0], matrix[3][0]],
        [matrix[0][1], matrix[1][1], matrix[2][1], matrix[3][1]],
        [matrix[0][2], matrix[1][2], matrix[2][2], matrix[3][2]],
        [matrix[0][3], matrix[1][3], matrix[2][3], matrix[3][3]],
    ]


def export_camera(camera):
    position = camera.location
    cmw = camera.matrix_world
    up = cmw.to_quaternion() @ Vector((0.0, 1.0, 0.0))
    cam_direction = cmw.to_quaternion() @ Vector((0.0, 0.0, -1.0))
    x = (cam_direction[0] * 10) + position[0]
    y = (cam_direction[1] * 10) + position[1]
    z = (cam_direction[2] * 10) + position[2]
    target = [x, y, z, 1]

    fov = bpy.data.cameras[camera.name].angle * 180 / math.pi
    aspect = (
        bpy.context.scene.render.resolution_x /
        bpy.context.scene.render.resolution_y
    )

    return {
        "position": list(position),
        "target": list(target),
        "up": list(up),
        "fov": fov,
        "aspect_ratio": aspect,
        "near": 0.01,
        "far": 10000,
        "perspective": True,
    }


def construct_scene():
    scene = {"objects": {}, "lights": [], "observers": []}

    bpy_scene = bpy.context.scene
    for obj in bpy_scene.objects:
        obj.select_set(True)
        bpy.context.view_layer.objects.active = obj
        bpy.ops.object.transform_apply(location=True,
                                       scale=True,
                                       rotation=True)
        bpy.ops.object.select_all(action="DESELECT")
        obj.select_set(False)

        if obj.type == "MESH":
            scene["objects"][obj.name] = export_object(obj)
        if obj.type == "LIGHT":
            scene["lights"].append(export_light(obj))
        if obj.type == "CAMERA":
            scene["observers"].append(export_camera(obj))
    return scene


class ExportRaylarData(Operator, ExportHelper):
    """This appears in the tooltip of the operator and in the generated docs"""

    bl_idname = "export_payton.scene_data"
    bl_label = "Export Scene to Payton/Raylar JSON"

    # ExportHelper mixin class uses this
    filename_ext = ".json"

    filter_glob: StringProperty(
        default="*.json",
        options={"HIDDEN"},
        maxlen=255,  # Max internal buffer length, longer would be clamped.
    )

    def execute(self, context):
        return write_raylar_data(self.filepath)


# Only needed if you want to add into a dynamic menu
def menu_func_export(self, context):
    self.layout.operator(ExportRaylarData.bl_idname,
                         text="Raylar Export (scene.json)")


def register():
    bpy.utils.register_class(ExportRaylarData)
    bpy.types.TOPBAR_MT_file_export.append(menu_func_export)


def unregister():
    bpy.utils.unregister_class(ExportRaylarData)
    bpy.types.TOPBAR_MT_file_export.remove(menu_func_export)


if __name__ == "__main__":
    register()

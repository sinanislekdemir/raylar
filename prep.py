import json

from payton.scene import Scene
from payton.scene.geometry import Wavefront, Plane, Cube
from payton.scene.light import Light

s = Scene(width=1600, height=900)


def timer(period, total):
    print(s.active_observer.position)
    print(s.active_observer.target)
    open("scene.json", "w").write(json.dumps(s.to_dict(), indent=1))
    exit(0)

g = Plane(width=20, height=20)


c = Cube()
c.position = [0, 0, 1]

c2 = Cube()
c2.position = [1, 1.2, 1.8]

c3 = Cube()
c3.position = [3, -1, 2]

# s.add_object("cube1", c)
# s.add_object("cube2", c2)
# s.add_object("cube3", c3)
# s.add_object("ground", g)
wavefront = Wavefront(filename="/home/sinan/X-WING.obj")
wavefront.position = [0, 0, 0.5]
s.add_object("house", wavefront)
# s.lights[0].position = [2, -15, 10] ---> room
s.lights[0].position = [13, 15, 12]
s.lights[0].color = [253/255, 240/255, 235/255]

# light2 = Light(position=[3.5, -20, 12], color=[253/255, 179/255, 83/255])
# s.lights.append(light2)

# light3 = Light(position=[3.25, -20, 12], color=[253/255, 179/255, 83/255])
# s.lights.append(light3)

c2.material.texture = (
    "/home/sinan/go/src/github.com/sinanislekdemir/raylar/cube.png"
)
s.active_observer.position = [1.9546827811305514, 16.219105441381064, 3.345885815902124]
s.active_observer.target = [1.4643116667866707, 0.1902765380218625, 2.542799100279808]

s.create_clock("timer", 1.0, timer)

s.run()

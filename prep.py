import json

from payton.scene import Scene
from payton.scene.geometry import Cube, Plane

s = Scene()


def timer(period, total):
    print(s.active_observer.position)
    print(s.active_observer.target)


g = Plane(width=20, height=20)
s.add_object("ground", g)

c = Cube()
c.position = [0, 0, 1]

c2 = Cube()
c2.position = [1, 1.2, 1.8]

s.add_object("cube1", c)
s.add_object("cube2", c2)

c2.material.texture = (
    "/home/sinan/go/src/github.com/sinanislekdemir/raylar/cube.png"
)
s.active_observer.position = [
    5.467899728250518,
    3.6667874605174835,
    1.490295302887703,
]
s.active_observer.target = [
    0.38242703676223755,
    0.4583794707432389,
    0.7661657929420471,
]

s.create_clock("timer", 1.0, timer)
open("scene.json", "w").write(json.dumps(s.to_dict(), indent=1))
s.run()

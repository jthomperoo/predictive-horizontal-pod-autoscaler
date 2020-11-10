

import random
import json
_i = {}
_p = {}
d = {}
for i in range(1, 5):
    _i[i] = random.randint(0, 10)
    _p[i] = random.randint(0, 10)
d["index"] = _i
d["value"] = _p

print(json.dumps(d))
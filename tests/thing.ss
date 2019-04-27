package main

import
  bytes
--

global
  AnimationRate       = 5
  Gravity             = 0.01
  AnimationNotDone    = 0
  AnimationAlmostDone = 1
  AnimationDone       = 2

  HumanUID  = uint16(0)
  BaronUID  = uint16(1)
  TreeUID   = uint16(2) 
  PlasmaUID = uint16(3)
  MedkitUID = uint16(4)

  NoGroup    = 0 
  HumanGroup = 1
  DemonGroup = 2
  
  mutable ThingNetworkNum = uint16(0)
--

object thing
  world          world
  uid            uint16
  nid            uint16
  animation      int
  animationFrame int
  x              float32
  y              float32
  z              float32
  angle          float32
  deltaX         float32
  deltaY         float32
  deltaZ         float32
  oldX           float32
  oldZ           float32
  minBX          int
  minBY          int
  minBZ          int
  maxBX          int
  maxBY          int
  maxBZ          int
  ground         bool
  radius         float32
  height         float32
  speed          float32
  health         int
  group          int
  deltaMoveXZ    bool
  deltaMoveY     bool
  update         func() bool
  damage         func(int)
  save           func(bytes.buffer)
  snap           func(bytes.buffer)
  binary         []byte
--

func NextNid() uint16
  ThingNetworkNum += 1
  return ThingNetworkNum
--

func LoadNewThing(w world uid uint16 x float32 y float32 z float32)
  switch uid
  case BaronUID
      NewBaron(w x y z)
  --
  case TreeUID
      NewBaron(w x y z)
  --
--

func thing.nopUpdate() bool
  return false
--

func thing.nopSnap(b bytes.Buffer)
  me.binary = null
--

func thing.nopDamage(amount int)
--

func thing.blockBorders()
  me.minBX = int((me.x - me.radius) * InverseBlockSize)
  me.minBY = int(me.y * InverseBlockSize)
  me.minBZ = int((me.z - me.radius) * InverseBlockSize)
  me.maxBX = int((me.x + me.radius) * InverseBlockSize)
  me.maxBY = int((me.y + me.height) * InverseBlockSize)
  me.maxBZ = int((me.z + me.radius) * InverseBlockSize)
--

func thing.addToBlocks()
  loop gx = me.minBX gx <= me.maxBX gx += 1
    loop gy = me.minBY gy <= me.maxBY gy += 1
      loop gz = me.minBZ gz <= me.maxBZ gz += 1
        me.world.getBlock(gx gy gz).addThing(me)
      --
    --
  --
--

func thing.removeFromBlocks()
  loop gx = me.minBX gx <= me.maxBX gx += 1
    loop gy = me.minBY gy <= me.maxBY gy += 1
      loop gz = me.minBZ gz <= me.maxBZ gz += 1
        me.world.getBlock(gx gy gz).removeThing(me)
      --
    --
  --
--

func thing.updateAnimation() int
  me.animationFrame += 1
  if me.animationFrame = me.animation - AnimationRate
    return AnimationAlmostDone
  --
  elif me.animationFrame = me.animation
    return AnimationDone
  --
  return AnimationNotDone
--

func thing.overlap(b thing) bool
  square = me.radius + b.radius
  return abs(me.x - b.x) <= square and abs(me.z - b.z) <= square
--

func thing.tryOverlap(x float32 z float32 b thing) bool
  square = me.radius + b.radius
  return abs(x - b.x) <= square and abs(z - b.z) <= square
--

func thing.approximateDistance(b thing) float32
  dx = abs(me.x - b.x)
  dz = abs(me.z - b.z)
  if dx > dz
    return dx + dz - dz * 0.5
  --
  return dx + dz - dx * 0.5
--

func thing.integrateY()
  if me.ground
    return
  --
  me.deltaY -= Gravity
  me.y += me.deltaY
  me.deltaMoveY = true
  me.terrainCollisionY()
  me.removeFromBlocks()
  me.blockBorders()
  me.addToBlocks()
--

---
title: "Starting a colony sim, lessons I'm carrying forward"
description: "Kicking off a new Odin prototype and the grid-game patterns I already trust: cameras, tilemaps, and how to lay out entity data."
date: 2026-07-03
type: lesson
series: colony-sim-prototype
series_order: 1
languages: ["odin"]
tags: ["odin", "colony-sim", "game-dev"]
---

I'm starting a second prototype, a colony simulation, in [Odin](https://odin-lang.org/) and Raylib. It's early days and there's not much game there yet. This post isn't a feature tour. It's the stuff I already know works because I learned it building the [tower defense prototype](https://github.com/Codegirl-Games/tower-defense-prototype) first.

If you're starting a top-down grid game, these are the foundations I keep reaching for.

## The 2D camera is just coordinate math

A 2D camera doesn't need a library. It's two numbers and two functions:

- **Offset**: which world point sits at the center of the screen
- **Zoom**: how many world pixels map to one screen pixel

Convert world → screen:

```
screen = (world - offset) * zoom + screen_center
```

Convert screen → world (for mouse input):

```
world = (screen - screen_center) / zoom + offset
```

That's it. Panning moves the offset. The scroll wheel clamps zoom between sensible min/max values. Pan speed gets divided by zoom so movement feels consistent when you're zoomed in.

The lesson I keep re-learning: **every mouse click must go through `screen_to_world` before you do anything useful.** Selection, movement commands, building placement: all of it. Forget this once and your clicks drift when the camera moves.

## Tilemaps are flat arrays with helpers

A tilemap is a width, a height, and a flat buffer. Tile `(x, y)` lives at index `y * width + x`. Wrap access in two helpers and never think about the math again:

- `tile_index(map, x, y)`: buffer lookup
- `tile_in_bounds(map, x, y)`: guard every read and write

World position to tile coordinate is just `floor(world / tile_size)` on each axis. I used the same pattern in both prototypes. The colony sim added a comment in `tile_in_bounds`, *"learning from another game, this will become handy"*, because I skipped it early in the tower defense project and paid for it later.

Terrain type can start as a `u8` per tile. A switch or lookup table maps type → color. Don't over-engineer biomes on day one; get the grid drawing and the coordinate conversions right first.

When drawing, multiply tile size by zoom and cull tiles that fall off-screen. An 80×60 map is 4,800 rectangles, fine for a prototype, but the cull pass is free and keeps the pattern honest for bigger maps.

## Logical tiles vs visual position

Grid games have two positions whether you plan for it or not:

- **Logical position**: which tile the entity occupies (`Tile_Coord{3, 7}`)
- **Visual position**: where the sprite actually renders (smoothly interpolated between tile centers)

The colonist's grid cell updates one step at a time. The circle on screen lerps toward the next tile center each frame. Gameplay stays discrete; motion looks continuous. Mix these up and pathfinding, collision, and selection all get harder.

I didn't need this in the tower defense game: enemies moved in continuous world space along a path. Colony sims live on tiles. Separate the two early.

## Struct of arrays vs array of structs

Both layouts show up in my code. Neither is always wrong.

**Array of structs (AoS)**, what the tower defense prototype uses:

```odin
enemies: [MAX_ENEMIES]Enemy,
```

Each slot is a full `Enemy` struct: position, health, speed, active flag, all together. Natural to read: `enemy.health -= damage`. Good when you often touch most fields on one entity at once.

**Struct of arrays (SoA)**, what the colony sim uses:

```odin
Entity_World :: struct {
    active:          [MAX_ENTITIES]bool,
    position:        [MAX_ENTITIES]Tile_Coord,
    move_state:      [MAX_ENTITIES]Move_State,
    visual_position: [MAX_ENTITIES]rl.Vector2,
    // ...
}
```

Each field is a parallel array across all entities. Good when you update one system at a time (move every entity, then draw every entity) and when entities are sparse (lots of inactive slots in a fixed pool).

My rule of thumb so far:

| Reach for AoS when… | Reach for SoA when… |
|---|---|
| Entities are few and always accessed whole | You iterate one component across many entities |
| Struct fits in cache and you're touching most of it | Many slots are inactive (object pool) |
| Code clarity matters more than layout | Systems are split (movement, render, AI) |

Both prototypes use **fixed pools** with an `active` flag, no allocate/free per spawn. That pattern transferred directly from tower defense to colony sim regardless of AoS vs SoA.

## Entity handles, not raw indices

The colony sim returns `Entity`, a `distinct u32`, instead of passing array indices around. Internally it's `index + 1`, with `0` meaning invalid. Small thing, but it stops you from accidentally passing a tile coordinate or a mouse value where an entity ID goes.

## Update and draw stay separate

Both games follow the same loop shape:

1. Read input
2. Update simulation (`world_update`, `entity_update_movement`)
3. Draw (`tilemap_draw`, `entity_draw`)

Simulation code never calls draw functions. Draw code never changes game state. Obvious, but worth stating because it's the seam that keeps things readable as files multiply.

## What I'm not writing about yet

Pathfinding, job queues, resources, building: none of that exists in the repo yet. When there's a full week of work to show, I'll write a proper devlog. For now, the [repo](https://github.com/Codegirl-Games/colony-sim-prototype) is a camera, a tilemap, one colonist, and a right-click move command.

The game will come. These patterns are the part I'm confident in.

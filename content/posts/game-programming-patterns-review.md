---
title: "Game Programming Patterns, first impressions"
description: "A book review of Robert Nystrom's Game Programming Patterns and how I use it in my tower defense prototype."
date: 2026-06-28
type: books
book: "Game Programming Patterns"
tags: ["patterns", "books"]
---

*Game Programming Patterns* by Robert Nystrom is the reference I'm using as I build. The patterns aren't abstract; I'm already applying several of them in my [Odin tower defense prototype](https://github.com/Codegirl-Games/tower-defense-prototype).

Here is where the book shows up in real code so far.

## Patterns in the tower defense game

### Object Pool

Enemies and projectiles are never allocated per spawn. Both use fixed arrays (`[MAX_ENEMIES]Enemy`, `[MAX_PROJECTILES]Projectile`) with `acquire_enemy` and `acquire_projectile` scanning for inactive slots. A spawned unit resets its fields and sets `active = true`; on death or impact, `active = false` returns the slot to the pool.

This was the first pattern I reached for, in week one. Every wave can spawn dozens of enemies and towers can fire many projectiles per second. Avoiding allocate/free in the hot path keeps the update loop predictable.

### Command

Player input goes through a command layer, not straight into game logic. `gather_commands` and `poll_controls_command` read keyboard and mouse state and return a `Command` value: `Place_Tower`, `Start_Wave`, `Upgrade_Tower`, or `None`. `execute_command` dispatches to the right handler.

UI buttons, hotkeys, and mouse clicks all produce the same command type. Adding a new input source does not mean rewriting placement or wave logic.

### Event Queue

Gold changes do not happen inside combat code directly. When an enemy dies or a wave is cleared, systems call `push_event` with `Enemy_Killed` or `Wave_Survived`. At the end of `update_world`, `process_events` drains the queue and applies gold rewards.

Combat systems announce what happened; the economy system decides what it costs. That separation kept the update loop readable as towers, waves, and enemy types piled on.

### State

The game runs in explicit phases: `Build`, `Combat`, `Game_Over`, and `Victory`. Phase controls what input is accepted (you cannot place towers during combat), when waves can start, and when the simulation stops updating.

This is a straightforward state machine, not a deep AI behavior tree, but the same idea: one variable drives which rules apply this frame.

### Update Method

Each system owns an update function called once per frame from `update_world`: `update_wave`, `update_enemies`, `update_projectiles`, `update_towers`, `update_phase`. No single giant function walks every entity type inline.

The book's "one game loop, many systems" structure maps cleanly onto separate `.odin` files as the project grew.

### Game Loop

`app.odin` runs the classic loop: read input, call `update_world`, render the map, entities, overlay, and controls. Update and draw stay separate; simulation code never calls draw functions.

### Data Locality

Enemies and projectiles live in contiguous fixed arrays rather than scattered heap allocations. I iterate the full pool each frame but skip inactive slots. Not as aggressive as a struct-of-arrays layout, but the fixed-buffer approach gives similar benefits: no pointer chasing, no allocator pressure during combat.

### Type Object (archetype tables)

Tower and enemy stats live in shared lookup tables (`TOWER_ARCHETYPES`, `ENEMY_DEF`), not duplicated on every instance. A placed tower stores its kind, position, cooldown, and upgrade level; range, damage, cost, and footprint come from the table row.

Adding a cannon or a tank enemy is mostly a new table entry plus a behavior branch, not a new class hierarchy.

### Subclass Sandbox

Tower types differ by `switch t.kind` in `update_towers`: archers spawn homing projectiles, cannons fire ballistic shots with splash, ice towers apply slow on hit. I skipped inheritance trees in favor of enum variants and explicit branches. Nystrom argues this is the right call when you only have a handful of types and the differences are behavioral, not structural.

### Spatial partition (grid)

The map is a 2D tile grid (`Blocked`, `Path`, `Build`). Tower placement queries `can_build_at` and footprint overlap against grid cells, not against every entity on the map. A full spatial hash would be overkill at this scale; the grid already gives O(1) tile lookups for build rules.

## What I have not used yet

Patterns I expect to need later but have not implemented in this prototype:

- **Pathfinding** for dynamic routes when tower placement blocks the path
- **Behavior trees** or a deeper **State** machine per enemy for complex AI
- **Observer** beyond the simple event queue (e.g. UI reacting to stat changes)
- **Component** or **Entity-Component-System** if entity types multiply significantly

## Why I keep the book nearby

Nystrom names the patterns, explains the tradeoffs, and shows when *not* to use them. That matches how I work: reach for Object Pool and Command early because the problem is obvious; hold off on ECS until the entity count justifies the complexity.

Expect follow-up posts that go deeper on individual chapters as I apply more patterns to the tower defense game and the colony sim.

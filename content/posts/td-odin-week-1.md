---
title: "Week 1: Tower defense in Odin, project kickoff"
description: "First week building my Odin tower defense prototype: grid map, object pool, commands, events, and a playable build/combat loop."
date: 2026-06-12
type: devlog
series: tower-defense-prototype
series_order: 2
languages: ["odin"]
tags: ["odin", "tower-defense", "devlog"]
---

First devlog in the tower defense series. This week I kicked off the [Odin prototype](https://github.com/Codegirl-Games/tower-defense-prototype) and went from a blank `main.odin` to a playable loop: build towers, start a wave, watch enemies walk the path, lose base health when they leak through.

![Early map with build zone and L-shaped path](/images/tower-defense/td-week-1-map.png)

## What I built

**Day 1: map and loop.** Switched the project to Odin + Raylib. Added a fixed 30×20 tile grid with three tile kinds: blocked grass, a build zone, and a hand-authored L-shaped path. The game window is exactly map-sized (`MAP_W * TILE_SIZE`), and the main loop separates `update_world` from `render_world`.

**Enemies that move.** Enemies spawn at the path start and step toward waypoints extracted from path tiles on the grid. Movement is simple vector math, no steering, no physics engine.

**Object pool from day one.** Enemies live in a fixed `[MAX_ENEMIES]Enemy` array. `acquire_enemy` reuses inactive slots instead of allocating every spawn. This is the first [Game Programming Patterns](https://gameprogrammingpatterns.com/object-pool.html) idea I reached for, and it fit Odin naturally: explicit memory, no hidden allocations.

**Commands and events.** Input goes through a small command layer (`Place_Tower`, `Start_Wave`). Gold changes go through an event queue (`Enemy_Killed`, `Wave_Survived`) so combat logic does not touch the economy directly. That separation made the update loop easier to read even before I had much gameplay.

**Towers, waves, and phases.** By the end of the week I had one tower type (Archer) backed by a `Tower_Archetype` table: range, damage, fire rate, cost in one place. The game alternates between **Build** and **Combat** phases. Press `N` to start a wave; a spawner drips out enemies with scaling health and speed. Towers pick the nearest target in range and apply damage directly. Survive the wave, earn gold, place more towers.

## What worked

- Odin's struct enums and fixed arrays made the object pool straightforward, no fighting the language
- Data-driven tower archetypes: adding stats later did not require rewriting placement logic
- Command + event split kept `update_world` readable as systems piled on
- Raylib got something on screen fast; I spent the week on game logic, not boilerplate

## What broke

- **No projectiles yet.** Towers subtract health instantly. It plays, but it does not look or feel like a tower defense game yet
- **Path is not pathfinding.** `build_path` scans the grid for path tiles in row order. Fine for a hand-drawn map, useless once I want procedural levels or tower placement that reroutes enemies
- **HUD lives in the renderer.** Gold, base HP, wave count, and phase hints are drawn inline in `render_world`. Works for now, will get messy
- **Tower placement during combat.** The command gatherer had a duplicate mouse handler that let you place towers mid-fight. Small bug, easy miss

## Repo snapshot

By June 12 the `game/` package had separate files for map, path, enemies, towers, waves, events, commands, and rendering: about 450 lines added across the week. Not pretty, but the skeleton for every port (C, C++) is visible: fixed pools, explicit phases, data tables instead of inheritance trees.

Next: projectiles, split out the HUD, and more tower types.

<figure class="post__video">
  <iframe
    class="post__iframe"
    src="https://www.youtube-nocookie.com/embed/MKBSuTkClys"
    title="Week 1: Tower defense in Odin, project kickoff"
    allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
    allowfullscreen
  ></iframe>
</figure>

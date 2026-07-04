---
title: "Week 4: Tests, refactors, and data tables"
description: "Fourth week on the Odin tower defense prototype: splitting tower placement, table-driven defs, first tests, and undoing a nested world refactor."
date: 2026-06-29
type: devlog
series: tower-defense-prototype
series_order: 5
languages: ["odin"]
tags: ["odin", "tower-defense", "devlog"]
---

Week four on the [Odin prototype](https://github.com/Codegirl-Games/tower-defense-prototype). Week three shipped fullscreen, three tower types, enemy variety, and splash damage. This week was quieter on features and heavier on structure: splitting files, pushing logic into data tables, and writing my first tests.

## What I built

**Split tower logic.** `tower.odin` had grown to handle combat, stats, rendering, and placement. I pulled placement into `tower_placement.odin`: footprint checks, overlap tests, `try_place_tower`, ghost preview. Combat and archetypes stay in `tower.odin`. Placement is geometry; combat is timing and targeting. Keeping them separate made both files easier to navigate.

**Data tables over switch statements.** Several `get_*` functions became lookup tables:

- `TILE_COLORS` replaced a `tile_color` switch
- `ENEMY_DEF` replaced per-kind switch logic with base stats, per-wave scaling, draw radius, and gold reward in one row per enemy
- Tower archetypes dropped a redundant `kind` field and gained a `color` column for rendering

Tweaking tank gold from 5 to 10 or archer cost became a one-line edit instead of hunting through switch arms.

**Constants consolidation.** UI layout lived in helper functions that recalculated button rectangles every frame. Moved to `constants.odin`: `ARCHER_BUTTON_RECT`, `START_WAVE_BUTTON_RECT`, `STARTING_GOLD`, `STARTING_BASE_HEALTH`, overlay colors, control bar dimensions. `controls.odin` got shorter and the layout stopped drifting.

**First tests.** Added `math2d_test.odin` (distance, `move_toward`) and `wave_test.odin` (start wave, clear wave, recipe progression). Small suite, but it caught a wave recipe memory leak during gold rebalancing: the kind of bug I'd fixed once by hand and reintroduced during a refactor.

**Economy pass.** Gold rewards moved into enemy definitions. Tank kills pay more than grunts. Tower costs live in archetype rows. Balanced enough to play-test without obvious snowball or stall.

**World struct simplified.** Flattened `base_health` and map init into cleaner helpers. Started the week by reverting the nested `Economy` / `Combat` / `Wave` sub-structs from week three, back to a flat `World` with direct field access.

## What worked

- Splitting placement from combat scaled better as tower logic grew
- Table-driven defs made adding and tuning enemy/tower stats mostly data changes
- Tests paid for themselves immediately on wave spawner edge cases
- Flat `World` struct was easier to reason about than nested sub-structs for a game this size

## What broke

- **Nested world refactor, reverted.** Week three split `World` into `Economy`, `Combat`, and `Wave` sub-structs. Looked clean on paper. Every system suddenly needed `world.combat.enemies` instead of `world.enemies`, accessors multiplied, and nothing got simpler. Reverted on June 27. Don't refactor structure until the current shape is actually hurting you, and wait until you have tests.

- **Test cleanup is manual.** Odin tests that allocate dynamic arrays need explicit `defer delete`. Forgot once; leak showed up in the test runner, not the game.

## Repo snapshot

June 27–29 touched 14 files, ~490 net lines. New modules: `tower_placement.odin`, `math2d_test.odin`, `wave_test.odin`. Three towers, three enemies, wave recipes, upgrades, splash damage, fullscreen, end screens, plus a test suite to build on.

Next: pathfinding, more tower shop wiring, ice slow polish, and whatever breaks once I add a fourth tower type.

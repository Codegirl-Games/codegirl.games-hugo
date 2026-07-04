---
title: "Week 3: Fullscreen, variety, and splash damage"
description: "Third week on the Odin tower defense prototype: end screens, render-to-texture scaling, three tower types, enemy archetypes, wave recipes, upgrades, and cannon splash."
date: 2026-06-26
type: devlog
series: tower-defense-prototype
series_order: 4
languages: ["odin"]
tags: ["odin", "tower-defense", "devlog"]
---

Week three on the [Odin prototype](https://github.com/Codegirl-Games/tower-defense-prototype). Week two left me with one tower, one enemy, and a HUD split across three files. This week the prototype started feeling like a game: win/lose screens, resizable fullscreen, three tower types, three enemy types, wave recipes, and cannons that splash.

## What I built

**End screens and unified overlay.** Added `endscreen.odin` for game over and victory overlays. Moved wave count and enemy totals into `overlay.odin` and deleted the leftover `hud.odin` split. One top bar for economy and combat status; end screens dim the world and show the result.

**Fullscreen via render texture.** New `display.odin` renders the fixed 960×640 game world into a `RenderTexture`, then scales it to whatever window size the player uses. `F11` toggles fullscreen. Mouse coordinates go through `game_mouse()` so clicks still map to grid tiles when the viewport letterboxes. The game logic stays pixel-fixed; only presentation scales.

**Three tower types.** Archer (homing arrows), Cannon (2×2 footprint, ballistic arc), Ice (slow effect, smaller footprint). Each has its own archetype row: cost, range, fire rate, footprint, upgrade caps. Shop buttons in `controls.odin` grew to match: select tower, click the map, place.

**Tower preview.** Hovering the build grid while a tower is selected draws a ghost footprint before you commit gold. Small UX win that made placement feel less like guessing.

**Enemy variety.** Replaced the single grunt with archetypes: Grunt, Runner (fast, fragile), Tank (slow, thick). Each scales health and speed with wave number and gets its own color.

**Wave recipes.** Waves are no longer `5 + wave * 2` of the same enemy. `build_wave_recipe` returns an ordered list of `(kind, count)` entries: wave 1 is six grunts, wave 2 mixes grunts and runners, later waves add tanks. The spawner drips through the recipe entry by entry.

**Tower upgrades.** Towers track `upgrade_level`. Stats scale via `tower_stats_at_level`: extra damage and range per level, paid from gold during build phase.

**Ballistic projectiles and splash.** Projectiles gained a `Projectile_Mode`: homing for archers, ballistic for cannons. Cannon shots fly in a fixed direction; on impact they call `apply_splash_damage` in a radius defined on the archetype. First time area damage changed how I thought about placement: kill zones, not just single targets.

**World refactor.** Split `World` into nested structs for economy, combat, and wave state. Gold and enemies moved out of flat fields. Cleaner ownership before more systems land.

## What worked

- Render-to-texture scaling solved fullscreen without rewriting every coordinate
- Enemy and tower archetype tables made adding Runner/Tank/Cannon mostly data changes
- Wave recipes are easy to read and tweak; no code change to reshuffle wave 3
- Projectile modes reused the same pool; ballistic and homing share acquire/update/render paths

## What broke

- **Half-finished tower rollout.** Cannon and Ice landed before projectiles and shop buttons caught up; several commits of "no projectiles, no button" in the log
- **Wave recipe memory leak.** Forgetting to `delete` the old recipe before building a new spawner leaked dynamic arrays every wave start. Fixed same day
- **Path length check inside the enemy loop.** A `return` on short paths aborted the entire update instead of skipping one enemy. Moved the guard outside the loop
- **Range typo.** A bug in `distance` made archer range longer than intended, caught and fixed at the end of the week along with a range buff

## Repo snapshot

June 20–26 added about 650 net lines across 16 files. New modules: `display.odin`, `endscreen.odin`. Deleted: `hud.odin`. Three towers, three enemies, two projectile modes, and a win/lose loop.

Next: tests, more refactors, and cleaning up the control panel as tower count grows.

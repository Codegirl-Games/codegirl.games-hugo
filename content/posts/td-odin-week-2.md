---
title: "Week 2: Projectiles, footprints, and a UI detour"
description: "Second week on the Odin tower defense prototype: homing projectiles, multi-tile towers, render split, and learning to stick with Raylib for UI."
date: 2026-06-19
type: devlog
series: tower-defense-prototype
series_order: 3
languages: ["odin"]
tags: ["odin", "tower-defense", "devlog"]
---

Week two on the [Odin prototype](https://github.com/Codegirl-Games/tower-defense-prototype). Last week ended with instant-hit archers and a keyboard-driven HUD. This week the game started to look like a tower defense: arrows fly, towers occupy real space on the grid, and I got my first proper UI, after a brief and expensive detour through third-party UI libraries.

## What I built

**Projectiles.** Archers no longer subtract health on the frame they fire. Towers spawn homing projectiles from a second object pool (`[MAX_PROJECTILES]Projectile`), track the target enemy by slot index, and deal damage on impact. Same pattern as the enemy pool from week one: acquire slot, reset fields, mark inactive when done.

![Wave 1 combat with homing projectiles](/images/tower-defense/td-week-2-projectiles.png)

**Render split.** `render_world` became a thin orchestrator: `render_map`, `render_enemies`, `render_towers`, `render_projectiles`. Each system owns its draw calls. Small refactor, big readability win as files grew.

**Multi-tile towers.** Tower archetypes gained `footprint_w` and `footprint_h`. The archer is 1×2 tiles, taller than a single cell. Placement now checks whether the full rectangle fits on build tiles and does not overlap other towers. Rendering draws a gold rectangle sized to the footprint instead of a fixed 28×28 square.

**Memory cleanup.** Added explicit `delete` calls for dynamic arrays (`path`, `towers`, `events`) when the game loop exits. Odin will not save you from leaking if you allocated with `append`.

**Overlay bar.** Moved gold and base health out of scattered `DrawText` calls into `overlay.odin`, a top bar with consistent padding and colors. Wave count and combat hints stayed in `hud.odin` for now.

**Control panel.** Bottom bar with clickable buttons: select Archer, start wave. `poll_controls_command` feeds into the existing command layer so UI clicks and keyboard shortcuts share one path. Added `utils.odin` with a reusable `draw_button` helper for centered labels.

## What worked

- Projectile pool mirrored the enemy pool: copy the pattern, ship faster
- Footprint-based placement forced me to think in grid coordinates early; multi-tower-type layouts will need this anyway
- Command layer absorbed UI input cleanly: buttons return `Command` values just like keyboard handlers
- Rip-and-replace on Clay was painful but left me with a simpler codebase than I started with

## What broke

- **UI library detour.** Tried ImGui bindings, then [Clay](https://github.com/nicklockwood/clay) for two days. Gold moved to Clay, start-wave became a Clay button, then I deleted all of it and rewrote the overlay in pure Raylib. Lesson: for a small game HUD, immediate-mode Raylib is enough. Do not import a layout engine until you have a layout problem.
- **Tower selection half-wired.** The Archer button sets `world.selected_tower`, but `execute_command` still hardcodes `.Archer` on placement. UI looks done; logic is not.
- **Split HUD.** Gold and health live in `overlay.odin`, wave/enemy count in `hud.odin`, controls in `controls.odin`. Three files for one screen; next cleanup pass needed.
- **Range bug (later fix).** A typo in the distance function made archer range longer than intended. Caught at the end of the week; fix landed June 26.

## Repo snapshot

June 13–19 added about 470 net lines across 18 files. New modules: `projectile.odin`, `overlay.odin`, `controls.odin`, `utils.odin`. Still one tower type, but the archer now shoots, occupies space, and has a shop button.

Next: end screens, fullscreen, more tower types, and enemy variety.


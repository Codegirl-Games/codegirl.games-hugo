---
title: "Building a browser deckbuilder in vanilla JavaScript"
description: "How a birthday gift became a 6,000-line roguelike deckbuilder: state machines, commands, data tables, and what the commit history looks like when you ship anyway."
date: 2026-07-04
type: lesson
languages: ["javascript"]
tags: ["javascript", "patterns", "deckbuilder", "web"]
---

Last year I built a parody *Slay the Spire* game in the browser as a birthday gift for [ThePrimeagen](https://www.twitch.tv/theprimeagen). It started as a joke. The [repo](https://github.com/codegirl-007/theprimeagen-spire) ended up at ~6,400 lines of JavaScript across 60 files, with two acts, dozens of cards, community-written birthday messages, and a full roguelike loop you can play without a build step.

You can play it at <a href="https://theprimeagenbirthday.com" target="_blank" rel="noopener noreferrer">theprimeagenbirthday.com</a>.

This post is not a feature tour. It is how the project was built, what the git history actually shows, and where *Game Programming Patterns* shows up outside my Odin prototypes.

## What shipped

The game is a static site: `index.html`, ES modules, CSS split by screen. No React, no bundler in production (npm was added briefly for tests, then removed).

The loop matches StS closely enough to feel familiar:

- Branching map with battles, elites, shops, rest sites, and events
- Turn-based combat with energy, block, weak/vulnerable, and intents
- Deck building: strike/defend staples plus dev-themed cards (`Terminal Coffee Rush`, `Production Deploy`, `Code Review`)
- Relics with hook functions (`onTurnStart`, `onDamageTaken`, etc.)
- Two acts with different enemy rosters and map layouts
- Win/lose screens, mid-run saves to `localStorage`, and a pre-launch countdown that blocked play until September 9, 2025

The flavor is extremely online. Enemies are stream/community in-jokes. Events quote Lewis and Tolkien. The victory screen unlocks birthday messages from people who sent notes for Prime. That part is personal. The architecture underneath is reusable.

## What the commits look like

There are 88 commits from August 30, 2025 to March 11, 2026. Rough phases:

**Week one (Aug 30–Sep 2): gameplay exists.** The initial commit already added ~7,600 lines: map, battle engine, card UI, styling. After that it was iteration: card costs, enemy tuning, keyboard vs mouse fixes, double-tap to play cards, swipe sounds, acts, and a commit literally titled `think this is the final gameplay commit`.

**Week two (Sep 3–Sep 10): content and polish.** Birthday messages landed in batches (`Birthday Messages`, `phpeepee!`, `casey!`, `DHH!`, `ken wheeler and AOP`). Bug fixes for deck initialization, event HP/energy leaks, block reset between turns. UI passes on the battle screen and welcome message. Balance commits: `nerf act 2`, `nerf dax`, `Un-nerf act 2w`.

**Week two, structure (Sep 8):** `implement state machine`. Gameplay was already there; this commit extracted map, battle, shop, rest, event, victory, defeat, and relic selection into discrete states. That refactor made the rest of the project maintainable.

**Quiet period, then March 2027 prep.** From September to March the repo sat mostly idle. Then a concentrated refactor week: split client vs shared code, moved data files, serialized shop/reward state for future networking, WebP assets, save behavior fixes (`fix block leak between enemy turn`), and a 1,700-line Cloudflare co-op design doc (`tutorial.md`) for authoritative multiplayer later.

The history is not a clean agile epic. It is burst development, meme commit messages, and a second pass when you already know the game works.

## Architecture that held up

### State machine for screens

`GameStateMachine` registers one class per screen: `MAP`, `BATTLE`, `REWARD`, `SHOP`, `REST`, `EVENT`, `VICTORY`, `DEFEAT`, `RELIC_SELECTION`. Each state implements `enter`, `exit`, `render`, and optional save/restore hooks.

Battle mid-run resume works because `BattleState.getSaveData()` persists the enemy, flags, and `battleInProgress`. On load, bootstrap checks whether you were mid-fight and routes back into combat instead of dropping you on the map with orphaned state.

This is the same *State* pattern I use for Build/Combat phases in the [tower defense prototype](https://github.com/Codegirl-Games/tower-defense-prototype), applied to UI flow instead of simulation phases.

### Command pattern for input

Player actions go through command objects (`PlayCardCommand`, `EndTurnCommand`, `MapMoveCommand`, etc.) executed by a `CommandInvoker`. Keyboard shortcuts, mouse clicks, and shop buttons all funnel into the same paths.

That separation mattered when input got fancy: single number press raises a card, double press plays it. InputManager handles code review picks, shop purchases, and map navigation without duplicating game rules in event listeners.

### Data tables for content

Cards, enemies, relics, and map nodes live in plain JS objects:

```javascript
coffee_rush: {
  id: "coffee_rush",
  name: "Terminal Coffee Rush",
  cost: 0,
  type: "skill",
  effect: (ctx) => { /* ... */ },
}
```

Adding content means adding rows, not subclass trees. Enemy AI is mostly `(turn) => ({ type, value })` functions. Relics use optional hook objects. The `Code Review` card sets `pendingCodeReview` on the root; battle render and InputManager know how to show the pick-one-of-three overlay.

Same idea as `TOWER_ARCHETYPES` and `ENEMY_DEF` in Odin: behavior stays in code, stats and identity stay in tables.

### Shared vs client split (March refactor)

Late refactors moved simulation-ish code under `src/shared/` (`engine/`, `data/`, `game/`) and kept DOM/render/input under `src/client/`. The goal was a future where a Cloudflare Durable Object owns authoritative state while the browser keeps rendering.

Multiplayer never shipped. The split still made the codebase easier to reason about: battle math does not live beside `innerHTML` templates.

## Bugs the commits keep fixing

Roguelikes hide nasty state bugs. This repo is no exception:

- **Block leaking between turns.** Fixed in separate commits for player and enemy turn boundaries. Block must reset at turn start; missing one side means silent damage inflation.
- **Mid-battle saves.** Saving `_battleInProgress`, enemy HP, and hand state to `localStorage`, then restoring on reload. Easy to get wrong when most testing happens in one sitting.
- **Event modifiers vs combat.** `fix health bug and energy bug inside events` shows how one-shot screens can corrupt run state if they touch player stats outside the battle engine's expectations.

The `?screen=battle` and `?screen=shop` URL params plus mock player data in `bootstrap.js` were added so individual screens could be tested without playing to them every time. Worth copying for any UI-heavy browser game.

## Performance passes

Late commits focused on load and layout cost: WebP conversion, dropping aggressive image preload, caching swipe sound, reducing layout churn during battle animations. For a static birthday game, this was optional polish. For a public deploy on slow mobile networks, it matters.

## How this connects to my other work

I keep [*Game Programming Patterns*](https://gameprogrammingpatterns.com/) nearby while building the Odin prototypes. This JavaScript project applies several of the same ideas in a different shape:

| Pattern | Here | Tower defense (Odin) |
|---|---|---|
| State | Screen flow (map, battle, shop) | Build / Combat / Game Over |
| Command | PlayCard, EndTurn, MapMove | Place_Tower, Start_Wave |
| Update method | Per-state `render()` + battle engine steps | `update_enemies`, `update_towers`, etc. |
| Data locality | Plain objects, shuffle/draw on arrays | Fixed pools, archetype tables |

Different language, same instinct: separate input from rules, separate screens from simulation, push content into data.

## What I would do differently

- **State machine earlier.** Gameplay landed first; the refactor on Sep 8 touched 15 files. Starting with states would have hurt day-one momentum but saved mid-project pain.
- **Smaller render files from the start.** `render.js` was enormous before feature folders (`battleRender`, `mapRender`, etc.) split it up.
- **Link the live site in the README.** The game runs at <a href="https://theprimeagenbirthday.com" target="_blank" rel="noopener noreferrer">theprimeagenbirthday.com</a>; the repo should say that up front next to the clone instructions.
- **Move the repo under the studio org.** It still lives at `codegirl-007/theprimeagen-spire`; my prototypes now live under [Codegirl-Games](https://github.com/Codegirl-Games).

## Worth a post?

Yes, but as a lesson, not a weekly devlog. There is no neat week-by-week timeline after launch week. The value is architectural: how far vanilla JS gets you, what patterns transfer to Odin, and an honest commit log that includes `Ligma balls` next to `implement state machine`.

Play at <a href="https://theprimeagenbirthday.com" target="_blank" rel="noopener noreferrer">theprimeagenbirthday.com</a>. To explore the code, start at `src/client/app/bootstrap.js` for the state registration, `src/shared/engine/battle.js` for combat rules, and `src/shared/data/cards.js` for content shape. Multiplayer design notes are in `tutorial.md` if you want to see the planned next step that never left the design doc.

The game was a gift. The structure is the part worth stealing for the next project.

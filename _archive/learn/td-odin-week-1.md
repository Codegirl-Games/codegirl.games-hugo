---
title: "Week 1: Tower defense in Odin — project kickoff"
description: Starting the tower defense prototype in Odin and defining the shared game spec.
date: 2026-07-02
type: devlog
series: td-language-comparison
series_order: 2
languages: ["odin"]
tags: ["odin", "tower-defense", "devlog"]
draft: true
---

First devlog in the tower defense series. This week: project structure, rendering a basic grid, and defining the shared spec that C and C++ ports will follow.

## What worked

- Odin's explicit memory model made the render loop straightforward
- Keeping the game spec language-agnostic from day one

## What broke

- Pathfinding placeholder — towers can be placed anywhere for now
- No entity system yet; everything is in one update function

Next: basic tower placement and enemy waves.

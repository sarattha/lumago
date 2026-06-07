# Platform and Shipping Production Plan

## Goal

Prepare LumaGo projects for reliable desktop shipping with packaging,
configuration profiles, diagnostics, input, audio, localization, asset manifests,
and CI verification.

## Why This Matters Compared With Unity

Unity provides build targets, packaging, player settings, input, audio,
localization workflows, and platform-specific asset handling. LumaGo needs a
smaller but dependable shipping layer for PC-first 2D games.

## Task Checklists

- [ ] Define supported production platforms: Windows, macOS, and Linux desktop.
- [ ] Add build profiles for development, staging, and release.
- [ ] Add packaging commands that include binaries, shaders, assets, configs,
      licenses, and platform runtime notes.
- [ ] Add stable save-data and config paths per platform.
- [ ] Add crash logging and startup diagnostics.
- [ ] Add controller input mapping in addition to keyboard input.
- [ ] Add an audio subsystem for music, sound effects, mixer groups, and volume
      controls.
- [ ] Add localization resource loading for UI text and game strings.
- [ ] Add versioned asset manifests and compatibility checks at startup.
- [ ] Add shader and asset validation as part of release builds.
- [ ] Add CI jobs for tests, asset validation, shader compilation, and smoke
      runs with `NopRenderer`.
- [ ] Add release notes and artifact checksums for packaged builds.

## Exception Criteria

- Do not add mobile or console targets before desktop shipping is stable.
- Do not require Vulkan display availability for headless CI smoke tests.
- Do not ship builds that rely on local development asset paths.
- Do not silently continue when required shaders or manifest entries are missing.
- Do not mix user save data with packaged read-only assets.

## Evaluation

- A release build can be produced for each supported desktop platform.
- Packaged builds start from their own asset manifest and do not rely on the repo
  working directory.
- Missing shader or asset files fail with clear startup diagnostics.
- CI can validate assets and run renderer-independent smoke tests.
- Save data, config, logs, and crash reports are written to documented platform
  locations.

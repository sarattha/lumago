# Scene and Entity Model Production Plan

## Goal

Introduce a stable scene and entity model that separates gameplay entities from
render commands while preserving LumaGo's simple Go-facing API.

## Why This Matters Compared With Unity

Unity production workflows rely on GameObjects, components, prefabs, serialized
scenes, and reusable templates. LumaGo currently focuses on direct scene render
data. Production projects need structured entities that can own sprites,
animation, collision, lights, and game state.

## Task Checklists

- [ ] Add stable entity IDs.
- [ ] Define transform data for position, scale, rotation, z, and optional parent
      relationships.
- [ ] Add components for sprite, animator, light, collider, occluder, and custom
      game tags.
- [ ] Add scene serialization for entities, components, asset references, and
      initial state.
- [ ] Add prefab-like reusable templates for common entity compositions.
- [ ] Add scene load and unload lifecycle hooks.
- [ ] Add a command-generation step that converts components into sprite, light,
      and occluder renderer submissions.
- [ ] Add query helpers for finding entities by ID, tag, or component type.
- [ ] Add validation for missing assets, duplicate IDs, invalid parent links, and
      unsupported component data.

## Exception Criteria

- Do not expose Vulkan or renderer backend details in entity/component data.
- Do not require a complex ECS framework before simple component ownership works.
- Do not remove the existing low-level `scene.AddSprite` and `scene.AddLight`
  style APIs until demos have migrated cleanly.
- Do not store transient runtime-only state in serialized prefab definitions.

## Evaluation

- A scene can be serialized, loaded, updated, and rendered without hand-built
  render command loops.
- Prefab-like templates can create repeated props, lights, and animated
  characters.
- Entity transforms produce the same visual output as direct sprite commands.
- Component validation catches broken asset references before runtime rendering.
- Existing demos can migrate incrementally.

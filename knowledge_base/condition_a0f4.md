# ✅ Condition

> URL: https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/condition
> Exported: 2026-01-22

---

copyCopychevron-down

1. [Plugin Wiki](/xiaomomi-plugins/customcrops/plugin-wiki) chevron-right
2. [🍅 CustomCrops](/xiaomomi-plugins/customcrops)

# ✅ Condition

Plugin provides a powerful condition system. You can use both simple conditions and advanced conditions at the same time. Here are some examples for you to learn conditions.

> simple biome condition only contains the necessary arguments

Copy

```inline-grid min-w-full grid-cols-[auto_1fr] [count-reset:line] print:whitespace-pre-wrap
%!(EXTRA string=✅ Condition, string=https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/condition)biome:
  - minecraft:ocean
  - minecraft:deep_ocean
  - minecraft:cold_ocean
  - minecraft:deep_cold_ocean
  - minecraft:frozen_ocean
  - minecraft:deep_frozen_ocean
  - minecraft:lukewarm_ocean
  - minecraft:deep_lukewarm_ocean
  - minecraft:warm_ocean
```

> advanced condition allows you to use conditions of the same type and add `not-met-actions`
> The section name ("requirement\_permission\_1" in this case) is completely customizable as long as it does not conflict with the condition type name.

Copy

```inline-grid min-w-full grid-cols-[auto_1fr] [count-reset:line] print:whitespace-pre-wrap
%!(EXTRA string=✅ Condition, string=https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/condition)# This example uses two permission conditions
requirement_permission_1:
  type: permission
  value:
    - xxx.1.xxx
requirement_permission_2:
  type: permission
  value:
    - xxx.2.xxx
  not-met-actions:
    action_1:
      type: xxx
      value: ...

# wrong usage (YAML format doesn't allow this)
permission:
  - xxx.1.xxx
permission:
  - xxx.2.xxx
```

## [hashtag](\#condition-library)    Condition Library

> time (Minecraft game time 0~23999)

> ypos (Player's Y coordinate)

> biome (Supports custom biomes)

> world

> weather

> date (Real world date)

> permission

> “>” “>=” “<” “<=” “==” “!=”

> “startsWith” “endsWith” “equals” “contains” “in-list”

> logic (Create complex conditions)

> level (Player exp level)

> random (0~1)

> cooldown

> regex

> environment

> potion-effect

> temperature

> sneak

> plugin-level

> season

> fertilizer

> item-in-hand

> light / natural-light

> gamemode

> crow attack

> water more than (For custom pots)

> water less than (For custom pots)

> point more than

> point less than

> moisture more than (For vanilla farmland)

> moisture less than (For vanilla farmland)

> pot

> fertilizer

> fertilizer type

> region (requires WorldGuard)

![](https://mo-mi.gitbook.io/xiaomomi-plugins/~gitbook/image?url=https%3A%2F%2F3367277500-files.gitbook.io%2F%7E%2Ffiles%2Fv0%2Fb%2Fgitbook-x-prod.appspot.com%2Fo%2Fspaces%252F3tsXOZ7EnqaBWiFptXXV%252Fuploads%252FwVRwa1mjoD4DUzEVDGB1%252Fimage.png%3Falt%3Dmedia%26token%3D4b17a00d-727b-4cb9-9a74-73a20ee8cbaf&width=768&dpr=4&quality=100&sign=ea1a59da&sv=2)

mode 1

![](https://mo-mi.gitbook.io/xiaomomi-plugins/~gitbook/image?url=https%3A%2F%2F3367277500-files.gitbook.io%2F%7E%2Ffiles%2Fv0%2Fb%2Fgitbook-x-prod.appspot.com%2Fo%2Fspaces%252F3tsXOZ7EnqaBWiFptXXV%252Fuploads%252FgYXNKR0VaB14hq5tiGiV%252Fimage.png%3Falt%3Dmedia%26token%3Da4af5c9d-7875-4ae5-bdb8-e3a9d4b41b5f&width=768&dpr=4&quality=100&sign=ceb62182&sv=2)

mode 2

> hand

[Previous💪 Actionchevron-left](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/action) [Next🅿️ Placeholder & Expressionchevron-right](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/placeholder-and-expression)

Last updated 10 months ago
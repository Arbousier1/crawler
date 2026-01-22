# 💦 Sprinkler

> URL: https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/sprinkler
> Exported: 2026-01-22

---

copyCopychevron-down

1. [Plugin Wiki](/xiaomomi-plugins/customcrops/plugin-wiki) chevron-right
2. [🍅 CustomCrops](/xiaomomi-plugins/customcrops) chevron-right
3. [📄 Format](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format)

# 💦 Sprinkler

/CustomCrops/contents/sprinklers/\_\_SPRINKLER\_\_.yml

Let's take `sprinkler_1` as an example to configure the sprinkler settings

**Unique Identifier for Your Sprinkler**:
Start by naming your sprinkler under a unique identifier like `sprinkler_1`. This makes it easy to reference and customize later on.

**Specify the Type of Sprinkler**:
The `type` parameter defines the sprinkler as `FURNITURE`, distinguishing it from block-based types and allowing for more flexibility in placement.

Copy

```inline-grid min-w-full grid-cols-[auto_1fr] [count-reset:line] print:whitespace-pre-wrap
%!(EXTRA string=💦 Sprinkler, string=https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/sprinkler)# The type of sprinkler (BLOCK or FURNITURE)
type: FURNITURE
```

**Choose a Sprinkling Pattern and Set Range**:
Decide how your sprinkler sprinkles! The `working-mode` allows you to set the pattern:

- `1` for a square pattern

- `2` for a rhombus pattern (currently selected)

- `3` for a circular pattern
This adds a creative twist to your watering strategy!


Copy

```inline-grid min-w-full grid-cols-[auto_1fr] [count-reset:line] print:whitespace-pre-wrap
%!(EXTRA string=💦 Sprinkler, string=https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/sprinkler)# Sprinkling pattern mode:
# 1 = square, 2 = rhombus, 3 = circle
working-mode: 2
```

The `range` parameter specifies the reach of the sprinkler. In this example, a range of `1` covers a compact area:

Copy

```inline-grid min-w-full grid-cols-[auto_1fr] [count-reset:line] print:whitespace-pre-wrap
%!(EXTRA string=💦 Sprinkler, string=https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/sprinkler)# Sprinkler's working range:
# □■□
# ■▼■
# □■□
range: 1
```

![](https://mo-mi.gitbook.io/xiaomomi-plugins/~gitbook/image?url=https%3A%2F%2F3367277500-files.gitbook.io%2F%7E%2Ffiles%2Fv0%2Fb%2Fgitbook-x-prod.appspot.com%2Fo%2Fspaces%252F3tsXOZ7EnqaBWiFptXXV%252Fuploads%252FIG1ZvPDhZ4m5BlNwNpKK%252Fimage.png%3Falt%3Dmedia%26token%3D2017acbc-37d6-46b7-a925-c99307f85878&width=768&dpr=4&quality=100&sign=2f482957&sv=2)

mode 1

![](https://mo-mi.gitbook.io/xiaomomi-plugins/~gitbook/image?url=https%3A%2F%2F3367277500-files.gitbook.io%2F%7E%2Ffiles%2Fv0%2Fb%2Fgitbook-x-prod.appspot.com%2Fo%2Fspaces%252F3tsXOZ7EnqaBWiFptXXV%252Fuploads%252Fap66dSjhKnk1vCSvYRtv%252Fimage.png%3Falt%3Dmedia%26token%3D515f554c-7115-4379-b730-2453648b0070&width=768&dpr=4&quality=100&sign=687f7da3&sv=2)

mode 2

![](https://mo-mi.gitbook.io/xiaomomi-plugins/~gitbook/image?url=https%3A%2F%2F3367277500-files.gitbook.io%2F%7E%2Ffiles%2Fv0%2Fb%2Fgitbook-x-prod.appspot.com%2Fo%2Fspaces%252F3tsXOZ7EnqaBWiFptXXV%252Fuploads%252FBfOsqGY2fiQgh5ouMnH0%252Fimage.png%3Falt%3Dmedia%26token%3Db71cc116-66e7-47fb-b85d-7b7cfffa6f85&width=768&dpr=4&quality=100&sign=f3a6745e&sv=2)

mode 3

**Set Water Storage Capacity**:
The `storage` parameter determines how much water your sprinkler can hold— `4` in this case. This limits how long the sprinkler can operate before needing a refill.

**Decide on Unlimited or Limited Water Supply**:
With `infinite` set to `false`, the sprinkler has a finite water supply. If you want a limitless sprinkler, change this to `true`!

**Control Water Usage Per Operation**:
The `water` and `sprinkling` parameters dictate how much water is added to a pot and how much is consumed per sprinkling cycle, both set to `1` here for balanced watering.

**Visual Representation of Your Sprinkler**:
Use `3D-item` and `3D-item-with-water` to define how the sprinkler looks in both dry and watered states. This visual cue enhances the gameplay experience by reflecting the sprinkler’s status.

**Optional 2D Model**:
The `2D-item` is an optional parameter for those who prefer a simpler visual representation or when using 2D views in certain scenarios.

**Whitelist for Planting Pots**:
Ensure your sprinkler only works in specified pots by adding them to `pot-whitelist`. The example allows the default pot type.

**Define Refilling Methods**:
Under `fill-method`, you can get creative with how the sprinkler is refilled. For instance, using a `WATER_BUCKET` returns an empty `BUCKET`, and adds `3` units of water, while using a `POTION` returns a `GLASS_BOTTLE` and adds `1` unit.

**Customize the Water Level Display**:
The `water-bar` section lets you create a unique visual representation of the water level with symbols. Adjust these characters to match your game's style or preferences.

**Set Up Events**:
Under `events`, you can configure how the sprinkler responds to different player actions. Available events: `break`/ `place`/ `interact`/ `work`/ `add_water`/ `full`/ `reach_limit`

**Set Up Requirements**:
Under `requirements`, you can configure the conditions player have to meet before using the sprinkler. Available events: `break`/ `place`/ `use`

[Previous🌽 Cropchevron-left](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/crop) [Next🚰 Watering Canchevron-right](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/watering-can)

Last updated 11 months ago
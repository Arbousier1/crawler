# 🌽 Crop

> URL: https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/crop
> Exported: 2026-01-22

---

copyCopychevron-down

1. [Plugin Wiki](/xiaomomi-plugins/customcrops/plugin-wiki) chevron-right
2. [🍅 CustomCrops](/xiaomomi-plugins/customcrops) chevron-right
3. [📄 Format](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format)

# 🌽 Crop

/CustomCrops/contents/crops/\_\_CROP\_\_.yml

Let's take `tomato` as an example to configure the crop settings

**Unique Identifier for Your Crop**:
Start by naming your crop under a unique identifier like `tomato`. This makes it easy to reference and customize later on.

**Define the Crop Type**:
Set the `type` to either `BLOCK` or `FURNITURE`. For "tomato", it is set to `BLOCK`. This setting affects all custom item types that appear in the entire configuration. But you can also set the type of an item individually through some additional configuration.

Copy

```inline-grid min-w-full grid-cols-[auto_1fr] [count-reset:line] print:whitespace-pre-wrap
%!(EXTRA string=🌽 Crop, string=https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/crop)# Item type
# BLOCK / FURNITURE
type: BLOCK
```

**Set Planting Restrictions**:
Use the `pot-whitelist` to specify which pots are allowed for planting. The "tomato" crop can only be planted in the default pot. If you modify the configuration of the pot, please be sure to modify this configuration simultaneously.

Copy

```inline-grid min-w-full grid-cols-[auto_1fr] [count-reset:line] print:whitespace-pre-wrap
%!(EXTRA string=🌽 Crop, string=https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/crop)# The crop can only be planted on whitelisted pots
pot-whitelist:
  - default
```

**Control Crop's Tick Mode**

Copy

```inline-grid min-w-full grid-cols-[auto_1fr] [count-reset:line] print:whitespace-pre-wrap
%!(EXTRA string=🌽 Crop, string=https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/crop)# Sometimes you may have multiple crops configured,
# and you want the pots to have different tick modes.
# You can change the tick mode to ALL in config.yml
# and then configure it separately here.
ignore-random-tick: false
ignore-scheduled-tick: true
```

**Configure Seed Information**:
The `seed` field identifies the item used to plant the crop. Here, `tomato_seeds` is the seed for the tomato crop.

Copy

```inline-grid min-w-full grid-cols-[auto_1fr] [count-reset:line] print:whitespace-pre-wrap
%!(EXTRA string=🌽 Crop, string=https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/crop)# Seed of the crop
seed: tomato_seeds
```

**Manage Rotation (FURNITURE Mode Only)**:
`random-rotation` controls whether the crop rotates randomly when planted. This only applies if the `type` is `FURNITURE`.

**Set Basic Requirements**:
Define conditions under `requirements` for this crop. For example, the "tomato" can only be planted in Spring or Autumn, and an action bar message is displayed if these conditions are not met.

**Configure Event Settings**:
Customize crop events under `events` for actions like planting or breaking. For instance, when planting a "tomato," a sound ( `minecraft:item.hoe.till`) is played, and a hand-swing animation occurs. Available events for crops: `reach_limit`/ `plant`/ `break`/ `interact`/ `death`

**Define Growth Stages and Models**:
Use the `points` section to outline crop growth stages. For each stage, specify a model (appearance) and actions that occur, such as seed dropping or hologram adjustments.

Available events for crops: `grow`/ `break`/ `interact`

**Customize Growth and Death Conditions**:
Use `grow-conditions` to set conditions for crop growth, such as the season or water level. Similarly, `death-conditions` determine when a crop should die, like during a crow attack or in an unsuitable season.

**Define Custom Bone Meal Effects**:
Under `custom-bone-meal`, configure special effects and actions triggered by using bone meal, such as particles, sounds, or the chance of growth.

[Previous✏️ Textchevron-left](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/text) [Next💦 Sprinklerchevron-right](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/sprinkler)

Last updated 11 months ago
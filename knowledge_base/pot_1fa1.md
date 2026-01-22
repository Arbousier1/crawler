# 🪴 Pot

> URL: https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/pot
> Exported: 2026-01-22

---

copyCopychevron-down

1. [Plugin Wiki](/xiaomomi-plugins/customcrops/plugin-wiki) chevron-right
2. [🍅 CustomCrops](/xiaomomi-plugins/customcrops) chevron-right
3. [📄 Format](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format)

# 🪴 Pot

/CustomCrops/contents/pots/\_\_POT\_\_.yml

Let's take `default` as an example to configure the pot settings

**Unique Identifier for Your Pot**:
Start by naming your pot under a unique identifier like `deault`. This makes it easy to reference and customize later on.

**Set the Maximum Water Storage Capacity**:
Use `max-water-storage` to define how much water your pot can store. For instance, setting it to `5` means the pot can hold up to five units of water. This ensures your plants have enough hydration between waterings.

Copy

```inline-grid min-w-full grid-cols-[auto_1fr] [count-reset:line] print:whitespace-pre-wrap
%!(EXTRA string=🪴 Pot, string=https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/pot)# Maximum capacity for storing water in the pot
storage: 5
```

**Customize the Pot’s Basic Appearance**:
The `base` section allows you to define the visual look of your pot under different conditions:

- `dry`: The model ID used when the pot is dry, showing that it needs water.

- `wet`: The model ID when the pot is hydrated and moist.
These settings add a visual cue to the player, making it easy to see at a glance whether the pot needs watering.


Copy

```inline-grid min-w-full grid-cols-[auto_1fr] [count-reset:line] print:whitespace-pre-wrap
%!(EXTRA string=🪴 Pot, string=https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/pot)# Basic settings for pot appearance and behavior
base:
  # Models for the pot's appearance
  dry: dry_pot  # Model ID when the pot is dry
  wet: wet_pot  # Model ID when the pot is wet
```

**Control Pot’s Interaction with the Environment**:

- `absorb-rainwater`: Set to `true` if you want your pot to automatically absorb water from rain, keeping it moist without manual watering.

- `absorb-nearby-water`: Set to `false` to prevent the pot from absorbing water from nearby water sources. This setting ensures that your pot only gets watered when intended.


Copy

```inline-grid min-w-full grid-cols-[auto_1fr] [count-reset:line] print:whitespace-pre-wrap
%!(EXTRA string=🪴 Pot, string=https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/pot)# Determines if the pot absorbs rainwater when it rains
absorb-rainwater: true
# Determines if water from nearby sources can wet the pot
absorb-nearby-water: false
```

**Control Pot's Tick Mode**

**Manage Fertilizer Application**:
The `max-fertilizers` parameter controls how many units of fertilizer can be applied to a pot at once. Setting it to `1` limits it to a single application, helping prevent over-fertilization which could harm the plants. For the moment it's recommended to keep it `1` as the hologram would show at most one fertilizer at the same time.

**Define Custom Appearances for Fertilized Pots**:
Under `fertilized-pots`, you can set unique models for pots based on different fertilizer effects. Each type (e.g., `quality`, `yield_increase`, `variation`, `soil_retain`, `speed_grow`) can have distinct `dry` and `wet` appearances. This customization adds depth, visually reflecting the benefits of different fertilizers applied to the pot.

**Configure Water Refill Methods**:
The `fill-method` section allows you to define various methods for refilling the pot with water:

- **Method 1**: Using a `WATER_BUCKET` adds `3` units of water. Upon refilling, the player receives an empty `BUCKET`.

- **Method 2**: Using a `POTION` adds `1` unit of water and returns a `GLASS_BOTTLE`.
Both methods include actions like playing a sound ( `minecraft:item.bucket.fill` or `minecraft:item.bottle.fill`) and a hand-swing animation to provide feedback during the refill process.


**Customize the Water Level Display Bar**:
The `water-bar` configuration allows you to create a unique visual indicator for the water level using custom characters. This display provides a quick and visually appealing way to check how much water is left in the can.

**Set Up Events**:
The `events` section is where the real magic happens. Here, you define how the game responds to different interactions with the pots. Available events: `place`/ `break`/ `interact`/ `tick`/ `reach_limitation`/ `add_water`/ `full`/ `max_fertilizers`

**Set Up Requirements**:
Under `requirements`, you can configure the conditions player have to meet before using the pot. Available events: `break`/ `place`/ `use`

[Previous🚰 Watering Canchevron-left](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/watering-can) [Next💩 Fertilizerchevron-right](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/fertilizer)

Last updated 8 months ago
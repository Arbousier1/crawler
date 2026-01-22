# 🚰 Watering Can

> URL: https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/watering-can
> Exported: 2026-01-22

---

copyCopychevron-down

1. [Plugin Wiki](/xiaomomi-plugins/customcrops/plugin-wiki) chevron-right
2. [🍅 CustomCrops](/xiaomomi-plugins/customcrops) chevron-right
3. [📄 Format](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format)

# 🚰 Watering Can

/CustomCrops/contents/watering-cans/\_\_WATERING-CAN\_\_.yml

Let's take `watering_can_1` as an example to configure the watering can settings

**Unique Identifier for Your Watering Can**:
Start by naming your watering-can under a unique identifier like `watering_can_1`. This makes it easy to reference and customize later on.

**Assign a Unique Identifier**:
Begin by setting a unique `item` identifier for your watering can using ItemsAdder or Oraxen. For instance, `watering_can_1` uniquely identifies this particular can, allowing it to be easily referenced throughout your game.

Copy

```inline-grid min-w-full grid-cols-[auto_1fr] [count-reset:line] print:whitespace-pre-wrap
%!(EXTRA string=🚰 Watering Can, string=https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/watering-can)# Unique item identifier for ItemsAdder/Oraxen
item: watering_can_1
```

**Customize the Watering Can's Appearance**:
You have the option to customize how the watering can looks at different water levels. Uncomment the `appearance` section and assign `CustomModelData` values to visually differentiate between an empty, partially filled, or fully filled watering can. This is a great way to add visual feedback for players on the can's status.

Copy

```inline-grid min-w-full grid-cols-[auto_1fr] [count-reset:line] print:whitespace-pre-wrap
%!(EXTRA string=🚰 Watering Can, string=https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/watering-can)# Optional appearance settings for watering cans based on water capacity
# Uncomment and customize CustomModelData if you want different appearances for different water levels
#appearance:
#  0: 1000 # CustomModelData for empty can
#  1: 1001 # CustomModelData for partially filled can
#  2: 1002 # CustomModelData for more filled can
#  3: 1003 # CustomModelData for fully filled can
```

**Set the Water Capacity**:
Define the maximum amount of water your watering can hold with the `capacity` parameter. Here, it is set to `3`, meaning the can can store up to three units of water at a time. This capacity controls how long the can can be used before needing a refill.

Copy

```inline-grid min-w-full grid-cols-[auto_1fr] [count-reset:line] print:whitespace-pre-wrap
%!(EXTRA string=🚰 Watering Can, string=https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/watering-can)# Maximum amount of water the watering can can store
capacity: 3
```

**Determine Water Usage per Use**:
The `water` parameter specifies the amount of water dispensed each time the can is used. Setting this to `1` means that each use of the watering can will consume one unit of water and add `1` unit of water to the pot.

Copy

```inline-grid min-w-full grid-cols-[auto_1fr] [count-reset:line] print:whitespace-pre-wrap
%!(EXTRA string=🚰 Watering Can, string=https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/watering-can)# Amount of water added to pot
water: 1
```

**Define the Effective Watering Range**:
Customize how far the water reaches when used by setting the `effective-range` parameters. `Width` and `length` define a rectangular area around the player where the water will be effective, in this case, a 1x1 area directly in front of the player.

**Decide on Unlimited or Limited Water Supply**:
With `infinite` set to `false`, the watering-can has a finite water supply. If you want a limitless watering-can, change this to `true`!

**Set Refill Methods**:
Define how players can refill the watering can by configuring the `fill-method`. In `method_1`, the player can refill the can by interacting with a `WATER` block (like a water source or well). The `amount` specifies that each interaction adds `1` unit of water to the can. This feature allows you to create interactive water sources throughout your game world.

**Specify Compatible Planting Pots**:
Use the `pot-whitelist` to specify which types of planting pots the watering can will work with. By listing `default`, you ensure that the can is compatible with the basic pot type, but you could add more pot types for variety.

**Allow Interaction with Sprinklers**:
If you want your watering can to be used for filling up sprinklers, list the compatible sprinkler IDs under `sprinkler-whitelist`. In this configuration, the can can fill `sprinkler_1`, `sprinkler_2`, and `sprinkler_3`, extending its functionality.

**Enable Dynamic Descriptions**:
Dynamic lore adds a layer of immersion by displaying real-time information about the can's status. By enabling `dynamic-lore`, the description will change based on the water level. Use placeholders like `{water_bar}` to visually represent the water level, `{current}` for the current water amount, and `{storage}` for maximum capacity. This ensures players always know the state of their watering can at a glance.

**Customize the Water Level Display Bar**:
The `water-bar` configuration allows you to create a unique visual indicator for the water level using custom characters. This display provides a quick and visually appealing way to check how much water is left in the can.

**Set Up Events**:
The `events` section is where the real magic happens. Here, you define how the game responds to different interactions with the watering can. Available events: `full`/ `add_water`/ `no_water`/ `consume_water`/ `wrong_pot`/ `wrong_sprinkler`

**Set Up Requirements**:
Under `requirements`, you can configure the conditions player have to meet before using the watering-can.

[Previous💦 Sprinklerchevron-left](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/sprinkler) [Next🪴 Potchevron-right](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/pot)

Last updated 11 months ago
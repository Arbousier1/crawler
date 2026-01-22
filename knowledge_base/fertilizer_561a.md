# 💩 Fertilizer

> URL: https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/fertilizer
> Exported: 2026-01-22

---

copyCopychevron-down

1. [Plugin Wiki](/xiaomomi-plugins/customcrops/plugin-wiki) chevron-right
2. [🍅 CustomCrops](/xiaomomi-plugins/customcrops) chevron-right
3. [📄 Format](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format)

# 💩 Fertilizer

/CustomCrops/contents/fertilizers/\_\_FERTILIZER\_\_.yml

Let's take `quality_1` as an example to configure the fertilizer settings

**Unique Identifier for the Fertilizer**:
Begin by assigning a unique identifier to each fertilizer type. In this example, the fertilizer is named `quality_1`. This identifier is crucial for distinguishing between different fertilizer types in the game.

**Define the Fertilizer Type and Icon**:

- `type`: Set to `QUALITY`, indicating that this fertilizer improves the quality of the crops it is applied to.

- `icon`: Represents the visual icon that will appear in the game interface when players view or use this fertilizer. The icon '뀆' is a unique symbol that visually represents this specific fertilizer.


Available fertilizer types and their effects:

Copy

```inline-grid min-w-full grid-cols-[auto_1fr] [count-reset:line] print:whitespace-pre-wrap
%!(EXTRA string=💩 Fertilizer, string=https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/fertilizer)QUALITY: Change the weights of star rated crops dropping
SOIL_RETAIN: Reduce the speed of water consumption
SPEED_GROW: Accelerate crop growth
VARIATION: Crops have a higher probability of variation
YIELD_INCREASE: Increase crop yield
```

Copy

```inline-grid min-w-full grid-cols-[auto_1fr] [count-reset:line] print:whitespace-pre-wrap
%!(EXTRA string=💩 Fertilizer, string=https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/fertilizer)type: QUALITY  # The type of fertilizer
```

**Set the Fertilizer’s Effectiveness**:

- `chance`: This is a probability factor (set to `1` here) that influences the effectiveness or success rate of the fertilizer. Adjusting this value can make the fertilizer more or less likely to produce its intended effects.

- `times`: Specifies the duration or number of in-game ticks that the fertilizer remains effective. With a value of `28`, the fertilizer will last for 28 ticking cycles, providing a sustained impact on crop growth.


Copy

```inline-grid min-w-full grid-cols-[auto_1fr] [count-reset:line] print:whitespace-pre-wrap
%!(EXTRA string=💩 Fertilizer, string=https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/fertilizer)chance: 1  # Probability factor influencing effect of fertilizer
```

**Assign a Unique Item ID**:

- `item`: The unique item ID `quality_1` links this configuration to an in-game item, ensuring that when players use this item, the game recognizes it as the `quality_1` fertilizer.


**Specify Application Timing and Pot Compatibility**:

- `before-plant`: Set to `true`, this parameter ensures that the fertilizer must be applied before planting any crops. It enforces strategic planning, requiring players to prepare their soil in advance.

- `pot-whitelist`: Lists all pot types where this fertilizer can be applied. In this example, only `default` pots are allowed, but you can add more pot types to expand compatibility.


**Set Up Events**:
The `events` section is where the real magic happens. Here, you define how the game responds to different interactions with the fertilizers. Available events: `use`/ `before_plant`/ `wrong_pot`

**Set Up Requirements**:
Under `requirements`, you can configure the conditions player have to meet before using the fertilizer.

[Previous🪴 Potchevron-left](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/pot) [Next⚙️ config.ymlchevron-right](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/config.yml)

Last updated 11 months ago
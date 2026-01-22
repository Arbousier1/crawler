# BetonQuest 3.0

> URL: https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/compatibility/betonquest-3.0
> Exported: 2026-01-22

---

copyCopychevron-down

1. [Plugin Wiki](/xiaomomi-plugins/customcrops/plugin-wiki) chevron-right
2. [🍅 CustomCrops](/xiaomomi-plugins/customcrops) chevron-right
3. [🤝 Compatibility](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/compatibility)

# BetonQuest 3.0

https://betonquest.org/RELEASE/

**Requires:** CustomCrops 3.6.48+ & BetonQuest 3.0-DEV+

Thank [Jiminarrow-up-right](https://github.com/mrjimin) for code contribution and examples!

#### [hashtag](\#objective-structures)    **Objective Structures**

Copy

```inline-grid min-w-full grid-cols-[auto_1fr] [count-reset:line] print:whitespace-pre-wrap
%!(EXTRA string=BetonQuest 3.0, string=https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/compatibility/betonquest-3.0)# The default value for amount is 1.
# The default value for targets is it anything is acceptable.

# crop_stage_id, crop_id, pot_id, can_id, and sprinkler_id
# are all defined in the CustomCrops configuration files.
# These values can also be provided as a list (A, B, C, ...)

objective:
  # Crops
  <objective name>: customcrops_harvest_crop <crop_stage_id> [amount:int]
  <objective name>: customcrops_plant_crop <crop_id> [amount:int]

  # Pots
  <objective name>: customcrops_place_pot <pot_id> [amount:int]
  <objective name>: customcrops_break_pot <pot_id> [amount:int]

  # Watering Cans
  <objective name>: customcrops_fill_can <can_id> [amount:int]
  <objective name>: customcrops_water_pot <can_id> [targets:pot_id] [amount:int]
  <objective name>: customcrops_water_sprinkler <can_id> [targets:sprinkler_id] [amount:int]

  # Sprinklers
  <objective name>: customcrops_place_sprinkler <sprinkler_id> [amount:int]
  <objective name>: customcrops_break_sprinkler <sprinkler_id> [amount:int]

  # Fertilizers
  <objective name>: customcrops_use_fertilizer <fertilizer_id> [targets:pot_id] [amount:int]

  # Common
  <objective name>: customcrops_place_scarecrow <scarecrow_id> [amount:int]
  <objective name>: customcrops_break_scarecrow <scarecrow_id> [amount:int]
```

* * *

## [hashtag](\#example-usage)    Example Usage

#### [hashtag](\#crops)    **Crops**

- **Harvest:** Harvest 5 fully grown tomatoes (stage 4).

- **Plant:** Plant 1 tomato seeds.


#### [hashtag](\#pots)    **Pots**

- **Place:** Place 4 pots with the ID `default`.

- **Break:** Break 1 pots with the ID `default`.


#### [hashtag](\#watering-can)    **Watering Can**

- **Fill:** Refill `watering_can_1` from a water source 3 times.

- **Water:** Water `default` pots 5 times using `watering_can_2`.

- **Sprinkler:** Activate or set up `sprinkler_1` using `watering_can_3`.


#### [hashtag](\#sprinklers)    **Sprinklers**

- **Place:** Place 1 sprinkler with the ID `sprinkler_1`.

- **Break:** Break 2 sprinkler with the ID `sprinkler_2`.


* * *

#### [hashtag](\#fertilizers)    **Fertilizers**

- **Use:** Use fertilizer 10 times.


#### [hashtag](\#common)    **Common**

This category covers general interaction objectives with various blocks and items in CustomCrops,

**Scarecrows**

- **Place:** Place 3 scarecrow with the specified ID.

- **Break:** Break 1 scarecrow with the specified ID.


> \[!IMPORTANT\] **Scarecrow ID Check** The `<scarecrow_id>` must match the ID defined in your server's configuration file. You can find these IDs in the following path: `yourServer/plugins/CustomCrops/config.yml`
>
> Specifically, look for the `mechanics.scarecrow.id` section to ensure you are using the correct identifier.

## [hashtag](\#message-configuration)    Message Configuration

**Location:** `yourServer/plugins/BetonQuest/lang/<your_language>.yml`

[PreviousSupported levelerschevron-left](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/compatibility/supported-levelers) [NextBattlePasschevron-right](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/compatibility/battlepass)

Last updated 16 days ago
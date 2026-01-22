# ❓️ Common Questions

> URL: https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/common-questions
> Exported: 2026-01-22

---

copyCopychevron-down

1. [Plugin Wiki](/xiaomomi-plugins/customcrops/plugin-wiki) chevron-right
2. [🍅 CustomCrops](/xiaomomi-plugins/customcrops)

# ❓️ Common Questions

### [hashtag](\#q1-how-to-use-vanilla-farmland-as-pot)    Q1: How to use vanilla farmland as pot

Now you have to make a choose between plugin's watering system(①) and vanilla moisture system(②).

① If you want to use plugin's watering & fertilizer system. It's necessary to disable vanilla moisture in config.yml to prevent conflicts

Copy

```inline-grid min-w-full grid-cols-[auto_1fr] [count-reset:line] print:whitespace-pre-wrap
%!(EXTRA string=❓️ Common Questions, string=https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/common-questions)  # Vanilla farmland settings
  vanilla-farmland:
    # Disable vanilla farmland moisture mechanics
    # This option exists because some users prefer to use the vanilla farmland but the water system conflicts with the vanilla one
    disable-moisture-mechanic: true
```

Then edit the default.yml at /contents/pots folder

Copy

```inline-grid min-w-full grid-cols-[auto_1fr] [count-reset:line] print:whitespace-pre-wrap
%!(EXTRA string=❓️ Common Questions, string=https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/common-questions)default:
  # Maximum water storage capacity
  max-water-storage: 7
  # The most basic settings
  base:
    # Pot models
    dry: minecraft:farmland[moisture=0]
    wet: minecraft:farmland[moisture=7]
  # Does the pot absorb raindrop
  absorb-rainwater: true
  # Does nearby water make the pot wet
  absorb-nearby-water: true
```

② If you prefer the vanilla farmland mechanics, you can just disable all the watering mechanics in CustomCrops.

What you need to do is to edit the default.yml under /contents/pots folder.

Then, you have to make changes on each crop's grow-conditions otherwise they won't grow on vanilla farmlands because `water-more-than` checks the water provided by the plugin while `moisture-more-than` checks the moisture of vanilla farmland block data.

### [hashtag](\#q2-i-cant-use-any-of-the-plugins-functions)    Q2: I can't use any of the plugin's functions

Make sure that your world is not in the blacklist / Make sure that your world is in whitelist

### [hashtag](\#q3-how-to-make-watering-cans-damageable)    Q3: How to make watering-cans damageable

Let's take ItemsAdder as example, what you need to do is to set the material to a damageable item for instance "WOODEN\_SWORD". CustomCrops would handle the durability system for you so you don't need to do anything else!

![](https://mo-mi.gitbook.io/xiaomomi-plugins/~gitbook/image?url=https%3A%2F%2F3367277500-files.gitbook.io%2F%7E%2Ffiles%2Fv0%2Fb%2Fgitbook-x-prod.appspot.com%2Fo%2Fspaces%252F3tsXOZ7EnqaBWiFptXXV%252Fuploads%252F2qDziUgtNFV4QNzEwLl8%252Fimage.png%3Falt%3Dmedia%26token%3D15fa950c-04cf-424b-9aaf-a541f845bbb0&width=768&dpr=4&quality=100&sign=d31d8afa&sv=2)

### [hashtag](\#q4-i-want-to-use-items-from-other-plugins-for-seeds-drops)    Q4: I want to use items from other plugins for seeds/drops

Firstly, add the plugin's name in config.yml. You can get all the compatible plugins on page:

[🤝 Compatibilitychevron-right](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/compatibility)

Then you can use MMOItems anywhere for instance

### [hashtag](\#q5-how-to-apply-levelers-level-to-the-drop-amount-of-crops-how-can-i-get-exp-for-the-leveler-from-ha)    Q5: How to apply leveler's level to the drop amount of crops? / How can I get exp for the leveler from harvesting the crops? / How can I set level requirements for planting a crop?

Firstly check if that plugin is compatible on [Supported levelers](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/compatibility/supported-levelers)

Then register that placeholder into CustomCrops

Now you are able to use the placeholder and expression in the amount of drops

To receive exp from harvesting, you have to add an action in break/interact event section

To use the level requirements for planting, you can follow this example

### [hashtag](\#q6-how-does-protect-original-lore-work)    Q6: How does "protect-original-lore" work?

You can find this option if you check the config.yml carefully

Take ItemsAdder as example

![](https://mo-mi.gitbook.io/xiaomomi-plugins/~gitbook/image?url=https%3A%2F%2F3367277500-files.gitbook.io%2F%7E%2Ffiles%2Fv0%2Fb%2Fgitbook-x-prod.appspot.com%2Fo%2Fspaces%252F3tsXOZ7EnqaBWiFptXXV%252Fuploads%252FSiL5FfMnl4rp39uPec22%252Fimage.png%3Falt%3Dmedia%26token%3Dcd779b8f-e4e9-4e3c-bc66-067d427920a4&width=768&dpr=4&quality=100&sign=1d074ecb&sv=2)

![](https://mo-mi.gitbook.io/xiaomomi-plugins/~gitbook/image?url=https%3A%2F%2F3367277500-files.gitbook.io%2F%7E%2Ffiles%2Fv0%2Fb%2Fgitbook-x-prod.appspot.com%2Fo%2Fspaces%252F3tsXOZ7EnqaBWiFptXXV%252Fuploads%252FwD5y9ufEopV9igjbBeqy%252Fimage.png%3Falt%3Dmedia%26token%3D4bfcb869-dcbc-43c6-abdf-df4b4fe60f63&width=768&dpr=4&quality=100&sign=a0ba3641&sv=2)

![](https://mo-mi.gitbook.io/xiaomomi-plugins/~gitbook/image?url=https%3A%2F%2F3367277500-files.gitbook.io%2F%7E%2Ffiles%2Fv0%2Fb%2Fgitbook-x-prod.appspot.com%2Fo%2Fspaces%252F3tsXOZ7EnqaBWiFptXXV%252Fuploads%252FVlAbFK2jXqdLHF5mRgdl%252Fimage.png%3Falt%3Dmedia%26token%3D07126563-0e01-401f-bc7c-9afa4a7196f0&width=768&dpr=4&quality=100&sign=e671cc04&sv=2)

### [hashtag](\#q7-i-cant-plant-crops)    Q7: I can't plant crops

Situation1: I can hear the sound of planting and seed is consumed

If you are using ItemsAdder and `FURNITURE` mode, open ItemsAdder's config.yml and set this value higher

If you are using ItemsAdder and `BLOCK` mode, open ItemsAdder's config.yml and set

Situation2: Nothing happened and you have already planted a lot of crops

Open CustomCrops' config.yml

### [hashtag](\#q8-how-to-notify-players-if-they-cant-plant-more-crops)    Q8: How to notify players if they can't plant more crops?

In pot/crop/sprinkler's configs, there's an event type called `reach_limit` where you can add custom actions.

### [hashtag](\#q9-how-to-disable-bone-meal-for-crops)    Q9: How to disable bone meal for crops?

Remove the `custom-bone-meal` section from the crop configs for instance

### [hashtag](\#q10-how-to-use-vanilla-crops-on-pots)    Q10: How to use vanilla crops on pots?

Firstly you have to disable vanilla mechanics for a certain crop by adding the block type in config.yml. This would prevent the vanilla crop from being ticked and dropping items.

Then create a new file under /contents/crops/ folder for instance `wheat.yml`

![](https://mo-mi.gitbook.io/xiaomomi-plugins/~gitbook/image?url=https%3A%2F%2F3367277500-files.gitbook.io%2F%7E%2Ffiles%2Fv0%2Fb%2Fgitbook-x-prod.appspot.com%2Fo%2Fspaces%252F3tsXOZ7EnqaBWiFptXXV%252Fuploads%252Fo5xQx1rDrEgGLAITMlFt%252Fimage.png%3Falt%3Dmedia%26token%3Db3792643-1a81-4108-a145-c3313b72b1ca&width=768&dpr=4&quality=100&sign=5a2fb1dd&sv=2)

### [hashtag](\#q11-how-to-use-vanilla-items-as-drops-seeds)    Q11: How to use vanilla items as drops/seeds?

Just use capital letter for instance "APPLE"

### [hashtag](\#q12-how-to-use-other-vanilla-blocks-as-pots)    Q12: How to use other vanilla blocks as pots?

To continue using the plugin's water and fertilizer mechanics, configure it as follows:

To disable the plugin's mechanisms, simply configure it as follows:

### [hashtag](\#q13-how-to-let-different-pots.-crops-have-different-tick-modes)    Q13: How to let different pots./crops have different tick modes

Firstly set the tick mode to `ALL` in config.yml

Then configure the detailed mode for specific pots in their own configs which can be found on

[🪴 Potchevron-right](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/pot) [🌽 Cropchevron-right](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/crop)

[Previous🍅 CustomCropschevron-left](/xiaomomi-plugins/customcrops) [Next📄 Formatchevron-right](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format)

Last updated 11 months ago
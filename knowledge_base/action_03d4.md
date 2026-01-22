# 💪 Action

> URL: https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/action
> Exported: 2026-01-22

---

copyCopychevron-down

1. [Plugin Wiki](/xiaomomi-plugins/customcrops/plugin-wiki) chevron-right
2. [🍅 CustomCrops](/xiaomomi-plugins/customcrops)

# 💪 Action

The action system offers pre-set effects provided by the plugin. However, you can also add your own actions by using the API. This system can be applied anywhere an action can be triggered, like when you break a crop, interact a pot etc.

Action is composed of three parts

Copy

```inline-grid min-w-full grid-cols-[auto_1fr] [count-reset:line] print:whitespace-pre-wrap
%!(EXTRA string=💪 Action, string=https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/action)type: The action type
value: The arguments of the action
chance: Optional (0~1 Default: 1)
```

If the value is an integer or a double value, you can use expression for instance

Copy

```inline-grid min-w-full grid-cols-[auto_1fr] [count-reset:line] print:whitespace-pre-wrap
%!(EXTRA string=💪 Action, string=https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/action)exp_action:
  type: exp
  value: '{level} * 3'
```

Here's an example that combines the conditions and actions. In this example, if you interact the ripe crop with empty hand, it would be harvested. If you interact it with a seed, it would be replanted.

![](https://mo-mi.gitbook.io/xiaomomi-plugins/~gitbook/image?url=https%3A%2F%2Fi.imgur.com%2F7oSqXUt.gif&width=768&dpr=4&quality=100&sign=fc3c8580&sv=2)

## [hashtag](\#action-library)    Action Library

> message (Send a message to player)

> broadcast (Send a message to online players)

> command (Execute a console command)

> player-command (Execute a command as player)

> random-command (Execute a random console command from list)

> close-inv (Close the inventory player is currently opening)

> actionbar

> random-actionbar

> mending (Give player exp that would be applied to mending)

> force-tick

> swing-hand

> exp (Give player exp that would directly go into levels)

> chain (Execute actions as a group)

> delay (Delay x ticks)

> timer

> hologram (Display a hologram for some time)

> fake-item (Display a fake item for some time)

> food

> saturation

> give-item

> item-amount

> durability

> variation (Crops turning into another block)

> quality-crops

> drop-item

> plant

> break

> particle

> give-money

> take-money

> title

> random-title

> sound

> potion-effect

> plugin-exp (Experience from other plugins for instance a job/skill plugin)

> conditional (Actions can only be triggered when player satisfy the conditions)

> priority (Execute the first action group that meets the conditions.)

> level

> spawn-entity

[Previous⚙️ config.ymlchevron-left](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/format/config.yml) [Next✅ Conditionchevron-right](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/condition)

Last updated 1 year ago
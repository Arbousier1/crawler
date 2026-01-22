# 🗡️ Random tick vs Scheduled tick

> URL: https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/random-tick-vs-scheduled-tick
> Exported: 2026-01-22

---

copyCopychevron-down

1. [Plugin Wiki](/xiaomomi-plugins/customcrops/plugin-wiki) chevron-right
2. [🍅 CustomCrops](/xiaomomi-plugins/customcrops)

# 🗡️ Random tick vs Scheduled tick

### [hashtag](\#how-does-random-tick-work)    How does random tick work?

> [Chunksarrow-up-right](https://minecraft.fandom.com/wiki/Chunk) consist of one subchunk per 16 blocks of height, each one being a 16×16×16=4096 block cube. Sections are distributed vertically starting at the lowest y level. Every chunk tick, some blocks are chosen at random from each section in the chunk. The blocks at those positions are given a "random tick".
>
> ——Minecraft wiki

CustomCrops give blocks random tick every second. The default value of random-tick-speed is 20, which means the plugin would choose 20 random blocks every second from a 16x16x16 section.

![](https://mo-mi.gitbook.io/xiaomomi-plugins/~gitbook/image?url=https%3A%2F%2Fi.imgur.com%2FUoxqpEa.gif&width=768&dpr=4&quality=100&sign=8f923a80&sv=2)

![](https://mo-mi.gitbook.io/xiaomomi-plugins/~gitbook/image?url=https%3A%2F%2F3367277500-files.gitbook.io%2F%7E%2Ffiles%2Fv0%2Fb%2Fgitbook-x-prod.appspot.com%2Fo%2Fspaces%252F3tsXOZ7EnqaBWiFptXXV%252Fuploads%252FmvtHBZSMm17WB0ETPsxZ%252Fimage.png%3Falt%3Dmedia%26token%3D884f8679-3a13-4798-95d4-4dc740a0307d&width=768&dpr=4&quality=100&sign=eb4c5adb&sv=2)

The size of one section (F3+G)

### [hashtag](\#how-does-scheduled-tick-work)    How does scheduled tick work?

Scheduled tick means that each tick is planned in advance. In each tick cycle, the plugin can ensure that all blocks are randomly distributed. Although the tick time of blocks are different in one cycle, if they are ticked for many cycles, the number of times they are ticked is similar.

![](https://mo-mi.gitbook.io/xiaomomi-plugins/~gitbook/image?url=https%3A%2F%2F3367277500-files.gitbook.io%2F%7E%2Ffiles%2Fv0%2Fb%2Fgitbook-x-prod.appspot.com%2Fo%2Fspaces%252F3tsXOZ7EnqaBWiFptXXV%252Fuploads%252FBVXPmjvJ6QqHAXmrINZ2%252Fscheduled%2520tick.png%3Falt%3Dmedia%26token%3D4d7c0b73-a902-4822-a17a-eb835bf31e19&width=768&dpr=4&quality=100&sign=b387c90f&sv=2)

![](https://mo-mi.gitbook.io/xiaomomi-plugins/~gitbook/image?url=https%3A%2F%2Fi.imgur.com%2FAelrXIM.gif&width=768&dpr=4&quality=100&sign=a70b2199&sv=2)

[Previous🅿️ Placeholder & Expressionchevron-left](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/placeholder-and-expression) [Next🤝 Compatibilitychevron-right](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/compatibility)

Last updated 1 year ago
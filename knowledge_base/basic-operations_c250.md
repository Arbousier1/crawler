# Basic Operations

> URL: https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/api/basic-operations
> Exported: 2026-01-22

---

copyCopychevron-down

1. [Plugin Wiki](/xiaomomi-plugins/customcrops/plugin-wiki) chevron-right
2. [🍅 CustomCrops](/xiaomomi-plugins/customcrops) chevron-right
3. [⌨️ API](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/api)

# Basic Operations

### [hashtag](\#adapt-bukkit-location)    Adapt Bukkit Location

Copy

```inline-grid min-w-full grid-cols-[auto_1fr] [count-reset:line] print:whitespace-pre-wrap
%!(EXTRA string=Basic Operations, string=https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/api/basic-operations)Pos3 pos3 = Pos3.from(location);
```

### [hashtag](\#get-customcrops-world)    Get CustomCrops world

Copy

```inline-grid min-w-full grid-cols-[auto_1fr] [count-reset:line] print:whitespace-pre-wrap
%!(EXTRA string=Basic Operations, string=https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/api/basic-operations)BukkitCustomCropsPlugin.getInstance().getWorldManager().getWorld(Bukkit.getWorld("world"));
```

### [hashtag](\#get-remove-blockstate)    Get/Remove Blockstate

Copy

```inline-grid min-w-full grid-cols-[auto_1fr] [count-reset:line] print:whitespace-pre-wrap
%!(EXTRA string=Basic Operations, string=https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/api/basic-operations)CustomCropsWorld<?> world = ...;
world.getBlockState(pos3);
world.removeBlockState(pos3);
```

### [hashtag](\#add-blockstate)    Add Blockstate

Copy

```inline-grid min-w-full grid-cols-[auto_1fr] [count-reset:line] print:whitespace-pre-wrap
%!(EXTRA string=Basic Operations, string=https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/api/basic-operations)CropBlock cropBlock = (CropBlock) BuiltInBlockMechanics.CROP.mechanic();
CustomCropsBlockState blockState = cropBlock.createBlockState();
cropBlock.id(blockState, "tomato");
cropBlock.point(blockState, 0);
world.addBlockState(pos3, blockState);
```

### [hashtag](\#set-remove-get-custom-data-in-blockstate)    Set/Remove/Get Custom Data in Blockstate

Copy

```inline-grid min-w-full grid-cols-[auto_1fr] [count-reset:line] print:whitespace-pre-wrap
%!(EXTRA string=Basic Operations, string=https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/api/basic-operations)SynchronizedCompoundMap compoundMap = blockState.compoundMap();
compoundMap.remove("key");
compoundMap.put("key", new StringTag("key", "test"));
compoundMap.get("key");
```

### [hashtag](\#get-the-block-type-from-blockstate)    Get the block type from blockstate

### [hashtag](\#get-built-in-block-item-type-from-registry)    Get built-in block/item type from Registry

### [hashtag](\#place-remove-blocks-on-bukkit-worlds)    Place/Remove blocks on Bukkit Worlds

### [hashtag](\#get-ids)    Get ids

### [hashtag](\#get-configs-of-the-built-in-items)    Get configs of the built-in items

[Previous⌨️ APIchevron-left](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/api) [NextCustom Mechanismchevron-right](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/api/custom-mechanism)

Last updated 1 year ago
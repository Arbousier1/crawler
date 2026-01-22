# Other Block System

> URL: https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/api/other-block-system
> Exported: 2026-01-22

---

copyCopychevron-down

1. [Plugin Wiki](/xiaomomi-plugins/customcrops/plugin-wiki) chevron-right
2. [🍅 CustomCrops](/xiaomomi-plugins/customcrops) chevron-right
3. [⌨️ API](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/api)

# Other Block System

In addition to the oraxen and ItemsAdder compatibility provided by the plugin, you can also adapt customcrops to your own server, especially for some large servers with independent development capabilities.

To adapt your own plugin to customcrops you only need to implement two classes: `AbstractCustomEventListener` & `CustomItemProvider`

Copy

```inline-grid min-w-full grid-cols-[auto_1fr] [count-reset:line] print:whitespace-pre-wrap
%!(EXTRA string=Other Block System, string=https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/api/other-block-system)package net.momirealms.customcrops.api.example;

import net.momirealms.customcrops.api.core.AbstractCustomEventListener;
import net.momirealms.customcrops.api.core.AbstractItemManager;
import org.bukkit.event.EventHandler;

public class MyCustomListener extends AbstractCustomEventListener {

    public MyCustomListener(AbstractItemManager itemManager) {
        super(itemManager);
    }

    @EventHandler(ignoreCancelled = true)
    public void onInteractFurniture(FurnitureInteractEvent event) {
        itemManager.handlePlayerInteractFurniture(...);
    }

    @EventHandler(ignoreCancelled = true)
    public void onInteractCustomBlock(CustomBlockInteractEvent event) {
        itemManager.handlePlayerInteractBlock(...);
    }

    @EventHandler(ignoreCancelled = true)
    public void onBreakFurniture(FurnitureBreakEvent event) {
        itemManager.handlePlayerBreak(...);
    }

    @EventHandler(ignoreCancelled = true)
    public void onBreakCustomBlock(CustomBlockBreakEvent event) {
        itemManager.handlePlayerBreak(..);
    }

    @EventHandler(ignoreCancelled = true)
    public void onPlaceFurniture(FurniturePlaceEvent event) {
        itemManager.handlePlayerPlace(...);
    }

    @EventHandler(ignoreCancelled = true)
    public void onPlaceCustomBlock(CustomBlockPlaceEvent event) {
        itemManager.handlePlayerPlace(...);
    }
}
```

At last, register them on plugin enable

[PreviousPlugin Eventschevron-left](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/api/plugin-events) [NextCustom Season Providerchevron-right](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/api/custom-season-provider)

Last updated 1 year ago
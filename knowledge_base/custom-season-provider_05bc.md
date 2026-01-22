# Custom Season Provider

> URL: https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/api/custom-season-provider
> Exported: 2026-01-22

---

copyCopychevron-down

1. [Plugin Wiki](/xiaomomi-plugins/customcrops/plugin-wiki) chevron-right
2. [🍅 CustomCrops](/xiaomomi-plugins/customcrops) chevron-right
3. [⌨️ API](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/api)

# Custom Season Provider

Create a class that implements `SeasonProvider` and then register it on plugin enable

Copy

```inline-grid min-w-full grid-cols-[auto_1fr] [count-reset:line] print:whitespace-pre-wrap
%!(EXTRA string=Custom Season Provider, string=https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/api/custom-season-provider)package net.momirealms.customcrops.api.example;

import net.momirealms.customcrops.api.core.world.Season;
import net.momirealms.customcrops.api.integration.SeasonProvider;
import org.bukkit.World;
import org.jetbrains.annotations.NotNull;

public class MySeasonProvider implements SeasonProvider {

    @Override
    public @NotNull Season getSeason(@NotNull World world) {
        return ...;
    }

    @Override
    public String identifier() {
        return "MySeasonPlugin";
    }
}
```

[PreviousOther Block Systemchevron-left](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/api/other-block-system)

Last updated 1 year ago
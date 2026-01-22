# Custom Mechanism

> URL: https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/api/custom-mechanism
> Exported: 2026-01-22

---

copyCopychevron-down

1. [Plugin Wiki](/xiaomomi-plugins/customcrops/plugin-wiki) chevron-right
2. [🍅 CustomCrops](/xiaomomi-plugins/customcrops) chevron-right
3. [⌨️ API](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/api)

# Custom Mechanism

This page will use sugarcane as an example to roughly implement a mechanism similar to the vanilla sugarcane.

Firstly create two classes for both item and block

Copy

```inline-grid min-w-full grid-cols-[auto_1fr] [count-reset:line] print:whitespace-pre-wrap
%!(EXTRA string=Custom Mechanism, string=https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/api/custom-mechanism)package net.momirealms.customcrops.api.example;

import net.momirealms.customcrops.api.core.item.AbstractCustomCropsItem;
import net.momirealms.customcrops.common.util.Key;

public class SugarCaneItem extends AbstractCustomCropsItem {

    private static SugarCaneItem instance;

    public SugarCaneItem() {
        super(Key.key("customcrops", "sugarcane_item"));
    }

    public static SugarCaneItem instance() {
        if (instance == null) {
            instance = new SugarCaneItem();
        }
        return instance;
    }
}
```

Then register it in `onEnable()` method

Then we can create some basic logics for instance placing the sugarcane block

Now you should be able to plant the sugarcane (I used a pineapple as a substitute in the gif because I haven't prepared the model for sugarcane yet)

![](https://mo-mi.gitbook.io/xiaomomi-plugins/~gitbook/image?url=https%3A%2F%2Fi.imgur.com%2Ff3eQwD8.gif&width=768&dpr=4&quality=100&sign=bcc43a75&sv=2)

![](https://mo-mi.gitbook.io/xiaomomi-plugins/~gitbook/image?url=https%3A%2F%2F3367277500-files.gitbook.io%2F%7E%2Ffiles%2Fv0%2Fb%2Fgitbook-x-prod.appspot.com%2Fo%2Fspaces%252F3tsXOZ7EnqaBWiFptXXV%252Fuploads%252F7lsytwLmI3BujY3EkYps%252Fimage.png%3Falt%3Dmedia%26token%3D33771e2c-ebe2-4976-9454-b328816fbf0a&width=768&dpr=4&quality=100&sign=b1c00579&sv=2)

insight mode

Then we can configure the grow/break logics for sugarcane

![](https://mo-mi.gitbook.io/xiaomomi-plugins/~gitbook/image?url=https%3A%2F%2Fi.imgur.com%2FYTNxP7R.gif&width=768&dpr=4&quality=100&sign=58c4cf8f&sv=2)

[PreviousBasic Operationschevron-left](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/api/basic-operations) [NextPlugin Eventschevron-right](/xiaomomi-plugins/customcrops/plugin-wiki/customcrops/api/plugin-events)

Last updated 1 year ago
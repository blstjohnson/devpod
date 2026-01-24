---
id: rename-provider
title: Rename a Provider
---

You can rename a provider using the `devpod provider rename` command. This is useful for organizing your providers and giving them more descriptive names.

## CLI

To rename a provider via the command-line interface (CLI), use the following command:

```bash
devpod provider rename [CURRENT_NAME] [NEW_NAME]
```

### Arguments

-   `CURRENT_NAME`: The current name of the provider you want to rename.
-   `NEW_NAME`: The new name for the provider.

### Example

If you have a provider named `my-docker` and you want to rename it to `local-docker`, you would run:

```bash
devpod provider rename my-docker local-docker
```

## GUI

You can also rename a provider from the DevPod desktop application:

1.  Navigate to the **Providers** section.
2.  Select the provider you want to rename.
3.  In the provider's configuration page, you will find an editable text field for the provider name.
4.  Change the name to your desired new name.
5.  Click **Update Options** to save the changes.

After renaming, DevPod will automatically update the provider's configuration.

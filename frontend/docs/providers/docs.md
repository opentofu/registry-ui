# Writing documentation for your provider

In order for your provider to show up in the OpenTofu Registry Search properly, you will need to write some documentation. Tools like [terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs) can help you by auto-generating much of the documentation based on your provider schema.

## Documentation structure

You can place your documentation in the `docs` folder in your repository. Please create the files using the following naming convention:

- `/docs/guides/<guide>.md` for guides.
- `/docs/resources/<resource>.md` for resources. (Note: if your resource is called `yourprovider_yourresource`, you should only include `yourresource` here.)
- `/docs/data-sources/<data-source>.md` for resources. (Note: same as for resources)
- `/docs/functions/<function>.md` for functions.

Additionally, if you would like to support CDKTF, you can create the following documents:

- `/docs/cdktf/[python|typescript|csharp|java|go]/resources/<resource>.md`
- `/docs/cdktf/[python|typescript|csharp|java|go]/data-sources/<data-source>.md`
- `/docs/cdktf/[python|typescript|csharp|java|go]/functions/<function>.md`

You can include the following header (front matter) in your markdown files:

```yaml
---
page_title: Title of the page
subcategory: Subcategory to place the page in on the sidebar (optional)
description: Description of the page
---
```

Once you have written your documentation, you can [proceed to publish your provider](/docs/providers/publishing).

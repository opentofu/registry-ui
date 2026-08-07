# Writing documentation for your provider

In order for your provider to show up in the OpenTofu Registry Search properly, you will need to write some documentation. Tools like [terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs) can help you by auto-generating much of the documentation based on your provider schema.

## Documentation structure

You can place your documentation in the `docs` folder in your repository. Please create the files using the following naming convention:

- `/docs/index.md` for the provider overview page.
- `/docs/guides/<guide>.md` for guides.
- `/docs/resources/<resource>.md` for resources. Name the file after your resource without the provider prefix: if your resource is called `yourprovider_yourresource`, name the file `yourresource.md`. The registry does not strip the provider prefix for you, so a file named `yourprovider_yourresource.md` will show up under that literal name instead.
- `/docs/data-sources/<data-source>.md` for data sources. (Note: same naming convention as for resources.)
- `/docs/functions/<function>.md` for functions.

~> If your repository still has a legacy `website/docs` folder (using the `r`/`d`/`f` subfolder names instead of `resources`/`data-sources`/`functions`), the registry will use `website/docs` instead of `docs` whenever both exist, and will ignore `docs` entirely in that case. If your documentation isn't showing up as expected, check whether an old `website/docs` folder is still present in your repository.

Additionally, if you would like to support CDKTF, you can create the following documents:

- `/docs/cdktf/[python|typescript|csharp|java|go]/index.md` for a language-specific overview page.
- `/docs/cdktf/[python|typescript|csharp|java|go]/guides/<guide>.md` for language-specific guides.
- `/docs/cdktf/[python|typescript|csharp|java|go]/resources/<resource>.md`
- `/docs/cdktf/[python|typescript|csharp|java|go]/data-sources/<data-source>.md`
- `/docs/cdktf/[python|typescript|csharp|java|go]/functions/<function>.md`

### Metadata

You can include the following header (front matter) in your markdown files:

```yaml
---
page_title: Title of the page
subcategory: Subcategory to place the page in on the sidebar (optional)
description: Description of the page
---
```

While you can put any metadata in the header, the following fields are used by the OpenTofu registry UI.

#### page_title

Title of the registry UI webpage (and some meta tags).

#### description

Description, used in html meta tags in the registry UI.

#### subcategory

Subcategory can be used to group resources, which is reflected in the UI sidebar. It adds categories in addition to the default `Resources`, `Datasources`, etc.

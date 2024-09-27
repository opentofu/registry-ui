# Publishing a provider version

Once you have finished [creating your provider](/docs/providers/creating) and [wrote your documentation](/docs/providers/docs), you can proceed to publish your provider. First, create a GitHub repository named `terraform-provider-YOURNAME` and publish your source code in this repository. You can then proceed to [create a GitHub release](https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-repository). 

~> Note, that your release name **must** start with a `v` and must follow [semantic versioning](https://semver.org/).

## Using the scaffolding (recommended)

If you used the [terraform-provider-scaffolding-framework](https://github.com/hashicorp/terraform-provider-scaffolding-framework) as your starting point, it also includes the necessary GitHub Actions and [goreleaser file](https://github.com/hashicorp/terraform-provider-scaffolding-framework/blob/main/.goreleaser.yml) needed to create a release.

You will need to create the secrets called `GPG_PRIVATE_KEY` and `PASSPHRASE` in order to sign your release. This is required for the Terraform Registry and recommended for the OpenTofu Registry.  Please follow the [GitHub documentation on generating a GPG key](https://docs.github.com/en/authentication/managing-commit-signature-verification/generating-a-new-gpg-key). (You do not need to add it to your GitHub account, but you will need to [submit the key to the OpenTofu Registry](/docs/providers/adding) in the next step).

Once this is set up, proceed to create a release named `vYOURVERSION` on GitHub and wait for the GitHub Actions job to complete.

## Manually (for the adventurous)

If you are feeling adventurous, you can also create your release artifacts manually. You will need to produce the following artifacts:

- `terraform-provider-YOURNAME_VERSION_PLATFORM_ARCH.zip` containing your provider binary named `terraform-provider-YOURNAME` or `terraform-provider-YOURNAME.exe` and any supplemental files. (e.g. `terraform-provider-aws_5.68.0_darwin_amd64.zip`)
- `terraform-provider-YOURNAME_SHA256SUMS` containing the SHA256 checksums and filenames for all files in your release, one in each line in the format of `<checksum>  <filename>\n` (e.g. `0501ccb379b74832366860699ca6d5993b164ec44314a054453877d39c384869  terraform-provider-aws_5.68.0_darwin_amd64.zip`)
- `terraform-provider-YOURNAME_SHA256SUMS.sig` containing a detached GPG signature for the SHA256SUMS file (without ASCII-armor). This file is optional for OpenTofu.

~> The versions in the filenames **must not** contain the `v` prefix.

OpenTofu supports the following platform names:

- `darwin` (MacOS)
- `linux`
- `windows`
- `freebsd`
- `openbsd`
- `solaris`

You can use these in conjunction with the following architecture names:

- `amd64`
- `arm64`
- `386`
- `arm`

Once you are done, you can upload your release to GitHub and [submit your provider to the OpenTofu Registry](/docs/providers/adding) and the [Terrafom Registry](https://developer.hashicorp.com/terraform/registry/providers/publishing).

# Adding your provider to the OpenTofu Registry

Once you have [published your provider](/docs/providers/publishing), you are ready to add your provider to the OpenTofu Registry. You only need to perform these steps once, the OpenTofu Registry will automatically discover new versions you publish.

## Adding the provider

To add your provider, please go to the [OpenTofu Registry repository](https://github.com/opentofu/registry/issues/new/choose) and select `Submit new Provider`. In the `Provider Repository` field please enter `YOURNAME/terraform-provider-YOURPROVIDER` and submit the issue. An OpenTofu team member will review your submission and merge it into the Registry. Your provider should be live within an hour after merging. 

## Adding the GPG key

After your provider is merged, you can proceed to add your GPG key. You will need to perform the following steps:

1. Export your *public* GPG key into a text file with ASCII-armor.
2. If your provider is located in an organization, make sure you make [your membership in the organization public](https://docs.github.com/en/account-and-profile/setting-up-and-managing-your-personal-account-on-github/managing-your-membership-in-organizations/publicizing-or-hiding-organization-membership). This is required to validate you have the rights to publish a GPG key.
3. Go to the [OpenTofu Registry repository](https://github.com/opentofu/registry/issues/new/choose) and select `Submit new Provider Signing Key`.
4. In the `Provider Namespace` field enter your username or organization name.
5. In the `Provider GPG Key` field paste your GPG key.

~> The GPG key applies to all providers under your username or organization. Don't submit a GPG key if you have providers that are not signed or are signed with a key you don't have.
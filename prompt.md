# Task: Refactor Tagging CCRs to Config-Driven Terraform

## Context

You are working in the `az-security-wiz-cfg` repository, which manages Wiz CSPM policy-as-code. The repo uses a data-driven Terraform pattern where:

1. CCR definitions live as directories under `policy/cloud-configuration-rules/`, each containing:
   ```
   policy/cloud-configuration-rules/
   └── tags-aws-vpc-ccr-000/
       ├── cloud/
       │   ├── query.rego        # The runtime Rego policy (cloud matcher)
       │   └── remediation.md    # Remediation instructions
       ├── description.md        # Human-readable description of the CCR
       └── metadata.json         # CCR config (name, native types, severity, tags, etc.)
   ```

2. A root `ccrs.tf` dynamically discovers all these directories using `fileset()`, reads each file, assembles them into an object per CCR, and passes them into `modules/wiz-ccr` via `for_each`. Here is the existing `ccrs.tf` for reference:

   ```hcl
   locals {
     ccr_base          = "${path.module}/policy/cloud-configuration-rules"
     ccr_dir_file_paths = fileset(local.ccr_base, "**")

     ccr_file_paths = {
       for p in local.ccr_dir_file_paths : regex("[^\\/]+", p) => p...
       if(!startswith(p, ".") && (lower(p) != "readme.md"))
     }

     all_ccr_objects = {
       for p in keys(local.ccr_file_paths) :
       p => {
         metadata                = one([for value in local.ccr_file_paths[p] : jsondecode(file("${local.ccr_base}/${value}")) if length(regexall("^.*/metadata\\.json$", value)) > 0])
         description             = one(try([for value in local.ccr_file_paths[p] : file("${local.ccr_base}/${value}") if length(regexall("^.*/description\\.md$", value)) > 0], []))
         remediation_instructions = one(try([for value in local.ccr_file_paths[p] : file("${local.ccr_base}/${value}") if length(regexall("^.*/cloud/remediation\\.md$", value)) > 0], []))
         cloud                   = one(try([for value in local.ccr_file_paths[p] : file("${local.ccr_base}/${value}") if length(regexall("^.*/cloud/query\\.rego$", value)) > 0], []))
         iac_matchers = {
           # value = {ccr}/{type}/{...}
           for value in local.ccr_file_paths[p] : value => {
             type          = split("/", "${value}")[1]
             rego          = try(file("${local.ccr_base}/${value}"), null)
             parameters    = try(jsondecode(file("${local.ccr_base}/${split("/", "${value}")[0]}/${split("/", "${value}")[1]}/parameters.json")), null)
             remediation   = try(file("${local.ccr_base}/${split("/", "${value}")[0]}/${split("/", "${value}")[1]}/remediation.md"), null)
           } if length(regexall("^.*\\.rego$", value)) > 0 && length(regexall("^.*/cloud/.*", value)) == 0 # if we've indexed a rego file that's not in the cloud matcher folder
         }
       }
     }

     custom_ccr_objects  = { for k, v in local.all_ccr_objects : k => v if length(try(v.metadata.override_builtin, "")) < 1 }
     builtin_ccr_objects = { for k, v in local.all_ccr_objects : k => v if length(try(v.metadata.override_builtin, "")) > 0 }
   }

   module "cloud_configuration_rules" {
     source   = "./modules/wiz-ccr"
     for_each = local.all_ccr_objects

     name                     = try(each.value.metadata.name, null)
     enabled                  = try(each.value.metadata.enabled, null)
     override_builtin         = try(each.value.metadata.override_builtin, null)
     function_as_control      = try(each.value.metadata.function_as_control, null)
     description              = try(each.value.description, null)
     remediation_instructions = try(each.value.remediation_instructions, null)
     severity                 = try(each.value.metadata.severity, null)
     scope_account_ids        = length(try(each.value.metadata.scope_account_ids, [])) > 0 ? each.value.metadata.scope_account_ids : null
     scope_project_id         = try(each.value.metadata.scope_project_id, null) == "*" ? null : each.value.metadata.scope_project_id, null)
     target_native_types      = length(try(each.value.metadata.target_native_types, [])) > 0 ? each.value.metadata.target_native_types : null

     opa_policy = try(each.value.cloud, null)

     iac_matchers = {
       for k, v in each.value.iac_matchers : k => {
         type                     = lower(v.type)
         rego_code                = v.rego
         parameters               = try(v.parameters, null)
         remediation_instructions = try(v.remediation, null)
       }
     }

     tags = try(each.value.metadata.tags, tolist([]))
   }
   ```

3. The `modules/wiz-ccr` module is a thin wrapper around `wiz_cloud_config_rule` — it accepts all these values and creates the resource.

This pattern is well-designed and works great for bespoke CCRs that have unique Rego logic. But for the 117+ tagging CCRs, which all share nearly identical Rego, maintaining 117 separate directory trees is unnecessary duplication.

## The Problem

There are 117+ directories matching `tags-{cloud}-{resource}-ccr-000/` under `policy/cloud-configuration-rules/`. The Rego at `cloud/query.rego` is nearly identical across all of them — only the input path and key field vary.

The Rego for every AWS tagging CCR (except S3) looks like this:

```rego
package wiz

default result := "fail"

requiredTagKeys := {"app_id", "app_resource_name", "class", "cost_center", "data_sensitivity", "ecosystem", "environment", "network_exposure"}

discoveredTagKeys := {key |
  tag := input.Tags[_]
  key := tag["Key"]
}

missingTags := requiredTagKeys - discoveredTagKeys

result := "pass" {
  count(missingTags) == 0
}

currentConfiguration := sprintf("Resource does not have the following tag keys defined: '%v'", [missingTags])
expectedConfiguration := sprintf("Resource has defined tags for all keys: '%v'", [requiredTagKeys])
```

S3 differs only in `input.bucketTags` instead of `input.Tags`. GCP uses `input.labels` and `tag["key"]` (lowercase).

The metadata.json for each follows this pattern:

```json
{
  "name": "TF - AWS - Elastic IP Address must include required tags.",
  "enabled": true,
  "function_as_control": true,
  "scope_account_ids": [],
  "scope_project_id": "*",
  "severity": "low",
  "target_native_types": ["elasticIP"],
  "tags": [
    {"key": "category", "value": "tags"},
    {"key": "cloud-platform", "value": "aws"}
  ]
}
```

## Your Task

Refactor the tagging CCRs so they are driven by a **config file read directly by Terraform** — no code generation step, no build script, no generated files. Terraform reads the config at plan time, templates the Rego inline, and feeds the results into the same `modules/wiz-ccr` module.

### Step 1: Audit existing tagging CCRs

Read ALL directories matching `tags-*` under `policy/cloud-configuration-rules/`. For each one, extract:

- The directory name (CCR identifier)
- From `metadata.json`: `name`, `target_native_types`, `tags` (specifically the `cloud-platform` value), `severity`, `enabled`, `function_as_control`, `scope_project_id`
- From `cloud/query.rego`: the `input.XXX` path and `tag["XXX"]` key field
- From `description.md`: the description text
- From `cloud/remediation.md`: the remediation text

Determine:
1. How many unique Rego patterns exist (likely 3-4 based on input_path and key_field combinations)
2. Whether `description.md` and `cloud/remediation.md` follow a template pattern or have unique content per CCR
3. Whether any tagging CCRs have IaC matchers (subdirectories other than `cloud/`)
4. Whether any metadata fields vary beyond `name`, `target_native_types`, and `tags`

Output this audit as a structured summary before proceeding.

### Step 2: Create the tagging CCR config file

Create a config file at `policy/tagging-ccrs.yaml` (or `.json` if the team prefers — check if the repo already uses YAML elsewhere) that defines all tagging resource types:

```yaml
required_tag_keys:
  - app_id
  - app_resource_name
  - class
  - cost_center
  - data_sensitivity
  - ecosystem
  - environment
  - network_exposure

defaults:
  enabled: true
  function_as_control: true
  scope_project_id: "*"
  severity: low
  description_template: "Checks that {display_name} resources in {cloud_upper} have all required tags defined."
  remediation_template: "Add the required tags to the {display_name} resource. Required tags: {required_tags}."

resource_types:
  aws:
    ec2:
      native_type: ec2Instance
      display_name: EC2 Instance
      input_path: Tags
      key_field: Key
      category: compute

    s3-bucket:
      native_type: bucket
      display_name: S3 bucket
      input_path: bucketTags
      key_field: Key
      category: storage-data

    # ... populate from audit

  azure:
    virtual-machine:
      native_type: virtualMachine
      display_name: Virtual Machine
      input_path: Tags
      key_field: Key
      category: compute

    # ... populate from audit

  gcp:
    compute-instance:
      native_type: gcpComputeInstance
      display_name: Compute Instance
      input_path: labels
      key_field: key
      category: compute

    # ... populate from audit
```

Populate this with EVERY resource type discovered in the audit. Do not omit any.

If the audit reveals that `description.md` or `cloud/remediation.md` have unique per-CCR content that does not fit a template, add an optional `description` and/or `remediation` field to those entries to override the default template.

### Step 3: Create `tagging-ccrs.tf`

Create a new file `tagging-ccrs.tf` (separate from the existing `ccrs.tf`) that:

1. Reads the YAML/JSON config file using `yamldecode(file(...))` or `jsondecode(file(...))`
2. Flattens the resource type map into a single map keyed by CCR identifier (e.g., `tags-aws-ec2-ccr-000`)
3. Templates the Rego inline for each entry using `templatefile()` or string interpolation
4. Templates the description and remediation text
5. Builds an object per CCR that matches the shape expected by `modules/wiz-ccr`
6. Calls `modules/wiz-ccr` with `for_each` over the config-driven CCR objects

The resulting module call should pass the same attributes as the existing `ccrs.tf` pattern so that the `modules/wiz-ccr` module works unchanged:

```hcl
locals {
  # Read the config
  tagging_config = yamldecode(file("${path.module}/policy/tagging-ccrs.yaml"))

  # Flatten into a single map: "tags-{cloud}-{resource}-ccr-000" => { ... }
  tagging_ccr_objects = merge([
    for cloud, resources in local.tagging_config.resource_types : {
      for resource_key, resource in resources :
      "tags-${cloud}-${resource_key}-ccr-000" => {
        metadata = {
          name                = "TF - ${upper(cloud)} - ${resource.display_name} must include required tags."
          enabled             = try(resource.enabled, local.tagging_config.defaults.enabled)
          function_as_control = try(resource.function_as_control, local.tagging_config.defaults.function_as_control)
          scope_account_ids   = try(resource.scope_account_ids, [])
          scope_project_id    = try(resource.scope_project_id, local.tagging_config.defaults.scope_project_id)
          severity            = try(resource.severity, local.tagging_config.defaults.severity)
          target_native_types = [resource.native_type]
          tags = [
            { key = "category", value = "tags" },
            { key = "cloud-platform", value = cloud },
            { key = "resource-category", value = try(resource.category, "general") },
          ]
        }
        description = try(
          resource.description,
          # Use default template, substituting values
          "Checks that ${resource.display_name} resources in ${upper(cloud)} have all required tags defined."
        )
        remediation_instructions = try(
          resource.remediation,
          "Add the required tags to the ${resource.display_name} resource. Required tags: ${join(", ", local.tagging_config.required_tag_keys)}."
        )
        # Template the Rego inline
        cloud = templatefile("${path.module}/policy/templates/tag-presence.rego.tpl", {
          input_path    = resource.input_path
          key_field     = try(resource.key_field, cloud == "gcp" ? "key" : "Key")
          required_keys = join(", ", [for k in sort(local.tagging_config.required_tag_keys) : "\"${k}\""])
        })
        iac_matchers = {}
      }
    }
  ]...)
}

module "tagging_ccrs" {
  source   = "./modules/wiz-ccr"
  for_each = local.tagging_ccr_objects

  name                     = each.value.metadata.name
  enabled                  = each.value.metadata.enabled
  function_as_control      = each.value.metadata.function_as_control
  description              = each.value.description
  remediation_instructions = each.value.remediation_instructions
  severity                 = each.value.metadata.severity
  scope_account_ids        = length(each.value.metadata.scope_account_ids) > 0 ? each.value.metadata.scope_account_ids : null
  scope_project_id         = each.value.metadata.scope_project_id == "*" ? null : each.value.metadata.scope_project_id
  target_native_types      = each.value.metadata.target_native_types

  opa_policy   = each.value.cloud
  iac_matchers = each.value.iac_matchers

  tags = each.value.metadata.tags
}
```

**IMPORTANT:** The above is a starting example to illustrate the pattern. You MUST verify it against the actual `modules/wiz-ccr` module interface by reading the module's `variables.tf`. Match the attribute names and types exactly.

### Step 4: Create the Rego template

Create `policy/templates/tag-presence.rego.tpl`:

```rego
package wiz

default result := "fail"

requiredTagKeys := {${required_keys}}

discoveredTagKeys := {key |
  tag := input.${input_path}[_]
  key := tag["${key_field}"]
}

missingTags := requiredTagKeys - discoveredTagKeys

result := "pass" {
  count(missingTags) == 0
}

currentConfiguration := sprintf("Resource does not have the following tag keys defined: '%v'", [missingTags])
expectedConfiguration := sprintf("Resource has defined tags for all keys: '%v'", [requiredTagKeys])
```

**CRITICAL:** HCL `templatefile()` uses `${}` for interpolation. Rego also uses `{}` for set comprehensions. You MUST verify that the template renders correctly by comparing the output against the existing `cloud/query.rego` files. If there are conflicts, use an alternative approach:
- Escape Rego braces where needed
- Use `replace()` on a raw string instead of `templatefile()`
- Use `%{if}` / `%{endif}` HCL template directives

Test the template rendering with at least one AWS (Tags/Key), one AWS S3 (bucketTags/Key), and one GCP (labels/key) example to confirm correctness.

### Step 5: Remove the old tagging CCR directories

Once the config-driven `tagging-ccrs.tf` is working and produces identical Terraform plan output, remove the 117+ `tags-*` directories from `policy/cloud-configuration-rules/`.

You will also need to update `ccrs.tf` to exclude the tagging CCRs from its `fileset()` discovery so they are not double-counted. The simplest approach: add a filter to the `ccr_file_paths` local that excludes directories starting with `tags-`:

```hcl
ccr_file_paths = {
  for p in local.ccr_dir_file_paths : regex("[^\\/]+", p) => p...
  if(!startswith(p, ".") && (lower(p) != "readme.md") && !startswith(regex("[^\\/]+", p), "tags-"))
}
```

Verify with `terraform plan` that:
- The config-driven tagging CCRs appear in the plan
- The directory-driven non-tagging CCRs still appear
- There are no duplicates
- The total CCR count matches what existed before

### Step 6: Validate

Run `terraform plan` and verify:
1. The total number of `wiz_cloud_config_rule` resources matches what existed before the refactor
2. No resources are being destroyed and recreated (the CCR names and configurations should be identical)
3. If there ARE destroy/create cycles, investigate whether the module instance address changed (e.g., `module.cloud_configuration_rules["tags-aws-ec2-ccr-000"]` vs `module.tagging_ccrs["tags-aws-ec2-ccr-000"]`) — this will require `terraform state mv` commands to migrate state

**State migration via `moved` blocks (REQUIRED):** Because the tagging CCRs are moving from `module.cloud_configuration_rules` to `module.tagging_ccrs`, Terraform will see them as destroy+create. We do NOT run Terraform locally — all applies happen in CI. Therefore, use `moved` blocks to handle the state migration declaratively in code.

Create a `tagging-ccrs-moved.tf` file containing a `moved` block for every tagging CCR:

```hcl
# Auto-generated moved blocks for tagging CCR state migration.
# Safe to delete after the first successful apply.

moved {
  from = module.cloud_configuration_rules["tags-aws-ec2-ccr-000"].wiz_cloud_config_rule.b-ccr[0]
  to   = module.tagging_ccrs["tags-aws-ec2-ccr-000"].wiz_cloud_config_rule.b-ccr[0]
}

moved {
  from = module.cloud_configuration_rules["tags-aws-s3-bucket-ccr-000"].wiz_cloud_config_rule.b-ccr[0]
  to   = module.tagging_ccrs["tags-aws-s3-bucket-ccr-000"].wiz_cloud_config_rule.b-ccr[0]
}

# ... one moved block per tagging CCR
```

**IMPORTANT:** You must check the actual resource address inside the `modules/wiz-ccr` module. Read `modules/wiz-ccr/main.tf` to find the exact resource name and index (it may be `wiz_cloud_config_rule.b-ccr[0]` or `wiz_cloud_config_rule.this` or something else). The `from` and `to` addresses must match exactly or the moved blocks will fail.

Generate a `moved` block for EVERY tagging CCR key (all 117+). You can generate these programmatically from the same YAML config — iterate the resource types and output one `moved` block per entry. After the first successful `terraform apply` in CI, these `moved` blocks can be removed in a follow-up cleanup PR.

## Important constraints

- The `modules/wiz-ccr` module MUST NOT be modified
- The existing directory-driven pattern in `ccrs.tf` MUST continue to work for non-tagging CCRs (overrides, security group rules, etc.)
- The YAML config file is the single source of truth for tagging CCRs going forward
- Adding a new tagging CCR = add an entry to the YAML config file. No directories, no Rego files, no metadata.json.
- Changing the required tag keys = edit one list in the YAML config. All CCRs update on next `terraform apply`.
- The Rego template must produce output that is functionally identical to the existing hand-written Rego

## Expected outcome

After this work:
- 117+ directories under `policy/cloud-configuration-rules/tags-*` are replaced by one YAML config file + one Rego template
- Adding a new resource type = add 4-5 lines to the YAML config
- Changing required tag keys = edit one list in the YAML config
- The `modules/wiz-ccr` module is unchanged
- Non-tagging CCRs continue to work via the existing directory-based pattern
- `terraform plan` shows zero diff against the current state

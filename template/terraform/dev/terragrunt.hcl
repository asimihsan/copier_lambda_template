include {
  path = find_in_parent_folders("root.hcl")
}

terraform {
  # This might run multiple modules in "run-all"
}

dependency "infra" {
  config_path = "../infrastructure"
}

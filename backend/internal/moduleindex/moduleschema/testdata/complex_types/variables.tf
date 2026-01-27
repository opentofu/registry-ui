# Test variables with complex types (similar to hpcugent/opennebula/vsc module from issue #3442)

variable "simple_string" {
  type        = string
  description = "A simple string variable"
}

variable "inferred_type_from_default" {
  description = "A variable with no explicit type (type inferred from default)"
  default     = "ALL"
}

variable "inferred_number" {
  description = "Number type inferred from default"
  default     = 42
}

variable "inferred_bool" {
  description = "Bool type inferred from default"
  default     = true
}

variable "inferred_list" {
  description = "List with no explicit type (should use dynamic)"
  default     = ["a", "b", "c"]
}

variable "inferred_map" {
  description = "Map with no explicit type (should use dynamic)"
  default     = { key = "value" }
}

variable "simple_number" {
  type        = number
  description = "A simple number variable"
}

variable "list_of_strings" {
  type        = list(string)
  default     = []
  description = "A list of strings"
}

variable "map_of_strings" {
  type        = map(string)
  default     = {}
  description = "A map of strings"
}

variable "object_type" {
  type = object({
    name = string
    size = number
  })
  description = "An object type"
}

variable "map_of_objects" {
  type = map(object({
    size       = number
    filesystem = optional(string)
  }))
  default     = {}
  description = "A map of objects with optional fields (like disks variable)"
}

variable "list_of_objects" {
  type = list(object({
    protocol  = string
    range     = string
    rule_type = string
  }))
  default     = []
  description = "A list of objects (like firewall_rules variable)"
}

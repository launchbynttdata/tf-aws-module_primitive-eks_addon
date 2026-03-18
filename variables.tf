variable "cluster_name" {
  description = "The name of the EKS cluster to which the addon will be attached."
  type        = string
}

variable "addon_name" {
  description = "The name of the addon."
  type        = string
}

variable "addon_version" {
  description = "The version of the addon."
  type        = string
  default     = null
}

variable "configuration_values" {
  description = "A JSON string that contains the configuration values for the addon."
  type        = string
  default     = null
}

variable "resolve_conflicts_on_create" {
  description = "How to resolve parameter value conflicts on addon creation."
  type        = string
  default     = null
}

variable "resolve_conflicts_on_update" {
  description = "How to resolve parameter value conflicts on addon update."
  type        = string
  default     = null
}

variable "pod_identity_association" {
  description = "Whether to associate the addon with a pod identity."
  type = object({
    role_arn        = string
    service_account = string
  })
  default = null
}

variable "preserve" {
  description = "Whether to preserve the addon when the cluster is deleted."
  type        = bool
  default     = false
}

variable "service_account_role_arn" {
  description = "The ARN of the IAM role to bind to the addons service account."
  type        = string
  default     = null
}

variable "tags" {
  description = "A map of tags to assign to the resource."
  type        = map(string)
  default     = {}
}

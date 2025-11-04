locals {
  tags = merge(
    {
      "ManagedBy" = "Terraform"
    },
    var.tags,
  )
}

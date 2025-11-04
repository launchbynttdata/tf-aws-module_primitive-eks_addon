output "arn" {
  description = "The Amazon Resource Name (ARN) of the EKS addon."
  value       = aws_eks_addon.this.arn
}

output "id" {
  description = "The ID of the EKS addon."
  value       = aws_eks_addon.this.id
}

output "created_at" {
  description = "The creation timestamp of the EKS addon."
  value       = aws_eks_addon.this.created_at
}

output "modified_at" {
  description = "The last modified timestamp of the EKS addon."
  value       = aws_eks_addon.this.modified_at
}

output "tags_all" {
  description = "A map of all tags assigned to the EKS addon."
  value       = aws_eks_addon.this.tags_all
}

output "addon_version" {
  description = "The version of the EKS addon."
  value       = can(var.addon_version) ? aws_eks_addon.this.addon_version : null
}

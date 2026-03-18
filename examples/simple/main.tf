############################
# Simple EKS Cluster Example
# Balanced configuration for integration testing
############################

module "resource_names" {
  # checkov:skip=CKV_TF_1: trusted module source
  source  = "terraform.registry.launch.nttdata.com/module_library/resource_name/launch"
  version = "~> 2.0"

  for_each = var.resource_names_map

  logical_product_family  = var.logical_product_family
  logical_product_service = var.logical_product_service
  region                  = var.region
  class_env               = var.class_env
  cloud_resource_type     = each.value.name
  instance_env            = var.instance_env
  maximum_length          = each.value.max_length
  instance_resource       = var.instance_resource
}

# Use Launch primitive module for IAM role
module "eks_service_role" {
  # checkov:skip=CKV_TF_1: trusted module source
  source  = "terraform.registry.launch.nttdata.com/module_primitive/iam_role/aws"
  version = "~> 0.1"

  name               = module.resource_names["iam_role"].minimal
  assume_role_policy = var.assume_role_policy
  tags               = var.tags
}

# Use Launch primitive module for policy attachments
module "eks_cluster_policy" {
  # checkov:skip=CKV_TF_1: trusted module source
  source  = "terraform.registry.launch.nttdata.com/module_primitive/iam_role_policy_attachment/aws"
  version = "~> 0.1"

  role_name  = module.eks_service_role.role_name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
}

module "eks_vpc_resource_controller_policy" {
  # checkov:skip=CKV_TF_1: trusted module source
  source  = "terraform.registry.launch.nttdata.com/module_primitive/iam_role_policy_attachment/aws"
  version = "~> 0.1"

  role_name  = module.eks_service_role.role_name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSVPCResourceController"
}

# Data source for availability zones
data "aws_availability_zones" "available" {
  state = "available"
}

# Use Launch primitive module for VPC
module "vpc" {
  # checkov:skip=CKV_TF_1: trusted module source
  source  = "terraform.registry.launch.nttdata.com/module_primitive/vpc/aws"
  version = "~> 1.0"

  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = merge(var.tags, {
    Name = module.resource_names["vpc"].standard
  })
}

# Use Launch primitive modules for subnets
module "subnet_1" {
  # checkov:skip=CKV_TF_1: trusted module source
  source  = "terraform.registry.launch.nttdata.com/module_primitive/subnet/aws"
  version = "~> 1.0"

  vpc_id                  = module.vpc.vpc_id
  cidr_block              = var.subnet_1_cidr
  availability_zone       = data.aws_availability_zones.available.names[0]
  map_public_ip_on_launch = true

  tags = merge(var.tags, {
    Name = module.resource_names["subnet"].minimal_random_suffix
  })
}

module "subnet_2" {
  # checkov:skip=CKV_TF_1: trusted module source
  source  = "terraform.registry.launch.nttdata.com/module_primitive/subnet/aws"
  version = "~> 1.0"

  vpc_id                  = module.vpc.vpc_id
  cidr_block              = var.subnet_2_cidr
  availability_zone       = data.aws_availability_zones.available.names[1]
  map_public_ip_on_launch = true

  tags = merge(var.tags, {
    Name = module.resource_names["subnet"].minimal_random_suffix
  })
}

# Configure default security group to deny all traffic
resource "aws_default_security_group" "default" {
  vpc_id = module.vpc.vpc_id

  tags = merge(var.tags, {
    Name = "${module.resource_names["vpc"].standard}-default-sg"
  })
}

# Use the primitive EKS cluster module
module "eks_cluster" {
  # checkov:skip=CKV_TF_1: trusted module source
  source  = "terraform.registry.launch.nttdata.com/module_primitive/eks_cluster/aws"
  version = "~> 0.1"

  name               = module.resource_names["eks"].minimal
  role_arn           = module.eks_service_role.role_arn
  kubernetes_version = var.kubernetes_version

  enabled_cluster_log_types = ["api", "audit"]

  vpc_config = {
    subnet_ids              = [module.subnet_1.subnet_id, module.subnet_2.subnet_id]
    endpoint_private_access = true
    endpoint_public_access  = true
  }

  tags = var.tags

  depends_on = [
    module.eks_cluster_policy,
    module.eks_vpc_resource_controller_policy
  ]
}

module "eks_addon" {
  source                      = "../../"
  cluster_name                = module.resource_names["eks"].minimal
  addon_name                  = var.addon_name
  addon_version               = var.addon_version
  configuration_values        = var.configuration_values
  resolve_conflicts_on_create = var.resolve_conflicts_on_create
  resolve_conflicts_on_update = var.resolve_conflicts_on_update
  pod_identity_association    = var.pod_identity_association
  preserve                    = var.preserve
  service_account_role_arn    = var.service_account_role_arn
  tags                        = var.tags
  depends_on = [
    module.eks_cluster
  ]
}

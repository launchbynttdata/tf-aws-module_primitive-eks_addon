logical_product_family  = "launch"
logical_product_service = "eks"
class_env               = "sandbox"
instance_env            = 1
instance_resource       = 1
region                  = "us-east-2"

vpc_cidr      = "10.0.0.0/16"
subnet_1_cidr = "10.0.1.0/24"
subnet_2_cidr = "10.0.2.0/24"

kubernetes_version = "1.34"

tags = {
  Environment = "test"
  ManagedBy   = "Terraform"
  Example     = "simple"
}

assume_role_policy = [
  {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"
    principals = [
      {
        type        = "Service"
        identifiers = ["eks.amazonaws.com"]
      }
    ]
  }
]

addon_name    = "vpc-cni" # able to deploy without a node pool
addon_version = "v1.20.4-eksbuild.1"

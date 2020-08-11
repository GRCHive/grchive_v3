terraform {
    backend "gcs" {
        credentials = "../../../deps/gcloud/grchive-v3-deploy.json"
        bucket = "grchive-v3-tf-state-prod"
        prefix = "terraform/state"
    }
}

provider "google" {
    credentials = file("../../../deps/gcloud/grchive-v3-deploy.json")
    project     = "grchive-v3"
    region      = "us-central1"
    zone        = "us-central1-c"
    version     =  "~> 3.7"
    scopes      = [
        "https://www.googleapis.com/auth/compute",
        "https://www.googleapis.com/auth/cloud-platform",
        "https://www.googleapis.com/auth/ndev.clouddns.readwrite",
        "https://www.googleapis.com/auth/devstorage.full_control",
        "https://www.googleapis.com/auth/userinfo.email",
        "https://www.googleapis.com/auth/cloud-platform",
        "https://www.googleapis.com/auth/sqlservice.admin",
    ]
}

module "public" {
    source = "../modules/public"

    wp_database_user = var.wp_database_user
    wp_database_password = var.wp_database_password
    wp_database_name = var.wp_database_name
    wp_instance_name = var.wp_instance_name
}

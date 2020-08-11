# Terraform

* Ensure you have the JSON key for the deployment service account in `$GRCHIVE/deps/gcloud/grchive-v3-deploy.json`. This role needs the following permissions:
    * `storage.objects.create`
    * `storage.objects.delete`
    * `storage.objects.get`
    * `storage.objects.list`
    * `cloudsql.databases.create`
    * `cloudsql.databases.delete`
    * `cloudsql.databases.get`
    * `cloudsql.instances.create`
    * `cloudsql.instances.delete`
    * `cloudsql.instances.get`
    * `cloudsql.sslCerts.create`
    * `cloudsql.sslCerts.delete`
    * `cloudsql.sslCerts.get`
    * `cloudsql.sslCerts.list`
    * `cloudsql.users.create`
    * `cloudsql.users.delete`
    * `cloudsql.users.list`
    * `cloudsql.users.update`
    * `compute.addresses.create`
    * `compute.addresses.delete`
    * `compute.addresses.get`
    * `compute.addresses.use`
    * `compute.disks.create`
    * `compute.disks.delete`
    * `compute.disks.get`
    * `compute.firewalls.create`
    * `compute.firewalls.delete`
    * `compute.firewalls.get`
    * `compute.firewalls.update`
    * `compute.instances.create`
    * `compute.instances.delete`
    * `compute.instances.get`
    * `compute.instances.setMetadata`
    * `compute.networks.create`
    * `compute.networks.delete`
    * `compute.networks.get`
    * `compute.networks.updatePolicy`
    * `compute.subnetworks.get`
    * `compute.subnetworks.use`
    * `compute.subnetworks.useExternalIp`
    * `compute.zones.get`
* `cd $GRCHIVE/pipeline/terraform/prod`
* `terraform init`
* `terraform apply`
